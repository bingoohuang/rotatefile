package rotatefile

import "os"

func MkdirAll(dir string, mod os.FileMode) error {
	return os.MkdirAll(dir, mod)
}
