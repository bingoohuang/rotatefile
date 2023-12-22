# flock

`flock` implements a thread-safe sync.Locker interface for file locking. It also
includes a non-blocking TryLock() function to allow locking without blocking execution.

## Go Compatibility

This package makes use of the `context` package that was introduced in Go 1.7. As such, this
package has an implicit dependency on Go 1.7+.

## Usage

```Go
import "github.com/bingoohuang/rotatefile/flock"

fileLock := flock.New("/var/lock/go-lock.lock")
locked, err := fileLock.TryLock()
if err != nil {
// handle locking error
}

if locked {
// do work
fileLock.Unlock()
}
```

