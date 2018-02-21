package mock

import (
	"fmt"
	"sync"
)

type Output struct {
	DebugMsgs    []string
	Exited       bool
	FatalMsgs    []Fatal
	InfoMsgs     []string
	KeyValueMsgs map[string]string
	SayMsgs      []string
	Tables       []Table
	WarnMsgs     []string
	lock         sync.Mutex
}

type Table struct {
	Header string
	Rows   [][]string
}

type Fatal struct {
	Errors []error
	Msg    string
}

func (o *Output) Info(msg string, a ...interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.InfoMsgs = append(o.InfoMsgs, fmt.Sprintf(msg, a...))
}

func (o *Output) Warn(msg string, a ...interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.WarnMsgs = append(o.WarnMsgs, fmt.Sprintf(msg, a...))
}

func (o *Output) Fatal(err error, msg string, a ...interface{}) {
	o.Fatals([]error{err}, msg, a...)
}

func (o *Output) Fatals(errs []error, msg string, a ...interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.FatalMsgs = append(o.FatalMsgs, Fatal{Msg: fmt.Sprintf(msg, a...), Errors: errs})
	o.Exited = true
}

func (o *Output) Say(msg string, indent int, a ...interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.SayMsgs = append(o.SayMsgs, fmt.Sprintf(msg, a...))
}

func (o *Output) Debug(msg string, a ...interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.DebugMsgs = append(o.DebugMsgs, fmt.Sprintf(msg, a...))
}

func (o *Output) KeyValue(key, value string, indent int, a ...interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.KeyValueMsgs == nil {
		o.KeyValueMsgs = make(map[string]string)
	}

	o.KeyValueMsgs[key] = value
}

func (o *Output) Table(header string, rows [][]string) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.Tables = append(o.Tables, Table{Header: header, Rows: rows})
}

func (o *Output) LineBreak() {
}
