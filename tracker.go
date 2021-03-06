package tracker

import (
	"context"
	"sync"
)

var globalDeferFunc *func()

// Thread safe go routine tracker
type Tracker struct {
	wg        *sync.WaitGroup
	ctx       context.Context
	cancel    *context.CancelFunc
	parent    *Tracker
	deferFunc *func()
}

func SetDefaultDefer(function func()) {
	globalDeferFunc = &function
}

func (t Tracker) SetDefer(function func()) {
	t.deferFunc = &function
}

// Root gets you the initial tracker, similar to combining context.Background and context.WithCancel with a waitgroup
func Root() Tracker {
	ctx, cancel := context.WithCancel(context.Background())
	return Tracker{wg: &sync.WaitGroup{}, ctx: ctx, cancel: &cancel}
}

// NewSubGroup is used to get a new sub cancel group basically
func (t Tracker) NewSubGroup() Tracker {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(t.ctx)
	return Tracker{parent: &t, wg: wg, ctx: ctx, cancel: &cancel}
}

// CancelAndWait cancels a tracker and all routines created from it, waiting till they have fully finished
func (t Tracker) CancelAndWait() {
	if t.cancel != nil {
		(*t.cancel)()
	}
	t.wg.Wait()
}

// CancelAndWait cancels a tracker and all routines created from it, without waiting
func (t Tracker) Cancel() {
	if t.cancel != nil {
		(*t.cancel)()
	}
}

// CancelAndWait cancels a tracker and all routines created from it, waiting till they have fully finished
func (t Tracker) Wait() {
	t.wg.Wait()
}

// Cancel multiple things at a time and wait on all of them
func CancelAndWaitMulti(trackers ...Tracker) {
	wg := sync.WaitGroup{}
	for _, t := range trackers {
		if t.cancel != nil {
			(*t.cancel)()
		}
		currentT := t
		wg.Add(1)
		go func() {
			currentT.wg.Wait()
			wg.Done()
		}()
	}

	wg.Wait()
}

// propergation functions
func (t Tracker) wgAdd() {
	t.wg.Add(1)
	if t.parent != nil {
		t.parent.wgAdd()
	}
}
func (t Tracker) wgDone() {
	t.wg.Done()
	if t.parent != nil {
		t.parent.wgDone()
	}
}

// Go starts a tracked go routine and injects a tracker that needs to be used. At a minimum use a select to listen to its Done() channel
func (t Tracker) Go(function func(tkr Tracker)) { // Always call before go routine creation, also always call defer done
	t.wgAdd()
	go func() {
		if t.deferFunc != nil {
			defer (*t.deferFunc)()
		} else if globalDeferFunc != nil {
			defer (*globalDeferFunc)()
		}
		function(t)
		t.wgDone()
	}()
}

// Run, same as Go but syncronus
func (t Tracker) Run(function func(tkr Tracker)) {
	t.wgAdd()
	if t.deferFunc != nil { // Is this a good choice ? I do not use run a lot anyways....
		defer (*t.deferFunc)()
	} else if globalDeferFunc != nil {
		defer (*globalDeferFunc)()
	}
	function(t)
	t.wgDone()
}

// Done is a channel like context.Done()
func (t Tracker) Done() <-chan struct{} {
	return t.ctx.Done()
}

// Context returns the context alone for usage with external libraries
func (t Tracker) Context() context.Context {
	return t.ctx
}
