# rotatefile

支持：全默认配置接入、按天滚动、按大小滚动、过期删除、日志总大小控制、磁盘空余控制、日志锁定（防止并发写日志）

rotatefile a Go package for writing logs to rolling files.

package rotatefile provides a rolling rotatefile.File.

rotatefile is intended to be one part of a logging infrastructure.
It is not an all-in-one solution, but instead is a pluggable
component at the bottom of the logging stack that simply controls the files
to which logs are written.

rotatefile plays well with any logging package that can write to an
io.Writer, including the standard library's log package.

rotatefile assumes that only one process is writing to the output files.
Using the same rotatefile configuration from multiple processes on the same
machine will result in improper behavior.

**Example**

To use rotatefile with the standard library's log package, just pass it into the SetOutput function when your
application starts.

Code:

```go
package main

import (
	"log"

	"github.com/bingoohuang/rotatefile"
)

func init() {
	log.SetOutput(rotatefile.NewFile(
		// rotatefile.WithFilename("/var/log/myapp/foo.log"), // 指定日志文件完整路径, 默认值见 Config.Filename 说明
		// rotatefile.WithMaxSize(100*1024),      // 单个日志文件最大大小，默认 100M
		// rotatefile.WithMaxDays(30),            // 最多保留天数，默认值30
		// rotatefile.WithTotalSizeCap(1024*1024*1024), // 最大总大小，默认 1G
		// rotatefile.WithMinDiskFree(300*1024),  // 最少磁盘空余，默认 100M
		// 以上默认值，还可以通过环境变量设置，参照环境变量说明
	))
}
```

## 环境变量

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

## type rotatefile.Config

``` go
type Config struct {
    // Filename is the file to write logs to. Backup log files will be retained
    // in the same directory.  
    // 如果设置为空，则自动按顺序在以下目录中写入（找到第一个可用目录为止)，具体位置可以见 $TMPDIR/{pid}.logfile 文件
    // 1. $HOME/log/{appName}/{appName}_{appWorkDirBase}.log
    // 2. $PWD/log/{appName}_{appWorkDirBase}.log
    // 3. /var/log/apps/{appName}/{appName}_{appWorkDirBase}.log
    // 4. $TMPDIR/{appName}/{appName}_{appWorkDirBase}.log
    Filename string `json:"filename" yaml:"filename"`

    // MaxSize is the maximum size of the log file before it gets
    // rotated. It defaults to 100 megabytes.
    MaxSize int `json:"maxSize" yaml:"maxSize"`

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
	TotalSizeCap int64 `json:"totalSizeCap" yaml:"totalSizeCap"`

    // UtcTime determines if the time used for formatting the timestamps in
    // backup files is the computer's local time. 
    // The default is to not to use local time.
    UtcTime bool `json:"utcTime" yaml:"utcTime"`

    // Compress determines if the rotated log files should be compressed
    // using gzip. 
    // The default is to perform compression.
    Compress bool `json:"compress" yaml:"compress"`
   
    // contains filtered or unexported fields
}
```

rotatefile.File is an io.WriteCloser that writes to the specified filename.

rotatefile.File opens or creates the logfile on first Write. If the file exists and
is fewer than MaxSize megabytes, rotatefile will open and append to that file.
If the file exists and its size is >= MaxSize megabytes, the file is renamed
by putting the current time in a timestamp in the name immediately before the
file's extension (or the end of the filename if there's no extension). A new
log file is then created using original filename.

Whenever a write would cause the current log file to exceed MaxSize megabytes,
the current file is closed, renamed, and a new log file created with the
original name. Thus, the filename you give rotatefile. File is always the "current" log
file.

Backups use the log file name given to rotatefile.File, in the form `name.timestamp.ext`
where name is the filename without the extension, timestamp is the time at which
the log was rotated formatted with the time.Time format of
`20060102T150405.000` and the extension is the original extension. For
example, if your rotatefile.File.Filename is `/var/log/foo/server.log`, a backup created
at 6:30pm on Nov 11 2016 would use the filename
`/var/log/foo/server.20161104T183000.000.log`

### Cleaning Up Old Log Files

Whenever a new logfile gets created, old log files may be deleted. The most
recent files according to the encoded timestamp will be retained, up to a
number equal to MaxBackups (or all of them if MaxBackups is 0). Any files
with an encoded timestamp older than MaxDays days are deleted, regardless of
MaxBackups. Note that the time encoded in the timestamp is the rotation
time, which may differ from the last time that file was written to.

If MaxBackups and MaxDays are both 0, no old log files will be deleted.
