package rotatefile

import (
	"log"
)

// To use rotatefile with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(New())
}
