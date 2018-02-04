package mock

import "fmt"

type Output struct {
	DebugMsgs    []string
	Exited       bool
	FatalMsgs    []Fatal
	SayMsgs      []string
	InfoMsgs     []string
	WarnMsgs     []string
	KeyValueMsgs map[string]string
	Tables       []Table
}

type Table struct {
	Header string
	Rows   [][]string
}

type Fatal struct {
	Msg    string
	Errors []error
}

func (c *Output) Info(msg string, a ...interface{}) {
	c.InfoMsgs = append(c.InfoMsgs, fmt.Sprintf(msg, a...))
}

func (c *Output) Warn(msg string, a ...interface{}) {
	c.WarnMsgs = append(c.WarnMsgs, fmt.Sprintf(msg, a...))
}

func (c *Output) Fatal(err error, msg string, a ...interface{}) {
	c.Fatals([]error{err}, msg, a...)
}

func (c *Output) Fatals(errs []error, msg string, a ...interface{}) {
	c.FatalMsgs = append(c.FatalMsgs, Fatal{Msg: fmt.Sprintf(msg, a...), Errors: errs})
	c.Exited = true
}

func (c *Output) Say(msg string, indent int, a ...interface{}) {
	c.SayMsgs = append(c.SayMsgs, fmt.Sprintf(msg, a...))
}

func (c *Output) Debug(msg string, a ...interface{}) {
	c.DebugMsgs = append(c.DebugMsgs, fmt.Sprintf(msg, a...))
}

func (c *Output) KeyValue(key, value string, indent int, a ...interface{}) {
	if c.KeyValueMsgs == nil {
		c.KeyValueMsgs = make(map[string]string)
	}

	c.KeyValueMsgs[key] = value
}

func (c *Output) Table(header string, rows [][]string) {
	c.Tables = append(c.Tables, Table{Header: header, Rows: rows})
}

func (c *Output) LineBreak() {
}
