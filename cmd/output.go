package cmd

import (
	"fmt"
	"os"
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

type ConsoleOutput struct {
	Color   bool
	Emoji   bool
	Verbose bool
}

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

func (c ConsoleOutput) Say(msg string, indent int, a ...interface{}) {
	for i := 0; i < indent; i++ {
		fmt.Print("    ")
	}

	fmt.Printf(msg, a...)
	fmt.Print("\n")
}

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

func (c ConsoleOutput) Fatal(err error, msg string, a ...interface{}) {
	c.Warn(msg, a...)

	if err != nil {
		c.Say(err.Error(), 1)
	}

	os.Exit(1)
}

func (c ConsoleOutput) Fatals(errs []error, msg string, a ...interface{}) {
	c.Warn(msg, a...)

	for _, err := range errs {
		c.Say(err.Error(), 1)
	}

	os.Exit(1)
}

func (c ConsoleOutput) KeyValue(key, value string, indent int, a ...interface{}) {
	if c.Color {
		c.Say(white+key+reset+": "+value, indent, a...)
	} else {
		c.Say(key+": "+value, indent, a...)
	}
}

func (c ConsoleOutput) Table(header string, rows [][]string) {
	if c.Color {
		c.Say(white+header+reset, 0)
	} else {
		c.Say(header, 0)
	}

	c.LineBreak()

	w := new(tabwriter.Writer)
	defer w.Flush()

	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	for _, row := range rows {
		for _, column := range row {
			fmt.Fprint(w, column+"\t")
		}

		fmt.Fprint(w, "\n")
	}

}

func (c ConsoleOutput) LineBreak() {
	fmt.Print("\n")
}
