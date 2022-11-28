//go:build !trackdebug

package tracker

func getPackage() string {
	return ""
}

// Go starts a tracked go routine and injects a tracker that needs to be used. At a minimum use a select to listen to its Done() channel
func (t Tracker) GoRef(ref string, function func(tkr Tracker)) { // Always call before go routine creation, also always call defer done
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
