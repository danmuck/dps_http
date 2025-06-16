package logs

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// types
type ServiceLog struct {
	ts  time.Time
	msg string
}

type ServiceLogger struct {
	logs map[string]*ServiceLog
}

var logger *ServiceLogger = &ServiceLogger{
	logs: make(map[string]*ServiceLog),
}

const (
	StyleReset     = "\033[0m"
	StyleBold      = "\033[1m"
	StyleDim       = "\033[2m"
	StyleItalic    = "\033[3m"
	StyleUnderline = "\033[4m"
	StyleBlink     = "\033[5m"
	StyleReverse   = "\033[7m"
	StyleHidden    = "\033[8m"
	// colors
	StyleBlack   = "\033[30m"
	StyleRed     = "\033[31m"
	StyleGreen   = "\033[32m"
	StyleYellow  = "\033[33m"
	StyleBlue    = "\033[34m"
	StyleMagenta = "\033[35m"
	StyleCyan    = "\033[36m"
	StyleWhite   = "\033[37m"
	StyleGray    = "\033[90m"
)

// checks if the log message contains any of the specified filters
// returns true if it does, false otherwise
func LogFilter(format string, filters ...string) bool {
	for _, filter := range filters {
		if strings.Contains(format, filter) {
			return true
		}
	}
	return false
}

// trimes the path to a maximum length, prefixing with "..." if it exceeds the limit
func FormatPath(path string, maxLength int) string {
	const ellipsis = "..."
	if len(path) <= maxLength {
		return fmt.Sprintf("%*s", maxLength, path) // pad left if short
	}

	// Trim from the left, prepend ellipsis
	trimStart := len(path) - (maxLength - len(ellipsis))
	if trimStart < 0 {
		trimStart = 0
	}
	return ellipsis + path[trimStart:]
}

// trims the file path to start from "dps_http/"
func TrimToProjectRoot(root, path string) string {
	root = root + "/"
	idx := strings.Index(path, root)
	if idx == -1 {
		return path // fallback to full path if not found
	}
	return FormatPath(path[idx:], 32)
}

// strips ANSI escape codes from a string
// useful for cleaning colors from logs
func StripANSI(s string) string {
	return regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(s, "")
}

// centers the tag within a given width, padding with spaces
func CenterTag(tag string, width int) string {
	visible := StripANSI(tag)
	tagLen := len(visible)
	if tagLen >= width {
		return tag
	}

	padding := width - tagLen
	left := padding / 2
	right := padding - left

	return strings.Repeat("", left) + tag + strings.Repeat(" ", right)
}

func ColorText(color, text string) string {
	return fmt.Sprintf("%s%s%s", color, text, StyleReset)
}
