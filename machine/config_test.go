package machine

import (
	"strings"
	"testing"
	"time"

	. "gbenson.net/go/zmachine/core"
	"github.com/BurntSushi/toml"
	"gotest.tools/v3/assert"
)

func TestNonExistingKeysNotReplaced(t *testing.T) {
	m := New()
	assert.Check(t, m.Config.Audio.SampleRate == DefaultSampleRate)
	assert.Check(t, m.Config.Audio.MaxLatency == DefaultMaxLatency)
	m.Config.Audio.SampleRate++
	m.Config.Audio.MaxLatency++
	assert.Check(t, m.Config.Audio.SampleRate != DefaultSampleRate)
	assert.Check(t, m.Config.Audio.MaxLatency != DefaultMaxLatency)
	assert.NilError(t, toml.Unmarshal([]byte{}, &m.Config))
	assert.Check(t, m.Config.Audio.SampleRate == DefaultSampleRate+1)
	assert.Check(t, m.Config.Audio.MaxLatency == DefaultMaxLatency+1)
}

func TestFilenameNotMarshaled(t *testing.T) {
	const filename = "lksdjfl"
	c := &Config{Filename: filename}
	b, err := toml.Marshal(&c)
	assert.NilError(t, err)
	assert.Check(t, !strings.Contains(string(b), filename))
}

func TestFilenameNotUnmarshaled(t *testing.T) {
	const filename = "lksdjfl"
	c := &Config{Filename: filename}
	assert.NilError(t, toml.Unmarshal([]byte(`Filename = "oops"`), c))
	assert.Equal(t, c.Filename, filename)
}

func TestMaxLatencyUnmarshal(t *testing.T) {
	for _, tc := range []string{
		`"1.4s"`,
		`"1400ms"`,
		`"+0m1.4s"`,
		`1400000000`,
		`1_400_000_000`,
	} {
		var c Config
		in := "audio.max_latency=" + tc
		assert.NilError(t, toml.Unmarshal([]byte(in), &c))
		assert.Equal(t, c.Audio.MaxLatency, 1400*time.Millisecond)
	}
}

func TestSampleRateUnmarshal(t *testing.T) {
	for tc, expect := range map[string]Frequency{
		`"44.1KHz"`:   44100 * Hz,
		`"12_000 hz"`: 12 * KHz,
		`"16 KHZ"`:    16 * KHz,
		`"120bpm"`:    120 * BPM,
		`"12.9 BPM"`:  12.9 * BPM,
		`12`:          12,
		`12.345`:      12.345,
	} {
		var c Config
		in := "audio.sample_rate = " + tc
		assert.NilError(t, toml.Unmarshal([]byte(in), &c))
		assert.Equal(t, c.Audio.SampleRate, expect)
	}
}
