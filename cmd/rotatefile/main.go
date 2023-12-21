package main

import (
	"log"
	"math/rand"
	"os"
	"syscall"
	"time"

	"github.com/bingoohuang/rotatefile"
)

func main() {
	log.SetOutput(&rotatefile.File{
		MaxSize:       100 * 1024,                  // 单个日志文件最大100K
		MaxBackups:    5,                           // 最多5个历史备份
		MaxDays:       30,                          // 最多保留30天
		TotalSizeCap:  300 * 1024,                  // 最大总大小1G
		Compress:      true,                        // 历史日志开启 Gzip 压缩
		MinDiskFree:   100 * 1024,                  // 最少 100M 空余
		RotateSignals: []os.Signal{syscall.SIGHUP}, // 在收到 SIGHUP 时，滚动日志
	})

	for {
		log.Print(RandStringBytesMaskImprSrc(1024))
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
