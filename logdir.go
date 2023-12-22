package rotatefile

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// getLogFileName 获取可执行文件 binName 的日志文件路径
func getLogFileName(logDir, logName string) string {
	if p := FindLogDir(logDir); p != "" {
		appName := filepath.Base(os.Args[0])
		if logName == "" {
			logName = appName + currentDirBase + ".log"
		}
		logFileName := filepath.Join(p, logName)
		writeLogFile(logFileName)
		return logFileName
	}

	panic("日志已经无处安放，君欲何为？")
}

func writeLogFile(logFileName string) {
	logdirFile := filepath.Join(os.TempDir(), strconv.Itoa(os.Getpid())+".logfile")
	_ = os.WriteFile(logdirFile, []byte(logFileName), os.ModePerm)
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
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
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
