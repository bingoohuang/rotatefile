package rotatefile

import (
	"os"
	"syscall"

	"github.com/bingoohuang/rotatefile/homedir"
	"golang.org/x/term"
)

const (
	// MB is mega
	MB = 1024 * 1024
	// GB is giga
	GB = 1024 * MB

	backupTimeFormat = "20060102T150405.000"
	compressSuffix   = ".gz"
	defaultMaxSize   = 100 * 1024 * 1024
)

func createConfig(fns ...ConfigFn) Config {
	c := Config{
		Filename:      os.Getenv("LOG_FILENAME"),
		RotateSignals: EnvSignals("LOG_ROTATE_SIGNALS", []os.Signal{syscall.SIGHUP}),
		MaxSize:       EnvSize("LOG_MAX_SIZE", 100*MB),
		MaxDays:       EnvInt("LOG_MAX_DAYS", 30),
		MaxBackups:    EnvInt("LOG_MAX_BACKUPS", 0),
		TotalSizeCap:  EnvSize("LOG_TOTAL_SIZE_CAP", GB),
		MinDiskFree:   EnvSize("LOG_MIN_DISK_FREE", 100*MB),
		UtcTime:       EnvBool("LOG_UTCTIME", false),
		Compress:      EnvBool("LOG_COMPRESS", true),
		PrintTerm:     EnvBool("LOG_PRINT_TERM", IsTerminal),
	}

	for _, f := range fns {
		f(&c)
	}

	return c
}

// IsTerminal tell is if it is on a terminal.
var IsTerminal = term.IsTerminal(1)

// Config 包括一些滚动文件的配置参数，所有参数，均有默认值，方便无脑集成
type Config struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>.log in os.TempDir() if empty.
	Filename string `json:"filename" yaml:"filename"`
	// Prefix 是日志基本文件名前缀，在 Filename 不指定的情况下，可以使用本字段给自动生成的日志文件名添加此前缀
	Prefix string `json:"prefix" yaml:"prefix"`

	// RotateSignals 设置滚动日志的信号
	RotateSignals []os.Signal `json:"-" yaml:"-"`

	// MaxSize is the maximum size of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize uint64 `json:"maxSize" yaml:"maxSize"`

	// MaxDays is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxDays int `json:"maxDays" yaml:"maxDays"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxDays may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxBackups" yaml:"maxBackups"`

	// TotalSizeCap 控制所有文件累积总大小
	// 如果超过该大小，则从最早的文件开始删除，直到删除到当前文件为止
	// 当前日志文件大小可以超过 TotalSizeCap
	// 0 不控制
	TotalSizeCap uint64 `json:"totalSizeCap" yaml:"totalSizeCap"`

	// MinDiskFree 日志文件所在磁盘分区最少空余
	MinDiskFree uint64 `json:"minDiskFree" yaml:"minDiskFree"`

	// UtcTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	UtcTime bool `json:"utcTime" yaml:"utcTime"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"`

	// PrintTerm 是否同时在终端上输出，只有在终端可用时输出
	PrintTerm bool `json:"printTerm" yaml:"printTerm"`
}

// ConfigFn 选项模式函数
type ConfigFn func(*Config)

// WithConfig 指定 Config 参数对象
func WithConfig(v Config) ConfigFn { return func(c *Config) { *c = v } }

// WithPrintTerm 指定是否同时打印到控制台
func WithPrintTerm(v bool) ConfigFn { return func(c *Config) { c.PrintTerm = v } }

// WithCompress 指定是否开启压缩
func WithCompress(v bool) ConfigFn { return func(c *Config) { c.Compress = v } }

// WithUtcTime 指定是否使用 UTC 时间
func WithUtcTime(v bool) ConfigFn { return func(c *Config) { c.UtcTime = v } }

// WithMinDiskFree 指定最小磁盘可用大小
func WithMinDiskFree(v uint64) ConfigFn { return func(c *Config) { c.MinDiskFree = v } }

// WithTotalSizeCap 指定日志总和大小上限
func WithTotalSizeCap(v uint64) ConfigFn { return func(c *Config) { c.TotalSizeCap = v } }

// WithMaxBackups 指定最大备份文件数量
func WithMaxBackups(v int) ConfigFn { return func(c *Config) { c.MaxBackups = v } }

// WithMaxDays 指定最大保存天数
func WithMaxDays(v int) ConfigFn { return func(c *Config) { c.MaxDays = v } }

// WithMaxSize 指定日志文件最大大小
func WithMaxSize(v uint64) ConfigFn { return func(c *Config) { c.MaxSize = v } }

// WithFilename 指定日志文件名字
func WithFilename(v string) ConfigFn {
	return func(c *Config) {
		if expanded, err := homedir.Expand(v); err == nil {
			c.Filename = expanded
		} else {
			c.Filename = v
		}
	}
}

// WithPrefix 指定日志基本文件名前缀，在 Filename 不指定的情况下，可以使用本字段给自动生成的日志文件名添加此前缀
func WithPrefix(v string) ConfigFn {
	return func(c *Config) {
		c.Prefix = v
	}
}

// WithRotateSignals 指定强制滚动信号
func WithRotateSignals(v ...os.Signal) ConfigFn { return func(c *Config) { c.RotateSignals = v } }
