package tracker_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/s00500/tracker"
)

func TestTracker(t *testing.T) {
	trk := tracker.Root()

	trk.Go(func(tkr tracker.Tracker) {
		someFunc(t, trk)
	})
	trk.Go(func(tkr tracker.Tracker) {
		someFunc(t, trk)
	})
	trk.Go(func(tkr tracker.Tracker) {
		someFunc(t, trk)
	})

	t.Log("Gonna wait for first")
	trk.CancelAndWait()
	t.Log("Done")
	trk.CancelAndWait()
	t.Log("Done2")
}

func TestEmptyTracker(t *testing.T) {
	trk := tracker.Tracker{}

	// Done() on an empty tracker must behave like a cancelled context: a closed
	// channel that stays readable, so polling it repeatedly never blocks.
	for i := 0; i < 3; i++ {
		select {
		case <-trk.Done():
		case <-time.After(time.Second):
			t.Fatalf("Done() blocked on read %d for an empty tracker", i)
		}
	}

	// Go/Run on an empty tracker must be no-ops rather than panic on the nil wg.
	trk.Go(func(tkr tracker.Tracker) { t.Error("Go ran on empty tracker") })
	trk.GoRef("ref", func(tkr tracker.Tracker) { t.Error("GoRef ran on empty tracker") })
	trk.Run(func(tkr tracker.Tracker) { t.Error("Run ran on empty tracker") })
}

func TestSetDefer(t *testing.T) {
	trk := tracker.Root().SetDefer(func() {}) // ensure SetDefer returns a usable tracker

	var deferCount atomic.Int32
	trk = trk.SetDefer(func() { deferCount.Add(1) })

	started := make(chan struct{})
	trk.Go(func(tkr tracker.Tracker) { close(started) })
	<-started
	trk.CancelAndWait()

	if got := deferCount.Load(); got != 1 {
		t.Fatalf("defer func ran %d times, want 1 (SetDefer return value must be used)", got)
	}
}

func someFunc(t *testing.T, trk tracker.Tracker) {
	subTrk := trk.NewSubGroup()

	// readloop
	subTrk.Go(func(tkr tracker.Tracker) {
		for {
			select {
			case <-tkr.Done():
				t.Log("read done")
				return
			}
		}
	})

	// Writeloop
	subTrk.Go(func(tkr tracker.Tracker) {
		for {
			select {
			case <-tkr.Done():
				t.Log("write done")
				return
			}
		}
	})

	select {
	case <-time.After(time.Second * 5):
	case <-trk.Done():
		return
	}

	time.Sleep(time.Second * 5)
}

func TestSubTrackers(t *testing.T) {
	some := struct {
		trk tracker.Tracker
	}{}

	trk := tracker.Root()

	sub := trk.NewSubGroup()

	if !trk.IsRoot() {
		t.Fail()
	}

	if sub.IsRoot() {
		t.Fail()
	}

	if !some.trk.IsRoot() {
		t.Fail()
	}
}
