# tracker

[![Go Reference](https://pkg.go.dev/badge/github.com/s00500/tracker.svg)](https://pkg.go.dev/github.com/s00500/tracker)

A small Go library for cancelling trees of goroutines and waiting for them to
finish.

On projects with many goroutines it gets hard to keep track of a growing
"cancel-tree" and wire up a `WaitGroup` for every layer. `tracker` combines a
`context.Context` and a `sync.WaitGroup` behind one interface, with an API for
starting goroutines similar to `errgroup` from `golang.org/x/sync`.

## Install

```bash
go get github.com/s00500/tracker
```

## Usage

```go
t := tracker.Root()

t.Go(func(tkr tracker.Tracker) {
    for {
        select {
        case <-tkr.Done():
            return // cancelled — clean up and exit
        default:
            // do work
        }
    }
})

// Create a sub-group whose goroutines are cancelled together with the parent.
sub := t.NewSubGroup()
sub.Go(func(tkr tracker.Tracker) {
    <-tkr.Done()
})

// Cancel everything and block until all goroutines have returned.
t.CancelAndWait()
```

Key operations:

- `Root()` / `FromContext(ctx)` — create a top-level tracker.
- `Go` / `GoRef` — start a tracked goroutine; `Run` runs synchronously.
- `NewSubGroup` — derive a child tracker that cancels with its parent.
- `Cancel`, `Wait`, `CancelAndWait` — stop and/or wait for goroutines.
- `Context()` — get the underlying `context.Context` for external libraries.

See the full API on [pkg.go.dev](https://pkg.go.dev/github.com/s00500/tracker).

Feedback on the design is welcome.

## License

MIT — see [LICENSE](LICENSE). You may use, modify, and distribute the software
freely; the only requirement is that the copyright notice and author
attribution be retained in copies.
