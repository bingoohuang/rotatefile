//go:build linux
// +build linux

package rotatefile

import (
	"log"
	"os"
	"syscall"
)

// Example of how to rotate in response to SIGHUP.
func ExampleLogger_Rotate() {
	l := &file{
		RotateSignals: []os.Signal{syscall.SIGHUP},
	}
	log.SetOutput(l)
}
