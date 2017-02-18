## Compiling

```sh
$ glide install
$ go build -o scat ./cmd
```

Running all tests:

```sh
$ tools/test
```

## Writing a new proc

* **delegator:** a(b) calls b, filters its results

	In `a.Process()`:
	
	```go
	// ...do things before
	ch := b.Process(c)    // within current goroutine
	out := make(chan Res) // after ch
	go func() {
	  defer close(out)
	  // ...consume ch
	}()
	return out
	```
	
	In `a.Finish()`: must call `b.Finish()`

	When either `a.Process()` or `a.Finish()` (or both) does nothing else than delegating to `b`, then `a` should be a `struct` that embeds `Proc` to inherit those methods and benefit from implicit delegation.
