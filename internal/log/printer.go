package log

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	dimWhite = color.New(color.FgHiBlack).SprintFunc()
	cyan     = color.New(color.FgCyan).SprintFunc()
	green    = color.New(color.FgGreen).SprintFunc()
	yellow   = color.New(color.FgYellow).SprintFunc()
	red      = color.New(color.FgRed).SprintFunc()
	magenta  = color.New(color.FgMagenta).SprintFunc()
)

func PrintEntry(e LogEntry) {
	ts := dimWhite(e.Timestamp.Format("15:04:05"))
	src := cyan(fmt.Sprintf("%-8s", e.Source))
	lvl := colorLevel(e.Level)
	fmt.Printf("%s %s %s %s\n", ts, src, lvl, e.Message)
}

func colorLevel(l LogLevel) string {
	padded := fmt.Sprintf("%-5s", l)
	switch l {
	case LevelInfo:
		return green(padded)
	case LevelWarn:
		return yellow(padded)
	case LevelError:
		return red(padded)
	case LevelDebug:
		return magenta(padded)
	default:
		return padded
	}
}
