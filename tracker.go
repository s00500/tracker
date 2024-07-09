package tracker

import (
	"context"
	"fmt"
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
	Logging   bool
	Reference string // Another debug field....
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

func FromContext(ctx context.Context) Tracker {
	internalCtx, cancel := context.WithCancel(ctx)
	return Tracker{wg: &sync.WaitGroup{}, ctx: internalCtx, cancel: &cancel}
}

func RootLogging() Tracker {
	ctx, cancel := context.WithCancel(context.Background())
	return Tracker{wg: &sync.WaitGroup{}, ctx: ctx, cancel: &cancel, Logging: true}
}

func (t Tracker) IsRoot() bool {
	return t.parent == nil
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

func (t Tracker) Go(function func(tkr Tracker)) { // Always call before go routine creation, also always call defer done
	if t.ctx == nil {
		fmt.Print("ERROR: Called go on empty tracker, not running")
		return
	}
	t.GoRef("", function)
}

// Run, same as Go but syncronus
func (t Tracker) Run(function func(tkr Tracker)) {
	if t.ctx == nil {
		fmt.Print("ERROR: Called run on empty tracker, not running")
		return
	}
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
	if t.ctx == nil {
		fmt.Print("ERROR: Called done on empty tracker, return instant cancel")
		stop := make(chan struct{})
		go func() {
			stop <- struct{}{}
		}()
		return stop
	}
	return t.ctx.Done()
}

// Context returns the context alone for usage with external libraries
func (t Tracker) Context() context.Context {
	return t.ctx
}
