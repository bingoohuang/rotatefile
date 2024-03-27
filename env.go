package rotatefile

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bingoohuang/q"
)

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

// Env 解析环境变量设置的 string 类型变量
func Env(envName, defaultValue string) string {
	if s := os.Getenv(envName); s != "" {
		return s
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

// Debugf print debug info.
func Debugf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	q.Q(s)
}
