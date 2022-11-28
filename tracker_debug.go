//go:build trackdebug

package tracker

import (
	"fmt"
	"runtime"

	"github.com/google/uuid"
)

func getPackage() string {
	// we get the callers as uintptrs - but we just need 1
	fpcs := make([]uintptr, 1)

	// skip 4 levels to get to the caller of whoever called getPackage()
	n := runtime.Callers(4, fpcs)
	if n == 0 {
		return ""
	}

	// get the info of the actual function that's in the pointer
	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return ""
	}

	name := fun.Name()
	return name
}

// Go starts a tracked go routine and injects a tracker that needs to be used. At a minimum use a select to listen to its Done() channel
func (t Tracker) GoRef(ref string, function func(tkr Tracker)) { // Always call before go routine creation, also always call defer done
	t.wgAdd()

	source := getPackage()
	routineid := uuid.New().String()
	if t.Logging {
		fmt.Printf("Start %s-%s from %s\n", routineid, ref, source)
	}

	go func() {
		if t.deferFunc != nil {
			defer (*t.deferFunc)()
		} else if globalDeferFunc != nil {
			defer (*globalDeferFunc)()
		}
		function(t)
		t.wgDone()
		if t.Logging {
			fmt.Printf("Stop %s-%s from %s\n", routineid, ref, source)
		}
	}()
}
