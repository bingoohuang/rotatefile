package rotatefile

import (
	"log"
	"os"
	"syscall"
)

// To use rotatefile with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(&File{
		Filename:      "/var/log/myapp/foo.log",
		MaxSize:       100 * MB,                    // 单个日志文件最大100M
		MaxBackups:    30,                          // 最多30个历史备份
		MaxDays:       30,                          // 最多保留30天
		TotalSizeCap:  GB,                          // 最大总大小1G
		Compress:      true,                        // 历史日志开启 Gzip 压缩
		MinDiskFree:   100 * MB,                    // 最少 100M 空余
		RotateSignals: []os.Signal{syscall.SIGHUP}, // 在收到 SIGHUP 时，滚动日志
	})
}
