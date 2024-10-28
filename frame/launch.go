package frame

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	ossignal "github.com/akley-MK4/go-tools-box/signal"
)

const (
	waitStartSubProcSec = 5
	defaultSignChanSize = 1
)

type componentConfigModel struct {
	ComponentType string                 `json:"component_type"`
	Disable       bool                   `json:"disable"`
	Kw            map[string]interface{} `json:"kw"`
}

type configInfoModel struct {
	Key                   string `json:"key"`
	Path                  string `json:"path"`
	EnableWatchLog        bool   `json:"enableWatchLog"`
	RetryWatchIntervalSec uint64 `json:"retryWatchIntervalSec"`
}

type GCControl struct {
	Percent               int   `json:"percent"`
	DisableDefaultGC      bool  `json:"disable_default_gc"`
	MemoryUsageLimitBytes int64 `json:"memory_usage_limit_bytes"`
	EnableForce           bool  `json:"enable_force"`
	ForcePolicy           struct {
		IntervalSecondS int `json:"interval_seconds"`
		MemPeak         int `json:"mem_peak"`
	} `json:"force_policy"`
}

type SubProcessList struct {
	Enable   bool                     `json:"enable"`
	Commands []map[string]interface{} `json:"commands"`
}

type LauncherConfigModel struct {
	AppID          string                 `json:"app_id"`
	PidFileDirPath string                 `json:"pid_file_dir_path"`
	LogLevel       string                 `json:"log_level"`
	GCControl      GCControl              `json:"gc_control"`
	ConfigInfoList []*configInfoModel     `json:"configs"`
	SubProcessList SubProcessList         `json:"sub_process_list"`
	Components     []componentConfigModel `json:"components"`
}

func LaunchDaemonApplication(processType ProcessType, workPath string, launchConf string, appArgs []interface{}, enabledDevMode bool) error {
	getLoggerInst().InfoF("Execution parameters: %v", strings.Join(os.Args, " "))

	// Load launcher config
	if enabledDevMode {
		launchConf = path.Join(GetConfigTemplatePath(workPath), launchConf)
	}

	fileData, readFileErr := os.ReadFile(launchConf)
	if readFileErr != nil {
		return fmt.Errorf("unable to load the startup configuration, %v", readFileErr)
	}
	launcherConf := &LauncherConfigModel{}
	if err := json.Unmarshal(fileData, launcherConf); err != nil {
		return fmt.Errorf("unable to unmarshal the startup configuration, %v", err)
	}
	getLoggerInst().InfoF("Loaded the startup configuration from path %s", launchConf)
	fmt.Println(string(fileData))

	// check and update process info
	setCurrentProcessType(processType)

	if launcherConf.AppID == "" {
		return fmt.Errorf("invalid app id")
	}
	pidFileDirPath := launcherConf.PidFileDirPath
	if pidFileDirPath == "" {
		pidFileDirPath = workPath
	}
	pid, pidFilePath, errPid := checkAndCreateProcessId(pidFileDirPath, launcherConf.AppID)
	if errPid != nil {
		return fmt.Errorf("checkAndCreateProcessId failed, %v", errPid)
	}
	getLoggerInst().InfoF("The current process id is %d, and the file path is %s", pid, pidFilePath)

	// Set signal handler
	signalHandler := &ossignal.Handler{}
	if err := signalHandler.InitSignalHandler(defaultSignChanSize); err != nil {
		return err
	}
	for _, sig := range []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT} {
		signalHandler.RegisterSignal(sig, func() {
			signalHandler.CloseSignalHandler()
		})
	}

	// log level
	getLoggerInst().SetLevelByDesc(launcherConf.LogLevel)

	// Set memory garbage collection policy
	setGCPolicy(launcherConf.GCControl)

	// Initialize the event message manager
	if err := initializeEventMessageMgr(); err != nil {
		return fmt.Errorf("unable to initialize EventMessageMgr, %v", err)
	}

	// Initialize and start the configuration watcher manager
	if err := GetConfigWatcherMgr().initialize(workPath, launcherConf.ConfigInfoList, enabledDevMode); err != nil {
		return fmt.Errorf("unable to initialize configuration watcher manager, %v", err)
	}
	GetConfigWatcherMgr().start()

	getLoggerInst().Info("Initialized the application")

	// Initialize and start all components
	var components []IComponent
	for componentIdx, cfg := range launcherConf.Components {
		if component, err := createAndInitializeComponent(componentIdx, cfg); err != nil {
			return fmt.Errorf("unable to create and initialize component type %v, ComponentIndex: %v, Error: %v", cfg.ComponentType, componentIdx, err)
		} else {
			components = append(components, component)
		}
	}
	getLoggerInst().Info("Successfully created and initialized all components")

	for _, component := range components {
		if err := component.Start(); err != nil {
			return fmt.Errorf("unable to start component %v, %v", component.GetID(), err)
		}
		getLoggerInst().InfoF("The component %v has started", component.GetID())
	}
	getLoggerInst().Info("Successfully started all components")

	// Check and create all child processes
	if launcherConf.SubProcessList.Enable {
		getLoggerInst().InfoF("Start sub process after %d seconds", waitStartSubProcSec)
		time.Sleep(time.Second * waitStartSubProcSec)

		for _, kwArgs := range launcherConf.SubProcessList.Commands {
			_, argsStr, newCmdErr := startSubprocess(os.Args[0], kwArgs)
			if newCmdErr != nil {
				getLoggerInst().WarningF("Failed to start sub process, Args: %s, Err: %v", argsStr, newCmdErr)
				continue
			}
			getLoggerInst().InfoF("Created a sub process, Args: %s", argsStr)
		}
	}

	getLoggerInst().Info("The application has been successfully launched and is currently running")

	// Record initial memory information snapshot
	fmt.Println("Current memory snapshot: \n", PrintCurrentMemorySnapshot())

	PublishEventMessage(EventAPPStarted)

	signalHandler.ListenSignal()
	// Stop process
	getLoggerInst().Info("Stopping the application")

	for _, component := range components {
		if err := component.Stop(); err != nil {
			getLoggerInst().WarningF("Failed to stop component %v, %v", component.GetID(), err)
			continue
		}
		getLoggerInst().InfoF("The component %v has stopped", component.GetID())
	}
	getLoggerInst().Info("Stopped all components")

	if err := deleteProcessIdFile(pidFilePath); err != nil {
		getLoggerInst().WarningF("Failed to delete the process id file, %v", err)
	} else {
		getLoggerInst().Info("Deleted the process id file")
	}

	getLoggerInst().Info("Stopped the application")

	return nil
}
