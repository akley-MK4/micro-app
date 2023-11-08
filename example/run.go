package example

import (
	"flag"
	"fmt"
	"github.com/akley-MK4/micro-app/frame"
	"github.com/akley-MK4/rstlog"
	"log"
	"os"
	"runtime"
)

func RunApp() {
	// Parsing Command Line Parameters

	launchConf := flag.String("launcher_cfg", "", "launcher_cfg")
	processType := flag.Int("process_type", 1, "process_type=1")
	numCPU := flag.Int("num_cpu", 0, "num_cpu=1")
	forceMultipleCores := flag.Bool("force_multiple_cores", false, "force_multiple_cores=false")
	logOutPrefix := flag.String("log_out_prefix", "", "log_out_prefix=App1")
	enableDevMode := flag.Bool("enable_dev_mode", false, "enable_dev_mode=false, true")
	logLevelDesc := flag.String("log_level", "INFO", "log_level=INFO")
	flag.Parse()

	// initialize logger instance
	var outPrefix string
	if *logOutPrefix != "" {
		outPrefix = fmt.Sprintf("[%s] ", *logOutPrefix)
	}

	if err := rstlog.InitializeDefaultLogger(*logLevelDesc, outPrefix, rstlog.DefaultCallDepth); err != nil {
		log.Printf("Failed to initialize logger instance, %v\n", err)
		os.Exit(1)
	}

	logger := rstlog.GetDefaultLogger()
	frame.SetFrameLoggerInstance(logger)

	// Check process type
	if *processType != int(frame.MainProcessType) && *processType != int(frame.SubProcessType) {
		logger.ErrorF("Invalid process type %v", *processType)
		os.Exit(1)
	}

	// Get working directory
	workPath, getWdErr := os.Getwd()
	if getWdErr != nil {
		logger.ErrorF("OS.Getwd Failed, %v", getWdErr)
		os.Exit(1)
	}

	// Manually adjusting the number of available cores
	availableNumCPU := runtime.NumCPU()
	if *numCPU > 0 {
		availableNumCPU = *numCPU
		runtime.GOMAXPROCS(*numCPU)
	}
	if *forceMultipleCores {
		runtime.GOMAXPROCS(availableNumCPU)
	}
	logger.InfoF("The current number of logical CPUs available for the process is %d", availableNumCPU)

	// Create and run application
	if err := frame.LaunchDaemonApplication(frame.ProcessType(*processType), workPath, *launchConf,
		nil, nil, *enableDevMode); err != nil {
		logger.ErrorF("Failed to launch application, %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
