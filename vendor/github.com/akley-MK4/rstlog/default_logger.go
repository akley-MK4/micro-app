package rstlog

import (
	"errors"
	"sync/atomic"
)

var (
	defaultLogger  = &Logger{}
	initializeDone uint32
)

func InitializeDefaultLogger(levelDesc string, outPrefix string, callDepth int) error {
	lv, exist := FindLogLevelByDesc(levelDesc)
	if !exist {
		return errors.New("invalid log level")
	}

	loggerInst, newErr := NewLogger(lv, callDepth, outPrefix, "", "")
	if newErr != nil {
		return newErr
	}

	if !atomic.CompareAndSwapUint32(&initializeDone, 0, 1) {
		return errors.New("default logger instance cannot be initialized repeatedly")
	}
	defaultLogger = loggerInst
	return nil
}

func GetDefaultLogger() *Logger {
	return defaultLogger
}

func SetDefaultLogLevel(logLevelDesc string) {
	lv, exist := FindLogLevelByDesc(logLevelDesc)
	if !exist {
		lv = LevelInfo
	}
	defaultLogger.SetLevel(lv)
}

func GetDefaultLogLevel() LogLevel {
	return defaultLogger.GetLogLevel()
}

func All(v ...interface{}) {
	if LevelALL < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelALL, "", v...)
}

func AllF(format string, v ...interface{}) {
	if LevelALL < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelALL, format, v...)
}

func Debug(v ...interface{}) {
	if LevelDebug < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelDebug, "", v...)
}

func DebugF(format string, v ...interface{}) {
	if LevelDebug < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelDebug, format, v...)
}

func Info(v ...interface{}) {
	if LevelInfo < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelInfo, "", v...)
}

func InfoF(format string, v ...interface{}) {
	if LevelInfo < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelInfo, format, v...)
}

func Warning(v ...interface{}) {
	if LevelWarning < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelWarning, "", v...)
}

func WarningF(format string, v ...interface{}) {
	if LevelWarning < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelWarning, format, v...)
}

func Error(v ...interface{}) {
	if LevelError < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelError, "", v...)
}

func ErrorF(format string, v ...interface{}) {
	if LevelError < defaultLogger.level {
		return
	}
	_ = defaultLogger.Output(LevelError, format, v...)
}
