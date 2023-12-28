package rotatefile

import (
	"github.com/bingoohuang/q"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/bingoohuang/rotatefile/flock"
)

// getLogFileName 获取可执行文件 binName 的日志文件路径
func getLogFileName(logDir, prefix, logName string, tryLock bool) (string, *flock.Flock) {
	if p := FindLogDir(logDir); p != "" {
		if logName == "" {
			appName := filepath.Base(os.Args[0])
			logName = appName + currentDirBase + ".log"
		}

		var logLock *flock.Flock
		if tryLock {
			logLock = flock.New(filepath.Join(p, logName+".lock"))
			if lock, _ := logLock.TryLock(); !lock {
				logName = logName[:len(logName)-len(".log")] + "." + pid + ".log"
			}
		}
		logFileName := filepath.Join(p, prefix+logName)
		writeLogFile(logFileName)
		return logFileName, logLock
	}

	panic("日志已经无处安放，君欲何为？")
}

// GetFilename 获得当前进程的日志文件路径
func GetFilename() string {
	logdirFile := filepath.Join(os.TempDir(), pid+".logfile")
	logfile, _ := os.ReadFile(logdirFile)
	return string(logfile)
}

var pid = strconv.Itoa(os.Getpid())

func writeLogFile(logFileName string) {
	q.Q(logFileName)
	logdirFile := filepath.Join(os.TempDir(), pid+".logfile")
	_ = q.AppendFile(logdirFile, []byte(logFileName+"\n"), os.ModePerm)
}

func handleSigint(f func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)
	go func() {
		<-ch
		f()
	}()
}

// FindLogDir 寻找日志合理的写入目录
// 0. 配置指定目录 /var/log/xxx/
// 1. $HOME/log/{appName}/{appName}_{appWorkDirBase}.log
// 2. $PWD/log/{appName}_{appWorkDirBase}.log
// 3. /var/log/apps/{appName}/{appName}_{appWorkDirBase}.log
// 4. $TMPDIR/{appName}/{appName}_{appWorkDirBase}.log
func FindLogDir(logDir string) string {
	if logDir != "" {
		if IsDirWritable(logDir) {
			return logDir
		}
	}

	appName := filepath.Base(os.Args[0])
	if home, _ := HomeDir(); home != "" {
		if p := filepath.Join(home, "log", appName); IsDirWritable(p) {
			return p
		}
	}
	if wd, _ := os.Getwd(); wd != "" {
		if p := filepath.Join(wd, "log", appName); IsDirWritable(p) {
			return p
		}
	}
	if p := filepath.Join("/var/log/apps", appName); IsDirWritable(p) {
		return p
	}
	if p := os.TempDir(); IsDirWritable(p) {
		return p
	}
	return ""
}

// IsDirWritable 测试目录是否可写
func IsDirWritable(dir string) bool {
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err := MkdirAll(dir, os.ModePerm); err != nil {
			return false
		}
	}

	temp, err := os.CreateTemp(dir, "*")
	if err != nil {
		return false
	}
	defer func() {
		temp.Close()
		os.Remove(temp.Name())
	}()

	const s = "bingoohuang_log_test"
	n, err := temp.WriteString(s)
	return err == nil && n == len(s)
}

var currentDirBase = func() string {
	wd, _ := os.Getwd()
	if wd != "" {
		return "_" + filepath.Base(wd)
	}
	return ""
}()

// HomeDir 返回当前用户的主目录
// 注意：有可能有的Linux用户没有主目录，此时，日志文件可能需要产生在 /var/log 目录下
func HomeDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	return u.HomeDir, nil
}
