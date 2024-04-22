package frame

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

const (
	waitStartSubProcSec = 5
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
	Percent                    int    `json:"percent"`
	DisableDefaultGC           bool   `json:"disable_default_gc"`
	MemoryUsageLimitBytes      int64  `json:"memory_usage_limit_bytes"`
	MemoryUsageLimitPercentage string `json:"memory_usage_limit_percentage"`
	EnableForce                bool   `json:"enable_force"`
	ForcePolicy                struct {
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

func LaunchDaemonApplication(processType ProcessType, workPath string, launchConf string, newApp NewApplication,
	appArgs []interface{}, enabledDevMode bool) error {
	getLoggerInst().InfoF("Execution parameters: %v", strings.Join(os.Args, " "))

	// Load launcher config
	if enabledDevMode {
		launchConf = path.Join(GetConfigTemplatePath(workPath), launchConf)
	}

	getLoggerInst().InfoF("Loading startup configuration from path %s", launchConf)
	fileData, readFileErr := ioutil.ReadFile(launchConf)
	if readFileErr != nil {
		return fmt.Errorf("unable to load startup configuration, %v", readFileErr)
	}
	getLoggerInst().InfoF("Startup Configuration Data: \n%s", fileData)
	launcherConf := &LauncherConfigModel{}
	if err := json.Unmarshal(fileData, launcherConf); err != nil {
		return fmt.Errorf("unable to unmarshal startup configuration, %v", err)
	}
	getLoggerInst().Info("Successfully loaded startup configuration")

	// Set log level
	getLoggerInst().SetLevelByDesc(launcherConf.LogLevel)

	// Set GC
	setGCPolicy(launcherConf.GCControl)

	// Initialize and start the configuration watcher manager
	getLoggerInst().Info("Initializing configuration watcher manager")
	if err := GetConfigWatcherMgr().initialize(workPath, launcherConf.ConfigInfoList, enabledDevMode); err != nil {
		return fmt.Errorf("unable to initialize configuration watcher manager, %v", err)
	}
	getLoggerInst().Info("Successfully initialized configuration watcher manager")
	getLoggerInst().Info("Starting configuration watcher manager")
	GetConfigWatcherMgr().start()
	getLoggerInst().Info("Successfully started configuration watcher manager")

	getLoggerInst().InfoF("New application with id (%v)", launcherConf.AppID)
	var app IApplication
	if newApp != nil {
		app = newApp()
	} else {
		app = &BaseApplication{}
	}

	// Initialize APP
	getLoggerInst().Info("Initializing application")
	if err := app.baseInitialize(ApplicationID(launcherConf.AppID), processType); err != nil {
		return fmt.Errorf("unable to initialize base application, %v", err)
	}
	if err := app.Initialize(appArgs...); err != nil {
		return fmt.Errorf("unable to initialize application, %v", err)
	}
	setAppProxy(app)
	getLoggerInst().Info("Successfully initialized application")

	// Import component list
	getLoggerInst().Info("Importing components")
	if err := app.importComponents(launcherConf.Components); err != nil {
		return fmt.Errorf("not all components imported, %v", err)
	}
	getLoggerInst().Info("All components imported successfully")

	getLoggerInst().Info("Starting application")
	if err := app.start(); err != nil {
		return fmt.Errorf("unable to start application, %v", err)
	}
	getLoggerInst().Info("Successfully started application")

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

	setInitialMemorySnapshot()
	memSnapshot := GetInitialMemorySnapshot()
	memSnapshotData, marshalMemSnapshotErr := json.Marshal(memSnapshot)
	if marshalMemSnapshotErr != nil {
		getLoggerInst().WarningF("Failed to marshal current memory snapshot, %v", marshalMemSnapshotErr)
	} else {
		getLoggerInst().InfoF("Current memory snapshot: %v", string(memSnapshotData))
	}

	// process pid file
	pidFileDirPath := launcherConf.PidFileDirPath
	if pidFileDirPath == "" {
		pidFileDirPath = workPath
	}
	var pidFilePath string
	if launcherConf.AppID != "" {
		pidFilePath = path.Join(pidFileDirPath, fmt.Sprintf("%s.pid", launcherConf.AppID))
	}
	if pidFilePath != "" {
		checkPidFileExist, checkPidFileErr := checkProcessIdFileExist(pidFilePath)
		if checkPidFileErr != nil {
			getLoggerInst().WarningF("Failed to check the process id file, Path: %s, Err: %v", pidFilePath, checkPidFileErr)
		}
		if checkPidFileExist {
			getLoggerInst().Warning("There are currently identical pid files, please check for any conflicts")
		}

		pidNum, pidFileErr := updateProcessIdFile(pidFilePath)
		if pidFileErr != nil {
			getLoggerInst().WarningF("Failed to update process id file, Path: %v, Err: %v", pidFilePath, pidFileErr)
		} else {
			getLoggerInst().InfoF("The current process id is %d, and the file path is %s", pidNum, pidFilePath)
		}
	} else {
		getLoggerInst().Warning("The process id file not created, unable to concatenate complete path")
	}

	getLoggerInst().Info("Application is running")
	app.forever()

	getLoggerInst().Info("Stopping application")
	if err := app.stop(); err != nil {
		return fmt.Errorf("an error occurred when stopping application, %v", err)
	}

	if err := deleteProcessIdFile(pidFilePath); err != nil {
		getLoggerInst().WarningF("Failed to delete the process id file, %v", err)
	} else {
		getLoggerInst().Info("Successfully deleted the process id file")
	}

	getLoggerInst().Info("Successfully stopped application")

	return nil
}
