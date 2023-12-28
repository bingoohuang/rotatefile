//go:build !windows

package rotatefile

import (
	"os"
	"syscall"
)

func MkdirAll(dir string, mod os.FileMode) error {
	// https://github.com/g10guang/g10guang.github.io/blob/master/src/md/2018-02-08-golang-make-dir-problem.md
	mask := syscall.Umask(0)  // 改为 0000 八进制
	defer syscall.Umask(mask) // 改为原来的 umask
	return os.MkdirAll(dir, mod)
}
