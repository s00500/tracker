package tracker_test

import (
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
	case <-time.Tick(time.Second * 5):
	case <-trk.Done():
		return
	}

	time.Sleep(time.Second * 5)
}
