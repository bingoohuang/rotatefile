package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/bingoohuang/rotatefile"
	"github.com/bingoohuang/rotatefile/stdlog"
)

func main() {
	logdir := flag.String("logdir", "", "log dir")
	flag.Parse()

	f := rotatefile.New(
		rotatefile.WithFilename(*logdir),      // 指定日志目录
		rotatefile.WithMaxSize(100*1024),      // 单个日志文件最大100K
		rotatefile.WithMaxDays(30),            // 最多保留30天
		rotatefile.WithTotalSizeCap(300*1024), // 最大总大小300K
		rotatefile.WithMinDiskFree(300*1024),  // 最少 100M 磁盘空余
		rotatefile.WithPrintTerm(false),       // 屏幕不打印输出
	)
	log.SetOutput(stdlog.NewLevelLog(f))
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
