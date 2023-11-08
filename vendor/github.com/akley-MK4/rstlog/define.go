package rstlog

type LogLevel uint8

const (
	DefaultCallDepth = 3
)

const (
	LevelInvalid LogLevel = iota
	LevelALL
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal

	LevelMaxInvalid
)

var (
	levelDescMap = map[LogLevel]string{
		LevelALL: "[ALL]", LevelDebug: "[DEBUG]", LevelInfo: "[INFO]",
		LevelWarning: "[WARNING]", LevelError: "[ERROR]", LevelFatal: "[FATAL]",
	}
)
