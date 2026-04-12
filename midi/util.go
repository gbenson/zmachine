package midi

import (
	"errors"
	"strings"

	"gitlab.com/gomidi/midi/v2/drivers"
)

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
