package machine

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gbenson.net/go/logger"
	"gbenson.net/go/logger/log"
	. "gbenson.net/go/zmachine/core"
	"gbenson.net/go/zmachine/util"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Filename string      `toml:"-"`
	Audio    AudioConfig `toml:"audio"`
}

type AudioConfig struct {
	SampleRate Frequency     `toml:"sample_rate"`
	MaxLatency time.Duration `toml:"max_latency"`
}

// ReadFile reads configuration from the default location.
func (c *Config) Read() error {
	name, err := c.locate()
	switch {
	case err != nil:
		return err
	case name != "":
		return c.ReadFile(name)
	default:
		return c.postInit()
	}
}

// locate returns the default configuration filename.
func (c *Config) locate() (string, error) {
	if c.Filename != "" {
		return c.Filename, nil
	}

	if name := os.Getenv("ZMACHINE_CONFIG_TOML"); name != "" {
		return c.checkExist(name)
	}

	const relpath = "zmachine/config.toml"
	dirname, err := os.UserConfigDir()
	if err != nil {
		c.log().Warn().Err(err).Msg("")
	} else {
		name := filepath.Join(dirname, relpath)
		if name, err := c.checkExist(name); name != "" || err != nil {
			return name, err
		}
	}

	name := filepath.Join("/etc", relpath)
	if name, err := c.checkExist(name); name != "" || err != nil {
		return name, err
	}

	return "", nil
}

func (c *Config) checkExist(name string) (string, error) {
	_, err := os.Stat(name)
	if err == nil {
		return name, nil // exists
	} else if errors.Is(err, fs.ErrNotExist) {
		c.log().Debug().Msg(err.Error())
		return "", nil // does not exist
	} else {
		return "", err // problem
	}
}

func (c *Config) log() *logger.Logger {
	l := log.With().Str("comp", util.ComponentName(c)).Logger()
	return &l
}

// ReadFile reads configuration from a named file.
func (c *Config) ReadFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	defer func() { c.Filename = name }()

	_, err = c.ReadFrom(f)
	return err
}

// ReadFrom implements [io.ReaderFrom].
func (c *Config) ReadFrom(r io.Reader) (int64, error) {
	b, err := io.ReadAll(r)
	if err == nil {
		c.Filename = ""
		err = toml.Unmarshal(b, c)
		if err == nil {
			err = c.postInit()
		}
	}
	return int64(len(b)), err
}

// postInit ensures all defaults are set and all values are consistent.
func (c *Config) postInit() error {
	return c.Audio.postInit()
}

// postInit ensures all defaults are set and all values are consistent.
func (c *AudioConfig) postInit() error {
	if c.SampleRate == 0 {
		c.SampleRate = DefaultSampleRate
	}
	if c.MaxLatency == 0 {
		c.MaxLatency = DefaultMaxLatency
	}

	return nil
}
