package rotatefile

import (
	"fmt"
	"github.com/ryboe/q"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// EnvSignals 解析环境变量设置的信号
func EnvSignals(envName string, defaultValue []os.Signal) []os.Signal {
	s := os.Getenv(envName)
	if s == "" {
		return defaultValue
	}
	var signals []os.Signal
	splits := strings.Split(s, ",")
	for _, item := range splits {
		switch strings.ToUpper(item) {
		case "SIGHUP":
			signals = append(signals, syscall.SIGHUP)
		case "SIGUSR1":
			signals = append(signals, syscall.SIGUSR1)
		case "SIGUSR2":
			signals = append(signals, syscall.SIGUSR2)
		}
	}

	return signals
}

// EnvBool 解析环境变量设置的 bool 类型变量
func EnvBool(envName string, defaultValue bool) bool {
	switch s := os.Getenv(envName); strings.ToLower(s) {
	case "yes", "y", "1", "on", "true", "t":
		return true
	case "no", "n", "0", "off", "false", "f":
		return false
	}
	return defaultValue
}

// EnvInt 解析环境变量设置的 int 类型变量
func EnvInt(envName string, defaultValue int) int {
	if s := os.Getenv(envName); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return defaultValue
		}
		return v
	}
	return defaultValue
}

// EnvSize 解析环境变量设置的字节大小类型的变量
func EnvSize(envName string, defaultValue uint64) uint64 {
	if s := os.Getenv(envName); s != "" {
		size, err := ParseBytes(s)
		if err != nil {
			return defaultValue
		}

		return size
	}
	return defaultValue
}

func Debugf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	q.Q(s)
}
