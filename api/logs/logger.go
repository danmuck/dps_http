package logs

import (
	"fmt"
	"strings"
	"time"

	"github.com/danmuck/dps_http/configs"
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
	fmt.Println()
	if configs.LOGGER_enable_timestamp {
		fmt.Printf("[ERROR]:%s:%s", time.Now().Format(time.Stamp), fmt.Sprintf(format, v...))
	} else {
		fmt.Printf("[ERROR]:%s", fmt.Sprintf(format, v...))
	}
	fmt.Println()
	fmt.Println()
}

func Warn(format string, v ...any) {
	if configs.LOGGER_enable_timestamp {
		fmt.Printf("[WARN]:%s:%s", time.Now().Format(time.Stamp), fmt.Sprintf(format, v...))
	} else {
		fmt.Printf("[WARN]:%s", fmt.Sprintf(format, v...))
	}
	fmt.Println()
}

func Info(format string, v ...any) {
	if configs.LOGGER_enable_timestamp {
		fmt.Printf("[INFO]:%s:%s", time.Now().Format(time.Stamp), fmt.Sprintf(format, v...))
	} else {
		fmt.Printf("[INFO]:%s", fmt.Sprintf(format, v...))
	}
	fmt.Println()
}

func Debug(format string, v ...any) {
	if configs.LOGGER_enable_timestamp {
		fmt.Printf("[DEBUG]:%s:%s", time.Now().Format(time.Stamp), fmt.Sprintf(format, v...))
	} else {
		fmt.Printf("[DEBUG]:%s", fmt.Sprintf(format, v...))
	}
	fmt.Println()
}

func Log(format string, v ...any) {

	log := &ServiceLog{
		t:   time.Now(),
		msg: fmt.Sprintf(format, v...),
	}
	for key, value := range configs.LOGGER_service_map {
		if strings.Contains(format, value) {
			logger.logs[key] = log
		}
	}

	// if LogFilter(format, configs.LOGGER_filter...) {
	if configs.LOGGER_enable_timestamp {
		fmt.Printf("%s:%s", log.t.Format(time.Stamp), log.msg)
	} else {
		fmt.Printf("%s", log.msg)
	}
	fmt.Println()
	// }
}
