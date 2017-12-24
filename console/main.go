package console

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mgutz/ansi"
)

var (
	Verbose = false
)

var (
	info  = white + "[" + blue + "i" + white + "]" + reset + " "
	debug = white + "[" + orange + "d" + white + "]" + reset + " "
	issue = white + "[" + red + "!" + white + "]" + reset + " "
	shell = white + "[" + green + ">" + white + "]" + reset + " "

	blue   = ansi.ColorCode("blue+bh")
	white  = ansi.ColorCode("white+bh")
	yellow = ansi.ColorCode("yellow+bh")
	green  = ansi.ColorCode("green+bh")
	red    = ansi.ColorCode("red+bh")
	reset  = ansi.ColorCode("reset")
	orange = ansi.ColorCode("214+bh")
)

func LogLine(prefix, msg string, color int) {
	colorCode := strconv.Itoa(color)
	fmt.Println(ansi.ColorCode(colorCode) + prefix + reset + " " + msg)
}

func KeyValue(key, value string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, white+key+reset+": "+value, a...)
}

func Header(s string) {
	fmt.Print("\n")
	fmt.Print(white + s + reset + "\n")
}

func Info(msg string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, info+msg+reset+"\n", a...)
}

func Debug(msg string, a ...interface{}) {
	if Verbose {
		fmt.Fprintf(os.Stdout, debug+msg+reset+"\n", a...)
	}
}

func Shell(msg string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, shell+green+msg+reset+"\n", a...)
}

func Issue(msg string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, issue+red+msg+reset+"\n", a...)
}

func Error(err error, msg string, a ...interface{}) {
	Issue(msg, a...)

	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
	}
}

func ErrorExit(err error, msg string, a ...interface{}) {
	Error(err, msg, a...)
	os.Exit(1)
}

func SetVerbose(verbose bool) {
	Verbose = verbose
}
