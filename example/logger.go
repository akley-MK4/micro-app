package main

import (
	"fmt"
	"github.com/akley-MK4/micro-app/frame"
	"log"
	"os"
)

func newExampleLogger(prefix string) frame.ILogger {
	eLogger := &exampleLogger{
		lg:        log.New(os.Stdout, prefix, log.Llongfile|log.Ldate|log.Ltime),
		calldepth: 2,
	}

	return eLogger
}

var (
	globalLoggerInstance = newExampleLogger("[ExampleApp ]")
)

func getGlobalLoggerInstance() frame.ILogger {
	return globalLoggerInstance
}

type exampleLogger struct {
	lg        *log.Logger
	calldepth int
}

func (t *exampleLogger) SetLevelByDesc(levelDesc string) bool {
	return true
}

func (t *exampleLogger) All(v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintln(v...))
}

func (t *exampleLogger) AllF(format string, v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintf(format, v...))
}

func (t *exampleLogger) Debug(v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintln(v...))
}

func (t *exampleLogger) DebugF(format string, v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintf(format, v...))
}

func (t *exampleLogger) Info(v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintln(v...))
}

func (t *exampleLogger) InfoF(format string, v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintf(format, v...))
}

func (t *exampleLogger) Warning(v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintln(v...))
}

func (t *exampleLogger) WarningF(format string, v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintf(format, v...))
}

func (t *exampleLogger) Error(v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintln(v...))
}

func (t *exampleLogger) ErrorF(format string, v ...interface{}) {
	t.lg.Output(t.calldepth, fmt.Sprintf(format, v...))
}
