package surface

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestInitialPairedCCState(t *testing.T) {
	cc := &pairedCC{}
	StartForTest(t, cc)

	time.Sleep(time.Millisecond)
	assert.Equal(t, cc.Load(), 0)
}

func TestIncompletePCCUpdate(t *testing.T) {
	cc := &pairedCC{}
	StartForTest(t, cc)

	cc.receiveMSB(36)

	time.Sleep(time.Millisecond)
	assert.Equal(t, cc.Load(), 0)
}

func TestCompletePCCUpdate(t *testing.T) {
	cc := &pairedCC{}
	StartForTest(t, cc)

	cc.receiveMSB(36)
	cc.receiveLSB(52)

	time.Sleep(time.Millisecond)
	assert.Equal(t, cc.Load(), 0x1234)
}

func TestPartiallyReplacedPCCUpdate(t *testing.T) {
	cc := &pairedCC{}
	StartForTest(t, cc)

	cc.receiveMSB(36)
	cc.receiveLSB(52)
	cc.receiveMSB(104)

	time.Sleep(time.Millisecond)
	assert.Equal(t, cc.Load(), 0x1234)
}

func TestReplacedPCCUpdate(t *testing.T) {
	cc := &pairedCC{}
	StartForTest(t, cc)

	cc.receiveMSB(36)
	cc.receiveLSB(52)
	cc.receiveMSB(104)
	cc.receiveLSB(86)

	time.Sleep(time.Millisecond)
	assert.Equal(t, cc.Load(), 0x3456)
}

func TestPCCUpdateAtomicity(t *testing.T) {
	cc := &pairedCC{}
	StartForTest(t, cc)

	var want [1 << 16]int
	for i := range want {
		want[i] = i & ((1 << 14) - 1)
	}

	rand.Seed(186283)
	rand.Shuffle(len(want), func(i, j int) {
		s := want[i]
		want[i] = want[j]
		want[j] = s
	})

	got := make([]int, 0, len(want))

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Second)
	defer cancel()
	start := make(chan struct{})

	// sender
	wg.Go(func() {
		<-start
		for _, v := range want {
			cc.receiveMSB(uint8(v >> 7))
			cc.receiveLSB(uint8(v & 127))
		}
	})

	// receiver
	wg.Go(func() {
		var lastv int
		if lastv == want[0] {
			lastv++
		}
		close(start)

		for _ = range cap(got) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				v := cc.Load()
				if v == lastv {
					continue
				}
				got = append(got, v)
				lastv = v
				break
			}
		}
	})

	wg.Wait()
	t.Logf("sent %d, received %d", len(want), len(got))

	var wantPos, gotPos int
	for gotPos < len(got) && wantPos < len(want) {
		w := want[wantPos]
		v := got[gotPos]
		if v == w {
			wantPos++
			gotPos++
			continue
		}

		found := false
		for wantPos < len(want) && !found {
			w := want[wantPos]
			found = w == v
			if found {
				break
			}
			t.Logf("%3d:%3d: skipping %02x:%02x",
				wantPos, gotPos-wantPos, w>>7, w&127)
			wantPos++
		}
		if !found {
			t.FailNow()
		}
		gotPos++
		wantPos++
	}

	assert.Check(t, gotPos > wantPos/2)
}
