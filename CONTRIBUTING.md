## Compiling

```sh
$ glide install
$ go build -o scat ./cmd
```

Running tests:

```sh
$ tools/test
```

## Writing a new proc

* **delegator:** a(b) calls b, filters its results

	In `a.Process()`:
	
	```go
	// ... do things before
	ch := b.Process(c)    // within current goroutine
	out := make(chan Res) // after ch
	go func() {
	  defer close(out)
	  // ... consume ch
	}()
	return out
	```
	
	In `a.Finish()`:
	
	Must call `b.Finish()`
