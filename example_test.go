package rotatefile

import (
	"log"
)

// To use rotatefile with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(&Logger{
		Filename:   "/var/log/myapp/foo.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxDays:    28,   // days
		Compress:   true, // disabled by default
	})
}
