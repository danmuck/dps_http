package logs

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/danmuck/dps_http/configs"
)

func ColorTest() {
	Err("This is an error message")
	Warn("This is a warning message")
	Info("This is an info message")
	Debug("This is a debug message")
	Dev("This is a dev message")
	Init("This is an init message")
}

func Log(format string, v ...any) {
	log := &ServiceLog{
		ts:  time.Now(),
		msg: fmt.Sprintf(format, v...),
	}
	for key, value := range configs.LOGGER_service_map {
		if strings.Contains(format, value) {
			logger.logs[key] = log
		}
	}
	Print(StyleWhite, "[logs]", format, v...)
}

func Err(format string, v ...any) {
	Print(StyleRed, "[error]", format, v...)
}

func Warn(format string, v ...any) {
	Print(StyleYellow, "[warn]", format, v...)
}

func Info(format string, v ...any) {
	Print(StyleBlue, "[info]", format, v...)
}

func Debug(format string, v ...any) {
	Print(StyleGreen, "[debug]", format, v...)
}

func Dev(format string, v ...any) {
	Print(StyleMagenta, "[dev_]", format, v...)
}

func Init(format string, v ...any) {
	Print(StyleBlack, "[init]", format, v...)
}

// T is the type of log, e.g. "dev", "error", "warn", etc.
// format, v... are the format string and values to Print
func Print(C, T, format string, v ...any) {
	if configs.LOGGER_enable_timestamp {
		fmt.Println(ColorText(C, String(C, T, format, v...)))
	} else {
		fmt.Println(ColorText(C, String(C, T, format, v...)))
	}

}

func String(C, T, format string, v ...any) string {
	msg := fmt.Sprintf(format, v...)
	ts := time.Now().Format(time.Stamp)

	_, file, line, ok := runtime.Caller(3)
	if ok {
		path := TrimToProjectRoot("dps_http", file)            // max 32 chars
		tag := CenterTag(T, 9)                                 // padded 9 and centered tag
		lineStr := fmt.Sprintf(":%4d", line)                   // pad line number
		prefix := fmt.Sprintf("%s[%s%s] ", tag, path, lineStr) // final prefix
		if configs.LOGGER_enable_timestamp {
			return fmt.Sprintf("%s %s%s", ts, prefix, msg)
		}
		return fmt.Sprintf("%s%s", prefix, msg)
	}
	return msg
}
