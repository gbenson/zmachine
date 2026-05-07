package ui

import (
	"math"

	. "gbenson.net/go/zmachine/core"
)

type parameter struct {
	Name  string
	Value *Parameter
	Curve parameterCurve
}

type parameterCurve interface {
	// SettingToValue calculates a parameter value from its linear equivalent.
	SettingToValue(float64) float64

	// ValueToSetting converts a parameter value to its linear equivalent.
	ValueToSetting(float64) float64
}

func (p *parameter) Update(delta float64) {
	if delta == 0 {
		return
	}

	pv := p.Value
	c := p.Curve

	// This isn't atomic, but only one goroutine calls pv.Store() so its ok.
	v0 := pv.Load()
	s0 := c.ValueToSetting(v0)
	s1 := s0 + delta
	v1 := c.SettingToValue(s1)
	pv.Store(v1)
}

type linearParameterCurve struct {
	M, C, invM float64
}

func newLinearParameterCurve(v *Parameter, n int) *linearParameterCurve {
	m := (v.Max - v.Min) / float64(n-1)
	return &linearParameterCurve{
		M:    m,
		C:    v.Min,
		invM: 1 / m,
	}
}

// y = mx + c
func (curve *linearParameterCurve) SettingToValue(x float64) float64 {
	return curve.M*x + curve.C
}

// x = (y - c)/m
func (curve *linearParameterCurve) ValueToSetting(y float64) float64 {
	return (y - curve.C) * curve.invM
}

type exponentialParameterCurve struct {
	M, C, invM float64
}

func newExponentialParameterCurve(v *Parameter, n int) *exponentialParameterCurve {
	// y = e^(mx) + c		—①
	// ∴ c = y - e^(mx)		—②
	// sub x=0, y=v.Min in ②:
	c := v.Min - 1 //		—③

	// ①: y = e^(mx) + c
	// ∴ m = ln(y - c)/x	—④
	// sub x=n-1, y=v.Max, ③ in ④:
	m := math.Log(v.Max-c) / float64(n-1)

	return &exponentialParameterCurve{
		M:    m,
		C:    c,
		invM: 1 / m,
	}
}

// y = e^(mx) + c
func (curve *exponentialParameterCurve) SettingToValue(x float64) float64 {
	return math.Exp(curve.M*x) + curve.C
}

// x = ln(y - c)/m
func (curve *exponentialParameterCurve) ValueToSetting(y float64) float64 {
	return math.Log(y-curve.C) * curve.invM
}
