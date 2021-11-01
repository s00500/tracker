package tracker

// cancel in easier ways
type TrackerGroup struct {
	trackers []Tracker
}

func (t *TrackerGroup) Add(tkr Tracker) {
	if t.trackers == nil {
		t.trackers = make([]Tracker, 0)
	}
	t.trackers = append(t.trackers, tkr)
}

func (t *TrackerGroup) CancelAndWait() {
	CancelAndWaitMulti(t.trackers...)
}
