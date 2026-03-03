package midi

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gbenson.net/go/logger"
	"gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/util"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

const DefaultStartupTimeout = 5 * time.Second
const DefaultRecheckInterval = 1 * time.Second

type Follower struct {
	Driver          drivers.Driver
	Receiver        core.MIDISink
	StartupTimeout  time.Duration
	RecheckInterval time.Duration

	wg   sync.WaitGroup
	stop context.CancelFunc

	ports  map[string]*port
	pollID atomic.Uintptr
}

func (f *Follower) Start(ctx context.Context) error {
	if f.Driver == nil {
		panic("nil driver")
	} else if f.Receiver == nil {
		panic("nil receiver")
	}

	if f.StartupTimeout < 1 {
		f.StartupTimeout = DefaultStartupTimeout
	}
	if f.RecheckInterval < 1 {
		f.RecheckInterval = DefaultRecheckInterval
	}

	log := util.Logger(ctx, f).With().
		Stringer("driver", f.Driver).
		Logger()
	ctx = log.WithContext(ctx)

	ctx, f.stop = context.WithCancel(ctx)
	errC := make(chan error)

	log.Debug().Msg("Starting")
	f.wg.Go(func() {
		defer func() { log.Debug().Msg("Follower stopped") }()
		f.start(ctx, errC)
	})

	launchCtx, cancel := context.WithTimeout(ctx, f.StartupTimeout)
	defer cancel()

	select {
	case err := <-errC:
		return err
	case <-launchCtx.Done():
		return launchCtx.Err()
	}
}

func (f *Follower) Stop(ctx context.Context) {
	log := logger.Ctx(ctx)

	stop := f.stop
	if stop == nil {
		log.Error().Msg("Not running")
		return
	}

	log.Debug().Msg("Stopping")
	defer func() { log.Debug().Msg("Stopped") }()
	stop()
	f.wg.Wait()
}

func (f *Follower) start(ctx context.Context, errC chan error) {
	log := logger.Ctx(ctx)

	// Delegate the error for f.Start to return if the first poll fails.
	if err := f.poll(ctx); err != nil {
		select {
		case errC <- err: // message sent (and received)
		default: // is this even possible??
			log.Err(err).Msg("First poll failed")
		}
		return
	}

	// Still here?  Close errC to release f.Start to return nil.
	close(errC)

	// Poll until cancelled.
	ticker := time.NewTicker(f.RecheckInterval)
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			if err := ctx.Err(); !errors.Is(err, context.Canceled) {
				log.Err(err).Msg("")
			}
			return
		}

		if err := f.poll(ctx); err != nil {
			log.Warn().Err(err).Msg("Poll failed")
		}
	}
}

func (f *Follower) poll(ctx context.Context) error {
	ins, err := f.Driver.Ins()
	if err != nil {
		return err
	}

	thisPoll := f.pollID.Add(1)

	for _, in := range ins {
		name := in.String()
		if p, ok := f.ports[name]; ok {
			p.lastSeen = thisPoll
			continue
		}

		p := &port{follower: f, port: in, lastSeen: thisPoll}
		if err := p.start(ctx); err != nil {
			logger.Ctx(p.ctx).Warn().
				Err(err).
				Msg("Error opening MIDI port")
			continue
		}

		if f.ports == nil {
			f.ports = make(map[string]*port)
		}
		f.ports[name] = p
	}

	for name, p := range f.ports {
		if p.lastSeen == thisPoll {
			continue
		}
		logger.Ctx(p.ctx).Debug().Msg("Closing")
		delete(f.ports, name)
		p.stop()
	}

	return nil
}

type port struct {
	ctx      context.Context
	follower *Follower
	port     drivers.In
	name     string
	lastSeen uintptr
	stop     func()
}

var portNameRx = regexp.MustCompile(
	`^(?P<short_name>[^:]+):` +
		`(?P<full_name>.+) ` +
		`(?P<alsa_client_number>\d+):` +
		`(?P<alsa_port_number>\d+)$`,
)

func (p *port) setName(name string) {
	// RtMidi port names be like:
	//  - "Midi Through:Midi Through Port-0 14:0"
	//  - "MPK mini Play mk3:MPK mini Play mk3 MIDI 1 20:0"
	//  - "MicroLab mk3:MicroLab mk3 MicroLab mk3 20:0"
	//  - "TiMidity:TiMidity port 0 128:0"
	// The 128:0 are the ALSA client and port numbers, and
	// they're not fixed, so we strip them to make routing
	// messages by port name simpler.
	if m := portNameRx.FindStringSubmatch(name); len(m) > 2 {
		dname := m[1]
		pname := m[2]
		if dname == pname || strings.HasPrefix(pname, dname+" ") {
			name = dname
		}
	}

	p.name = name
}

func (p *port) start(ctx context.Context) error {
	p.ctx = ctx
	ctx = nil // crowbar

	in := p.port
	p.setName(in.String())

	log := logger.Ctx(p.ctx).With().
		Int("midi_port", in.Number()).
		Str("source", p.name).
		Logger()
	p.ctx = log.WithContext(p.ctx)

	log.Debug().Msg("Opening")
	if err := in.Open(); err != nil {
		return err
	}

	p.ctx, p.stop = context.WithCancel(p.ctx)
	p.follower.wg.Go(func() {
		defer func() { log.Info().Msg("Closed") }()
		defer util.LoggedClose(p.ctx, in)
		defer p.stop()
		if err := p.listen(); err != nil {
			log.Err(err).Msg("")
		}
	})

	return nil
}

func (p *port) listen() error {
	stop, err := midi.ListenTo(
		p.port,
		p.receiveMessage,
		midi.HandleError(p.receiveError),
	)
	if err != nil {
		return err
	}
	defer stop()

	ctx := p.ctx
	log := logger.Ctx(ctx)
	log.Info().Msg("Opened")

	// I tried calling midi.ListenTo from the follower, but it didn't
	// receive messages and then segfaulted when the process exited,
	// so this does have to be in its own goroutine, with rtmididrv at
	// least, unless i did something stupid like running with the wrong
	// log level, which isn't 100% out of the question.  Anyway, *this*
	// setup works...

	<-ctx.Done()

	log.Debug().Str("reason", ctx.Err().Error()).Msg("Listener stopping")
	return nil
}

func (p *port) receiveMessage(msg midi.Message, timestampms int32) {
	p.follower.Receiver.Receive(msg)
}

func (p *port) receiveError(err error) {
	logger.Ctx(p.ctx).Err(err).Msg("")
}
