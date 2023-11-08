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
	DisableDefaultGC bool `json:"disable_default_gc"`
	EnableForce      bool `json:"enable_force"`
	ForcePolicy      struct {
		IntervalSecondS int `json:"interval_seconds"`
		MemPeak         int `json:"mem_peak"`
	} `json:"force_policy"`
}

type LauncherConfigModel struct {
	AppID          string                   `json:"app_id"`
	LogLevel       string                   `json:"log_level"`
	GCControl      GCControl                `json:"gc_control"`
	ConfigInfoList []*configInfoModel       `json:"configs"`
	SubProcessList []map[string]interface{} `json:"sub_process_list"`
	Components     []componentConfigModel   `json:"components"`
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

	// Initialize and start the configuration handler manager
	getLoggerInst().Info("Initializing configuration manager")
	if err := GetConfigHandlerMgr().initialize(workPath, launcherConf.ConfigInfoList, enabledDevMode); err != nil {
		return fmt.Errorf("unable to initialize configuration manager, %v", err)
	}
	getLoggerInst().Info("Successfully initialized configuration manager")
	getLoggerInst().Info("Starting configuration manager")
	GetConfigHandlerMgr().start()
	getLoggerInst().Info("Successfully started configuration manager")

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
	if err := app.AfterStart(); err != nil {
		return fmt.Errorf("an error occurred when triggering event APP_AfterStart, %v", err)
	}
	getLoggerInst().Info("Successfully triggered event APP_AfterStart")

	if len(launcherConf.SubProcessList) > 0 {
		getLoggerInst().InfoF("Start sub process after %d seconds", waitStartSubProcSec)
		time.Sleep(time.Second * waitStartSubProcSec)
	}
	for _, kwArgs := range launcherConf.SubProcessList {
		_, argsStr, newCmdErr := startSubprocess(os.Args[0], kwArgs)
		if newCmdErr != nil {
			getLoggerInst().WarningF("Failed to start sub process, Args: %s, Err: %v", argsStr, newCmdErr)
			continue
		}
		getLoggerInst().InfoF("Created a sub process, Args: %s", argsStr)
	}

	getLoggerInst().Info("Application is running")
	app.forever()

	getLoggerInst().Info("Preparing to stop application")
	if err := app.StopBefore(); err != nil {
		return fmt.Errorf("an error occurred when triggering event APP_StopBefore, %v", err)
	}
	getLoggerInst().Info("Stopping application")
	if err := app.stop(); err != nil {
		return fmt.Errorf("an error occurred when stopping application, %v", err)
	}

	getLoggerInst().Info("Successfully stopped application")

	return nil
}
