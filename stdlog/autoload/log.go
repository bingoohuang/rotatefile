package autoload

import (
	"io"
	"log"

	"github.com/bingoohuang/rotatefile"
	"github.com/bingoohuang/rotatefile/stdlog"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("")
	RotateWriter = rotatefile.NewFile()
	LevelLog = stdlog.NewLevelLog(RotateWriter)
	log.SetOutput(LevelLog)
}

var (
	RotateWriter rotatefile.RotateFile
	LevelLog     io.Writer
)
