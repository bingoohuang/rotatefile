package autoload

import (
	"log"

	"github.com/bingoohuang/rotatefile"
)

func init() {
	log.SetOutput(rotatefile.NewFile())
}
