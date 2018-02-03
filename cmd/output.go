package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kyokomi/emoji"
	"github.com/mgutz/ansi"
)

var (
	debug   = emoji.Sprintf(" :wrench: ")
	info    = emoji.Sprintf(" :information_source: ")
	warning = emoji.Sprintf(" :warning: ")

	blue   = ansi.ColorCode("blue+bh")
	orange = ansi.ColorCode("214+bh")
	red    = ansi.ColorCode("red+bh")
	reset  = ansi.ColorCode("reset")
	white  = ansi.ColorCode("white+bh")
)

// Output represents a channel for sending messages to a user.
type Output interface {
	Debug(string, ...interface{})
	Fatal(error, string, ...interface{})
	Fatals([]error, string, ...interface{})
	Info(string, ...interface{})
	KeyValue(string, string, int, ...interface{})
	LineBreak()
	Say(string, int, ...interface{})
	Table(string, [][]string)
	Warn(string, ...interface{})
}

// ConsoleOutput implements a channel for sending messages to a user over standard output.
type ConsoleOutput struct {
	Color   bool
	Emoji   bool
	Verbose bool
	Test    bool
}

// Debug prints a formatted message to standard output if `Verbose` is set to `true`. Messages are
// prefixed to indicate they are for debugging with :wrench: or [d].
func (c ConsoleOutput) Debug(msg string, a ...interface{}) {
	if c.Verbose {
		switch {
		case c.Emoji && c.Color:
			fmt.Printf(debug+orange+msg+reset+"\n", a...)
		case c.Emoji:
			fmt.Printf(debug+msg+"\n", a...)
		case c.Color:
			fmt.Printf("["+orange+"d"+reset+"] "+orange+msg+reset+"\n", a...)
		default:
			fmt.Printf("[d] "+msg+"\n", a...)
		}
	}
}

// Say prints an optionally indented, formatted message followed by a line break to standard
// output.
func (c ConsoleOutput) Say(msg string, indent int, a ...interface{}) {
	for i := 0; i < indent; i++ {
		fmt.Print(strings.Repeat(" ", 4))
	}

	fmt.Printf(msg+"\n", a...)
}

// Info prints a formatted message to standard output. Messages are prefixed to indicate they are
// informational with :information_source: or [i].
func (c ConsoleOutput) Info(msg string, a ...interface{}) {
	switch {
	case c.Emoji && c.Color:
		fmt.Printf(info+white+msg+reset+"\n", a...)
	case c.Emoji:
		fmt.Printf(info+msg+"\n", a...)
	case c.Color:
		fmt.Printf("["+blue+"i"+reset+"] "+white+msg+reset+"\n", a...)
	default:
		fmt.Printf("[i] "+msg+"\n", a...)
	}
}

// Warn prints a formatted message to standard output. Messages are prefixed to indicate they are
// warnings with :warning: or [!].
func (c ConsoleOutput) Warn(msg string, a ...interface{}) {
	switch {
	case c.Emoji && c.Color:
		fmt.Printf(warning+red+msg+reset+"\n", a...)
	case c.Emoji:
		fmt.Printf(warning+msg+"\n", a...)
	case c.Color:
		fmt.Printf("["+red+"!"+reset+"] "+red+msg+reset+"\n", a...)
	default:
		fmt.Printf("[!] "+msg+"\n", a...)
	}
}

// Fatal prints a formatted message and an error string to standard output  Messages are prefixed
// to indicate they are fatals with :warning: or [!].
func (c ConsoleOutput) Fatal(err error, msg string, a ...interface{}) {
	c.Fatals([]error{err}, msg, a...)
}

// Fatals prints a formatted message and one or more error strings to standard output. Messages
// are prefixed to indicate they are fatals with :warning: or [!].
func (c ConsoleOutput) Fatals(errs []error, msg string, a ...interface{}) {
	c.Warn(msg, a...)

	for _, err := range errs {
		c.Say(err.Error(), 1)
	}

	if !c.Test {
		os.Exit(1)
	}
}

// KeyValue prints a formatted, optionally indented key and value pair to standard output.
func (c ConsoleOutput) KeyValue(key, value string, indent int, a ...interface{}) {
	if c.Color {
		c.Say(white+key+reset+": "+value, indent, a...)
	} else {
		c.Say(key+": "+value, indent, a...)
	}
}

// Table prints a formatted table with optional header to standard output.
func (c ConsoleOutput) Table(header string, rows [][]string) {
	if len(header) > 0 {
		if c.Color {
			c.Say(white+header+reset, 0)
		} else {
			c.Say(header, 0)
		}

		c.LineBreak()
	}

	w := new(tabwriter.Writer)
	defer w.Flush()

	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	for _, row := range rows {
		for i, column := range row {
			fmt.Fprint(w, column)

			if i != len(row)-1 {
				fmt.Fprint(w, "\t")
			}
		}

		fmt.Fprint(w, "\n")
	}

}

// LineBreak prints a single line break.
func (c ConsoleOutput) LineBreak() {
	fmt.Print("\n")
}
