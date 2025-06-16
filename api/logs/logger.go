package logs

import (
	"fmt"
	"strings"
	"time"
)

var (
	LOGFILTER  = []string{"api:users"}
	TIMESTAMP  = false
	SERVICEMAP = map[string]string{
		"api":     "api",
		"users":   "users",
		"metrics": "metrics",
		"auth":    "auth",
	}
)

type ServiceLog struct {
	t   time.Time
	msg string
}
type ServiceLogger struct {
	logs map[string]*ServiceLog
}

var logger *ServiceLogger = &ServiceLogger{
	logs: make(map[string]*ServiceLog),
}

func LogFilter(format string, filters ...string) bool {
	for _, filter := range filters {
		if strings.Contains(format, filter) {
			return true
		}
	}
	return false
}

func Err(format string, v ...any) {
	if TIMESTAMP {
		fmt.Printf("[ERROR]:%s:%s", time.Now().Format(time.Stamp), fmt.Sprintf(format, v...))
	} else {
		fmt.Printf("[ERROR]:%s", fmt.Sprintf(format, v...))
	}
	fmt.Println()
}

func Log(format string, v ...any) {

	log := &ServiceLog{
		t:   time.Now(),
		msg: fmt.Sprintf(format, v...),
	}
	for key, value := range SERVICEMAP {
		if strings.Contains(format, value) {
			logger.logs[key] = log
		}
	}

	if LogFilter(format, LOGFILTER...) {
		if TIMESTAMP {
			fmt.Printf("%s:%s", log.t.Format(time.Stamp), log.msg)
		} else {
			fmt.Printf("%s", log.msg)
		}
		fmt.Println()
	}
}
