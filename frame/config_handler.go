package frame

import (
	"crypto/md5"
	"fmt"
	"github.com/akley-MK4/go-tools-box/ctime"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"path"
	"time"
)

const (
	defaultWaitConfInitDoneSec = 2
	intervalRetryAddWatchSec   = 2
)

type ConfigCallbackType = uint8

const (
	ConfigCallbackTypeCreate ConfigCallbackType = iota
	ConfigCallbackTypeUpdate
	ConfigCallbackTypeRemove
)

type newConfigFunc func() IConfig

type ConfigRegInfo struct {
	Key                   string
	Suffix                string
	NewFunc               newConfigFunc
	MustLoad              bool
	EnableWatchLog        bool
	RetryWatchIntervalSec uint64
}

var (
	configHandlerMgr = &ConfigHandlerMgr{
		handlerMap: make(map[string]*ConfigHandler),
	}
	configRegInfoMap = make(map[string]*ConfigRegInfo)
)

func GetConfigHandlerMgr() *ConfigHandlerMgr {
	return configHandlerMgr
}

func RegisterConfigInfo(info ConfigRegInfo) {
	configRegInfoMap[info.Key] = &info
}

type ConfigCallback func()

func RegisterConfigCallback(cbType ConfigCallbackType, conf IConfig, f ConfigCallback) bool {
	if cbType < ConfigCallbackTypeCreate || cbType > ConfigCallbackTypeRemove {
		return false
	}

	for _, handler := range configHandlerMgr.handlerMap {
		if handler.config != conf {
			continue
		}

		var cbList *[]ConfigCallback

		switch cbType {
		case ConfigCallbackTypeCreate:
			cbList = &handler.createTypeCallbacks
			break
		case ConfigCallbackTypeUpdate:
			cbList = &handler.updateTypeCallbacks
			break
		case ConfigCallbackTypeRemove:
			cbList = &handler.removeTypeCallbacks
			break
		default:
			return false
		}

		*cbList = append(*cbList, f)
		return true
	}

	return false
}

func GetConfigTemplatePath(workPath string) string {
	return path.Join(workPath, "configs", "template")
}

type IConfig interface {
	EncodeConfig(data []byte) error
	OnUpdate()
	GetModelData() ([]byte, error)
}

type ConfigHandlerMgr struct {
	handlerMap map[string]*ConfigHandler
}

func (t *ConfigHandlerMgr) initialize(workPath string, configInfoList []*configInfoModel, enabledDevMode bool) error {
	t.handlerMap = make(map[string]*ConfigHandler)

	for _, info := range configInfoList {
		regInfo := configRegInfoMap[info.Key]
		if regInfo == nil {
			getLoggerInst().WarningF("Configuration key %v not registered", info.Key)
			continue
		}

		if enabledDevMode {
			fileName := info.Key
			if regInfo.Suffix != "" {
				fileName = fmt.Sprintf("%s.%s", info.Key, regInfo.Suffix)
			}
			info.Path = path.Join(GetConfigTemplatePath(workPath), fileName)
		}
		if info.EnableWatchLog {
			regInfo.EnableWatchLog = info.EnableWatchLog
		}
		if info.RetryWatchIntervalSec > 0 {
			regInfo.RetryWatchIntervalSec = info.RetryWatchIntervalSec
		}

		handler := &ConfigHandler{}
		if err := handler.initialize(info.Key, info.Path, regInfo); err != nil {
			return fmt.Errorf("unable to initialize ConfigHandler, Key: %v, Err: %v", info.Key, err)
		}

		t.handlerMap[handler.GetKey()] = handler
	}

	return nil
}

func (t *ConfigHandlerMgr) start() {
	for _, handler := range t.handlerMap {
		if err := handler.start(); err != nil {
			getLoggerInst().WarningF("failed to start ConfigHandler, key: %v, Err: %v", handler.GetKey(), err)
		}
	}
}

func (t *ConfigHandlerMgr) stop() {
	for _, handler := range t.handlerMap {
		if err := handler.stop(); err != nil {
			getLoggerInst().WarningF("failed to stop ConfigHandler, key: %v, Err: %v", handler.GetKey(), err)
		}
	}
}

func (t *ConfigHandlerMgr) GetConfigHandlerListInfo() (retList []ConfigHandlerInfo) {
	for _, handler := range t.handlerMap {
		retList = append(retList, handler.GetInfo())
	}

	return
}

type ConfigHandler struct {
	version               int
	updateTimestamp       int64
	key                   string
	path                  string
	dir                   string
	fileName              string
	hashVal               [md5.Size]byte
	watched               bool
	enableWatchLog        bool
	retryWatchIntervalSec uint64

	watcher             *fsnotify.Watcher
	config              IConfig
	updateTypeCallbacks []ConfigCallback
	createTypeCallbacks []ConfigCallback
	removeTypeCallbacks []ConfigCallback
}

func (t *ConfigHandler) initialize(key, filePath string, regInfo *ConfigRegInfo) error {
	t.key = key
	t.dir, t.fileName = path.Split(filePath)
	t.path = path.Join(t.dir, t.fileName)
	t.config = regInfo.NewFunc()

	w, newWatcherErr := fsnotify.NewWatcher()
	if newWatcherErr != nil {
		return fmt.Errorf("failed to create watcher, %v", newWatcherErr)
	}
	t.watcher = w
	t.enableWatchLog = regInfo.EnableWatchLog
	t.retryWatchIntervalSec = regInfo.RetryWatchIntervalSec
	if t.retryWatchIntervalSec <= 0 {
		t.retryWatchIntervalSec = defaultWaitConfInitDoneSec
	}

	if err := w.Add(t.dir); err != nil {
		getLoggerInst().WarningF("Failed to add watch path %s, Key: %v, Err: %v", t.dir, t.key, err)
		if regInfo.MustLoad {
			return err
		}

		go func() {
			t.intervalRetryWatchPath(t.dir)
		}()
		return nil
	}
	t.watched = true

	loadErr := t.loadFiled()
	if loadErr != nil {
		getLoggerInst().WarningF("Failed to load configuration file from path %s, Key: %v, Err: %v", t.path, t.key, loadErr)
	}
	if regInfo.MustLoad && loadErr != nil {
		return loadErr
	}

	getLoggerInst().InfoF("Successfully initialized ConfigHandler %v, Dir: %s, Path: %s", t.key, t.dir, t.path)
	return nil
}

func (t *ConfigHandler) start() error {
	go t.loopWatch()
	return nil
}

func (t *ConfigHandler) stop() error {
	return t.watcher.Close()
}

func (t *ConfigHandler) loadFiled() error {
	data, readErr := ioutil.ReadFile(t.path)
	if readErr != nil {
		return readErr
	}
	if len(data) <= 0 {
		return nil
	}

	hashVal := md5.Sum(data)
	if hashVal == t.hashVal {
		//logger.DebugFmt("The config content has not changed, and the config will not be updated, "+
		//	"Path: %s",
		//	t.path)
		return nil
	}

	t.hashVal = hashVal
	getLoggerInst().InfoF("Read the configuration file from path %s, data: \n%s", t.path, string(data))
	if err := t.config.EncodeConfig(data); err != nil {
		return err
	}
	t.version += 1
	t.updateTimestamp = ctime.CurrentTimestamp()

	updateCallbacks := t.updateTypeCallbacks
	for _, f := range updateCallbacks {
		f()
	}
	getLoggerInst().InfoF("Updated the configuration model of ConfigHandler %v, Version: %v", t.key, t.version)
	return nil
}

func (t *ConfigHandler) loopWatch() {
	if !t.watched {
		getLoggerInst().WarningF("Configuration %v dose not watch successfully, start timing check operation", t.key)
		for {
			if t.enableWatchLog {
				getLoggerInst().DebugF("Configuration %v dose not watch successfully. "+
					"Wait for %d seconds before check and watch the config", t.key, t.retryWatchIntervalSec)
			}
			time.Sleep(time.Second * time.Duration(t.retryWatchIntervalSec))
			if t.watched {
				break
			}
		}
		getLoggerInst().InfoF("Configuration %v exits the timed check operation", t.key)

		if err := t.loadFiled(); err != nil {
			//return err
			getLoggerInst().WarningF("Failed to load configuration file, Key: %v, Err: %v", t.key, err)
		}
	}

	for {
		if t.watch() {
			break
		}
	}

	getLoggerInst().Warning("LoopWatch break, Path: %v", t.dir)
}

func (t *ConfigHandler) watch() bool {
	select {
	case err, ok := <-t.watcher.Errors:
		if !ok {
			getLoggerInst().Warning("watcher.Errors not ok")
			return true
		}
		getLoggerInst().WarningF("ConfigHandler %v has failed to watch, Err: %v", t.key, err)
	case e, ok := <-t.watcher.Events:
		if !ok {
			getLoggerInst().Warning("watcher.Events not ok")
			return true
		}

		//logger.WarningF("The config file of path %s has changed, op: %v", e.Name, e.Op)
		needLoad := true
		var cbs []ConfigCallback
		switch e.Op {
		case fsnotify.Create:
			cbs = t.createTypeCallbacks
		case fsnotify.Remove:
			cbs = t.removeTypeCallbacks
			needLoad = false
		case fsnotify.Rename:
			return false
		}
		if len(cbs) > 0 || !needLoad {
			for _, cb := range cbs {
				cb()
			}
			return false
		}

		//if e.Op&fsnotify.Write != fsnotify.Write && e.Op&fsnotify.Chmod != fsnotify.Chmod {
		//	break
		//}

		//idx := strings.LastIndex(e.Name, t.fileName)
		//if (idx + len(t.fileName)) != len(e.Name) {
		//	break
		//}

		if err := t.loadFiled(); err != nil {
			getLoggerInst().WarningF("Failed to load configuration file from path %s, %v", e.Name, err)
		}
	}

	return false
}

func (t *ConfigHandler) GetVersion() int {
	return t.version
}

func (t *ConfigHandler) GetKey() string {
	return t.key
}

func (t *ConfigHandler) GetPath() string {
	return t.path
}

func (t *ConfigHandler) intervalRetryWatchPath(dirPath string) {
	var retryTotal int
	for {
		time.Sleep(time.Second * time.Duration(t.retryWatchIntervalSec))
		retryTotal += 1
		if watchErr := t.watcher.Add(dirPath); watchErr != nil {
			if t.enableWatchLog {
				getLoggerInst().WarningF("Failed to watch path %s, Key: %s, RetryTotal: %d, Err: %v",
					dirPath, t.key, retryTotal, watchErr)
			}
			continue
		}

		t.watched = true
		getLoggerInst().InfoF("Watched path %s, Key: %s, RetryTotal: %d", dirPath, t.key, retryTotal)
		return
	}
}

type ConfigHandlerInfo struct {
	Version         int
	UpdateTimestamp int64
	Key             string
	Path            string
	Dir             string
	FileName        string
	//HashVal         string
	Watched   bool
	ModelData string
}

func (t *ConfigHandler) GetInfo() (retInfo ConfigHandlerInfo) {
	hashVal := make([]byte, 0, md5.Size)
	for _, v := range t.hashVal {
		hashVal = append(hashVal, v)
	}

	retInfo.Version = t.version
	retInfo.UpdateTimestamp = t.updateTimestamp
	retInfo.Key = t.key
	retInfo.Path = t.path
	retInfo.Dir = t.dir
	retInfo.FileName = t.fileName
	//retInfo.HashVal = string(hashVal)
	retInfo.Watched = t.watched

	if t.config == nil {
		return
	}

	modelData, getModelDataErr := t.config.GetModelData()
	if getModelDataErr != nil {
		getLoggerInst().WarningF("Failed to get all configurations info, %v", getModelDataErr)
		return
	}

	retInfo.ModelData = string(modelData)
	return
}
