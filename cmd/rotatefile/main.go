package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	_ "github.com/bingoohuang/rotatefile/stdlog/autoload"
)

func main() {
	flag.Bool("v", false, `
通过环境变量设置：

| 序号 | 变量名                | 默认值                       | 含义              |
|----|--------------------|---------------------------|-----------------|
| 1  | LOG_APPNAME        | filepath.Base(os.Args[0]) | 日志基础文件名         |
| 2  | LOG_FILENAME       | 见下面 Config.Filename 说明    | 日志文件完整路径        |
| 3  | LOG_ROTATE_SIGNALS | SIGHUP                    | 强制当前日志滚动信号      |
| 4  | LOG_MAX_SIZE       | 100M                      | 单个日志文件最大大小      |
| 5  | LOG_MAX_DAYS       | 30                        | 最多保留天数          |
| 6  | LOG_MAX_BACKUPS    | 0                         | 最大历史文件个数        |
| 7  | LOG_TOTAL_SIZE_CAP | 1G                        | 最大总大小           |
| 8  | LOG_MIN_DISK_FREE  | 100M                      | 最少磁盘空余          |
| 9  | LOG_UTCTIME        | 0                         | 是否使用 UTC 时间     |
| 10 | LOG_COMPRESS       | 1                         | 是否启用gzip 压缩历史文件 |
| 11 | LOG_PRINT_TERM     | 根据进程是否有终端                 | 同时在终端打印         |

`)
	flag.Parse()

	for {
		log.Printf("I! %s", RandStringBytesMaskImprSrc(1024))
		time.Sleep(time.Second)
	}
}

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
