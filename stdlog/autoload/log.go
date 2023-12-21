package autoload

import (
	"log"

	"github.com/bingoohuang/rotatefile"
	"github.com/bingoohuang/rotatefile/stdlog"
)

func init() {
	log.SetOutput(stdlog.NewLevelLog(rotatefile.NewFile()))
}
