//go:build !windows

package rotatefile

import (
	"os"
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
