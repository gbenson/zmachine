package ui

import (
	"fmt"
	"math"
	"testing"

	. "gbenson.net/go/zmachine/core"
	"gotest.tools/v3/assert"
)

func TestLinearParameterCurve(t *testing.T) {
	p := &Parameter{Min: -1, Max: 3}
	c := newLinearParameterCurve(p, 41)

	t.Logf("curve: C = %v, M = %v, invM = %v", c.C, c.M, c.invM)

	for setting, value := range map[int]float64{
		0:  -1.0,
		10: 0.0,
		20: 1.0,
		40: 3.0,
	} {
		assert.Check(t, NearlyEqual(c.SettingToValue(float64(setting)), value))
		assert.Equal(t, int(math.Round(c.ValueToSetting(value))), setting)
	}
}

func TestExponentialParameterCurve(t *testing.T) {
	p := &Parameter{Min: 0.5, Max: 30000}
	c := newExponentialParameterCurve(p, 512)

	t.Logf("curve: C = %v, M = %v, invM = %v", c.C, c.M, c.invM)

	for setting, value := range map[int]string{
		0:   "0.50",
		20:  "1.00",
		46:  "2.03",
		511: "30000.00",
	} {
		assert.Equal(t, fmt.Sprintf("%.2f", c.SettingToValue(float64(setting))), value)
	}

	for setting, value := range map[int]float64{
		0:   0.5,
		85:  5,
		189: 45,
		297: 400,
		405: 3500,
		511: 30000,
	} {
		assert.Equal(t, int(math.Round(c.ValueToSetting(value))), setting)
	}
}

func TestEnvelopePhaseDurationParameterCurve(t *testing.T) {
	p := &Parameter{Max: 30000}
	c := newExponentialParameterCurve(p, 512)

	assert.Equal(t, c.SettingToValue(0), 0.0)
	assert.Equal(t, c.ValueToSetting(0), 0.0)

	assert.Check(t, NearlyEqual(c.SettingToValue(511), 30000))
	assert.Check(t, NearlyEqual(c.ValueToSetting(30000), 511))
}
