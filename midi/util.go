package midi

import (
	"errors"
	"regexp"
	"strings"

	"gitlab.com/gomidi/midi/v2/drivers"
)

var portNameRx = regexp.MustCompile(
	`^(?P<short_name>[^:]+):` +
		`(?P<full_name>.+) ` +
		`(?P<alsa_client_number>\d+):` +
		`(?P<alsa_port_number>\d+)$`,
)

// humanizePortName humanizes gomidi port names.
//
// gomidi (with RtMidi, at least) uses port names like:
//   - "Midi Through:Midi Through Port-0 14:0"
//   - "MPK mini Play mk3:MPK mini Play mk3 MIDI 1 20:0"
//   - "MicroLab mk3:MicroLab mk3 MicroLab mk3 20:0"
//   - "TiMidity:TiMidity port 0 128:0"
//
// The 128:0 part is the ALSA client and port numbers, and AIUI
// they're not fixed, so we strip them to make routing messages
// by port name simpler.
func humanizePortName(name string) string {
	if m := portNameRx.FindStringSubmatch(name); len(m) > 2 {
		dname := m[1]
		pname := m[2]
		if dname == pname || strings.HasPrefix(pname, dname+" ") {
			name = dname
		}
	}

	return name
}

// outByName returns the first MIDI out matching the given name.
func outByName(d drivers.Driver, name string) (drivers.Out, error) {
	outs, err := d.Outs()
	if err != nil {
		return nil, err
	}

	for _, out := range outs {
		if strings.Contains(out.String(), name) {
			return out, nil
		}
	}

	return nil, errors.New("port not found " + name)
}
