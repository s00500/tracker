//go:build !trackdebug

package tracker

import "fmt"

func getPackage() string {
	return ""
}

// Go starts a tracked go routine and injects a tracker that needs to be used. At a minimum use a select to listen to its Done() channel
func (t Tracker) GoRef(ref string, function func(tkr Tracker)) { // Always call before go routine creation, also always call defer done
	if t.ctx == nil || t.wg == nil {
		fmt.Println("ERROR: Called go on empty tracker, not running")
		return
	}
	t.wgAdd()

	go func() {
		defer t.wgDone() // registered first so it runs last: survives panics and waiters only unblock after cleanup
		if t.deferFunc != nil {
			defer (*t.deferFunc)()
		} else if globalDeferFunc != nil {
			defer (*globalDeferFunc)()
		}
		function(t)
	}()
}
