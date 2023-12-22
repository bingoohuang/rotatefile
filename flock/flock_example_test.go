// Copyright 2015 Tim Heckman. All rights reserved.
// Copyright 2018 The Gofrs. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package flock_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bingoohuang/rotatefile/flock"
)

func ExampleFlock_Locked() {
	f := flock.New(os.TempDir() + "/go-lock.lock")
	locked, err := f.TryLock()

	fmt.Printf("locked error: %v\n", err)
	fmt.Printf("locked: %v\n", locked)
	fmt.Printf("locked: %v\n", f.Locked())

	err = os.Remove(f.Path())
	fmt.Printf("Remove error: %v\n", err)

	err = f.Unlock()
	fmt.Printf("Unlock error: %v\n", err)
	fmt.Printf("locked: %v\n", f.Locked())
	// Output:
	// locked error: <nil>
	// locked: true
	// locked: true
	// Remove error: <nil>
	// Unlock error: <nil>
	// locked: false
}

func ExampleFlock_TryLock() {
	// should probably put these in /var/lock
	fileLock := flock.New(os.TempDir() + "/go-lock.lock")

	locked, err := fileLock.TryLock()

	if err != nil {
		// handle locking error
	}

	if locked {
		fmt.Printf("path: %s; locked: %v\n", fileLock.Path(), fileLock.Locked())

		if err := fileLock.Unlock(); err != nil {
			// handle unlock error
		}
	}

	fmt.Printf("path: %s; locked: %v\n", fileLock.Path(), fileLock.Locked())
}

func ExampleFlock_TryLockContext() {
	// should probably put these in /var/lock
	fileLock := flock.New(os.TempDir() + "/go-lock.lock")

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, 678*time.Millisecond)

	if err != nil {
		// handle locking error
	}

	if locked {
		fmt.Printf("path: %s; locked: %v\n", fileLock.Path(), fileLock.Locked())

		if err := fileLock.Unlock(); err != nil {
			// handle unlock error
		}
	}

	fmt.Printf("path: %s; locked: %v\n", fileLock.Path(), fileLock.Locked())
}
