package mock

import "fmt"

type Output struct {
	DebugMsgs []string
	Exited    bool
	FatalMsgs map[string][]error
	SayMsgs   []string
	InfoMsgs  []string
	WarnMsgs  []string
}

func (c *Output) Info(msg string, a ...interface{}) {
	c.InfoMsgs = append(c.InfoMsgs, fmt.Sprintf(msg, a...))
}

func (c *Output) Warn(msg string, a ...interface{}) {
	c.WarnMsgs = append(c.WarnMsgs, fmt.Sprintf(msg, a...))
}

func (c *Output) Fatal(err error, msg string, a ...interface{}) {
	if c.FatalMsgs == nil {
		c.FatalMsgs = make(map[string][]error)
	}

	c.FatalMsgs[fmt.Sprintf(msg, a...)] = []error{err}
	c.Exited = true
}

func (c *Output) Fatals(errs []error, msg string, a ...interface{}) {
	if c.FatalMsgs == nil {
		c.FatalMsgs = make(map[string][]error)
	}

	c.FatalMsgs[fmt.Sprintf(msg, a...)] = errs
	c.Exited = true
}

func (c *Output) Say(msg string, indent int, a ...interface{}) {
	c.SayMsgs = append(c.SayMsgs, fmt.Sprintf(msg, a...))
}

func (c *Output) Debug(msg string, a ...interface{}) {
	c.DebugMsgs = append(c.DebugMsgs, fmt.Sprintf(msg, a...))
}

func (c *Output) LineBreak() {
}
