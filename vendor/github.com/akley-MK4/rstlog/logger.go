package rstlog

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func NewLogger(level LogLevel, callDepth int, outPrefix, dirPath, fileName string) (*Logger, error) {
	if level <= LevelInvalid || level >= LevelMaxInvalid {
		return nil, errors.New("invalid log level")
	}

	inst := &Logger{
		level:     level,
		callDepth: callDepth,
		outPrefix: outPrefix,
	}

	if err := inst.bindLogFile(dirPath, fileName); err != nil {
		return nil, err
	}

	inst.logger = log.New(inst.logFile, inst.outPrefix, log.Llongfile|log.Ldate|log.Ltime)

	return inst, nil
}

type Logger struct {
	logger    *log.Logger
	level     LogLevel
	callDepth int
	dirPath   string
	fileName  string
	logFile   *os.File
	isStdOut  bool
	outPrefix string
}

func (t *Logger) SetCallDepth(depth int) {
	t.callDepth = depth
}

func (t *Logger) SetLevel(level LogLevel) {
	t.level = level
}

func (t *Logger) SetLevelByDesc(levelDesc string) bool {
	lv, ok := FindLogLevelByDesc(levelDesc)
	if !ok {
		return false
	}

	t.level = lv
	return true
}

func (t *Logger) SetOutPrefix(outPrefix string) {
	t.outPrefix = outPrefix
	if t.logger == nil {
		return
	}
	t.logger.SetPrefix(outPrefix)
}

func (t *Logger) GetLogLevel() LogLevel {
	return t.level
}

func (t *Logger) bindLogFile(dirPath, fileName string) error {
	if dirPath == "" || fileName == "" {
		t.isStdOut = true
		t.logFile = os.Stdout
		return nil
	}

	if !t.checkPathExist(dirPath) {
		return fmt.Errorf("%s path dose not exist", dirPath)
	}

	filePath := path.Join(dirPath, fmt.Sprintf("%s.log", fileName))
	f, openErr := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if openErr != nil {
		return openErr
	}

	t.logFile = f
	return nil
}

func (t *Logger) checkPathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func (t *Logger) Release() error {
	if t.isStdOut || t.logFile == nil {
		return nil
	}

	return t.logFile.Close()
}

func (t *Logger) Output(level LogLevel, format string, v ...interface{}) error {
	if t.level == LevelInvalid || level < t.level {
		return nil
	}

	if t.logger == nil {
		return nil
	}

	var content string
	if format != "" {
		content = fmt.Sprintf(format, v...)
	} else {
		content = fmt.Sprintln(v...)
	}

	return t.logger.Output(t.callDepth, strings.Join([]string{levelDescMap[level], content}, " "))
}

func (t *Logger) All(v ...interface{}) {
	_ = t.Output(LevelALL, "", v...)
}

func (t *Logger) AllF(format string, v ...interface{}) {
	_ = t.Output(LevelALL, format, v...)
}

func (t *Logger) Debug(v ...interface{}) {
	_ = t.Output(LevelDebug, "", v...)
}

func (t *Logger) DebugF(format string, v ...interface{}) {
	_ = t.Output(LevelDebug, format, v...)
}

func (t *Logger) Info(v ...interface{}) {
	_ = t.Output(LevelInfo, "", v...)
}

func (t *Logger) InfoF(format string, v ...interface{}) {
	_ = t.Output(LevelInfo, format, v...)
}

func (t *Logger) Warning(v ...interface{}) {
	_ = t.Output(LevelWarning, "", v...)
}

func (t *Logger) WarningF(format string, v ...interface{}) {
	_ = t.Output(LevelWarning, format, v...)
}

func (t *Logger) Error(v ...interface{}) {
	_ = t.Output(LevelError, "", v...)
}

func (t *Logger) ErrorF(format string, v ...interface{}) {
	_ = t.Output(LevelError, format, v...)
}

func FindLogLevelByDesc(levelDesc string) (LogLevel, bool) {
	levelDesc = fmt.Sprintf("[%s]", levelDesc)
	for lv, desc := range levelDescMap {
		if levelDesc == desc {
			return lv, true
		}
	}

	return LevelInvalid, false
}
