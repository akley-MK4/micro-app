# rstlog
## Introduction
This is a simple log library that supports log levels such as ALL, DEBUG, INFO, WARNING, ERROR, FATAL, and outputs logs to stdout.
## example
```go

logLevel := rstlog.LevelDebug
callDepth := 3
outPrefix := "TestService"

logInst, newErr := rstlog.NewLogger(logLevel, callDepth, outPrefix, "", "")
if newErr != nil {
	log.Printf("Failed to create logger instance, %v\n", newErr)
	os.Exit(1)
}

logInst.Debug("test output")
logInst.DebugF("test output %v", 10)

```