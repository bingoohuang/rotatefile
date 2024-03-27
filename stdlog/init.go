package stdlog

import (
	"io"
	"log"

	"github.com/bingoohuang/rotatefile"
)

var (
	RotateWriter rotatefile.RotateFile
	LevelLog     io.Writer
)

// Init initialize rotate log module.
func Init(fns ...rotatefile.ConfigFn) {
	log.SetFlags(0)
	log.SetPrefix("")
	RotateWriter = rotatefile.New(fns...)
	LevelLog = NewLevelLog(RotateWriter)
	log.SetOutput(LevelLog)
}
