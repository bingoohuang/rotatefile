package autoload

import (
	"log"

	"github.com/bingoohuang/rotatefile"
	"github.com/bingoohuang/rotatefile/stdlog"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(stdlog.NewLevelLog(rotatefile.NewFile()))
}
