package printer

import (
	"fmt"
	"io"
	"os"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-wordwrap"
)

const (
	wordWrap = 80
)

var (
	noteColor                = aurora.Blue
	highlightColor           = aurora.Red
	errorColor               = aurora.Red
	stderr         io.Writer = os.Stderr
	writer                   = func() io.Writer {
		if f, ok := stderr.(*os.File); ok {
			fd := f.Fd()
			if isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd) {
				return colorable.NewColorable(f)
			}
		}
		return colorable.NewNonColorable(stderr)
	}()
)

// Printf prints formated message to stdout wrapped
func Printf(format string, a ...interface{}) {
	fmt.Print(wordwrap.WrapString(fmt.Sprintf(format, a...), wordWrap))
}

// Print prints message to stdout wrapped
func Print(a ...interface{}) {
	fmt.Print(wordwrap.WrapString(fmt.Sprint(a...), wordWrap))
}

// Println prints message to stdout wrapped and follows it with new line
func Println(a ...interface{}) {
	fmt.Println(wordwrap.WrapString(fmt.Sprint(a...), wordWrap))
}

// ColorPrintf prints formated message to stdout wrapped and colored
func ColorPrintf(format string, args ...interface{}) {
	colorPrintf(nil, highlightColor, format, args...)
}

// ColorPrint prints message to stdout wrapped and colored
func ColorPrint(args ...interface{}) {
	colorPrint(nil, args...)
}

// ColorPrintln prints message to stdout wrapped and colored followed by new line
func ColorPrintln(args ...interface{}) {
	colorPrintln(nil, args...)
}

// NotePrintf prints formated message to stdout wrapped and colored
func NotePrintf(format string, args ...interface{}) {
	colorPrintf(noteColor, highlightColor, format, args...)
}

// NotePrint prints message to stdout wrapped and colored
func NotePrint(args ...interface{}) {
	colorPrint(noteColor, args...)
}

// NotePrintln prints message to stdout wrapped and colored followed by new line
func NotePrintln(args ...interface{}) {
	colorPrintln(noteColor, args...)
}

// ErrorPrintf prints formated message to stdout wrapped and colored
func ErrorPrintf(format string, args ...interface{}) {
	colorPrintf(errorColor, nil, format, args...)
}

// ErrorPrint prints message to stdout wrapped and colored
func ErrorPrint(args ...interface{}) {
	colorPrint(errorColor, args...)
}

// ErrorPrintln prints message to stdout wrapped and colored followed by new line
func ErrorPrintln(args ...interface{}) {
	colorPrintln(errorColor, args...)
}

func colorize(color func(a interface{}) aurora.Value, args ...interface{}) interface{} {
	s := fmt.Sprint(args...)
	var t interface{} = wordwrap.WrapString(s, wordWrap)
	if color != nil {
		t = color(t)
	}
	return t
}

func colorPrintf(
	baseColor func(a interface{}) aurora.Value,
	argColor func(a interface{}) aurora.Value,
	format string,
	args ...interface{},
) {
	hArgs := []interface{}{}
	for _, a := range args {
		if argColor != nil {
			hArgs = append(hArgs, argColor(a))
		} else {
			hArgs = append(hArgs, a)
		}
	}
	var t interface{} = format
	if baseColor != nil {
		t = baseColor(t)
	}
	fmt.Print(wordwrap.WrapString(
		aurora.Sprintf(t, hArgs...), wordWrap,
	))
}

func colorPrint(color func(a interface{}) aurora.Value, args ...interface{}) {
	fmt.Print(colorize(color, args...))
}

func colorPrintln(color func(a interface{}) aurora.Value, args ...interface{}) {
	fmt.Println(colorize(color, args...))
}
