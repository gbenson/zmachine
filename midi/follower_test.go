package midi

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSetPortName(t *testing.T) {
	p := &port{}

	const whackName = ControlSurfaceName + ":" + ControlSurfaceName + " xox 23:0"
	for wantShortName, fullName := range map[string]string{
		"MicroLab mk3":      "MicroLab mk3:MicroLab mk3 MicroLab mk3 20:0",
		"Midi Through":      "Midi Through:Midi Through Port-0 14:0",
		"MPK mini Play mk3": "MPK mini Play mk3:MPK mini Play mk3 MIDI 1 20:0",
		"TiMidity":          "TiMidity:TiMidity port 0 128:0",
		ControlSurfaceName:  "Control Surface:Control Surface ZMCS 20:0",
		whackName:           whackName, // not rewritten
	} {
		p.setName(fullName)
		assert.Equal(t, p.name, wantShortName)
	}
}
