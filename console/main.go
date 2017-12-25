package console

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mgutz/ansi"
)

var (
	Verbose = false
	Color   = true
)

var (
	info  = "[i] "
	debug = "[d] "
	issue = "[!] "
	shell = "[>] "

	colorInfo  = white + "[" + blue + "i" + white + "]" + reset + " "
	colorDebug = white + "[" + orange + "d" + white + "]" + reset + " "
	colorIssue = white + "[" + red + "!" + white + "]" + reset + " "
	colorShell = white + "[" + green + ">" + white + "]" + reset + " "

	blue   = ansi.ColorCode("blue+bh")
	white  = ansi.ColorCode("white+bh")
	yellow = ansi.ColorCode("yellow+bh")
	green  = ansi.ColorCode("green+bh")
	red    = ansi.ColorCode("red+bh")
	reset  = ansi.ColorCode("reset")
	orange = ansi.ColorCode("214+bh")
)

func LogLine(prefix, msg string, color int) {
	if Color {
		colorCode := strconv.Itoa(color)
		fmt.Println(ansi.ColorCode(colorCode) + prefix + reset + " " + msg)
	} else {
		fmt.Println(prefix + " " + msg)
	}
}

func KeyValue(key, value string, a ...interface{}) {
	if Color {
		fmt.Fprintf(os.Stdout, white+key+reset+": "+value, a...)
	} else {
		fmt.Fprintf(os.Stdout, key+": "+value, a...)
	}
}

func Header(s string) {
	fmt.Print("\n")

	if Color {
		fmt.Print(white + s + reset + "\n")
	} else {
		fmt.Println(s)
	}
}

func Info(msg string, a ...interface{}) {
	if Color {
		fmt.Fprintf(os.Stdout, colorInfo+msg+reset+"\n", a...)
	} else {
		fmt.Fprintf(os.Stdout, info+msg+"\n", a...)
	}
}

func Debug(msg string, a ...interface{}) {
	if Verbose {
		if Color {
			fmt.Fprintf(os.Stdout, colorDebug+msg+reset+"\n", a...)
		} else {
			fmt.Fprintf(os.Stdout, debug+msg+"\n", a...)
		}
	}
}

func Shell(msg string, a ...interface{}) {
	if Color {
		fmt.Fprintf(os.Stdout, colorShell+green+msg+reset+"\n", a...)
	} else {
		fmt.Fprintf(os.Stdout, shell+msg+"\n", a...)
	}
}

func Issue(msg string, a ...interface{}) {
	if Color {
		fmt.Fprintf(os.Stderr, colorIssue+red+msg+reset+"\n", a...)
	} else {
		fmt.Fprintf(os.Stderr, issue+msg+"\n", a...)
	}
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

func IssueExit(msg string, a ...interface{}) {
	Issue(msg, a...)
	os.Exit(1)
}

func SetVerbose(verbose bool) {
	Verbose = verbose
}
