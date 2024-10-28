package frame

import (
	"crypto/md5"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/akley-MK4/go-tools-box/ctime"
	"github.com/fsnotify/fsnotify"
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

type NewConfigHandlerFunc func() IConfigHandler

type ConfigRegInfo struct {
	Key                   string
	Suffix                string
	NewConfigHandlerFunc  NewConfigHandlerFunc
	MustLoad              bool
	EnableWatchLog        bool
	RetryWatchIntervalSec uint64
}

var (
	configWatcherMgr = &ConfigWatcherMgr{
		watcherMap: make(map[string]*ConfigWatcher),
	}
	configRegInfoMap = make(map[string]*ConfigRegInfo)
)

func GetConfigWatcherMgr() *ConfigWatcherMgr {
	return configWatcherMgr
}

func RegisterConfigInfo(info ConfigRegInfo) {
	configRegInfoMap[info.Key] = &info
}

type ConfigCallback func()

func RegisterConfigCallback(cbType ConfigCallbackType, confHandler IConfigHandler, f ConfigCallback) bool {
	if cbType < ConfigCallbackTypeCreate || cbType > ConfigCallbackTypeRemove {
		return false
	}

	for _, watcher := range configWatcherMgr.watcherMap {
		if watcher.confHandler != confHandler {
			continue
		}

		var cbList *[]ConfigCallback

		switch cbType {
		case ConfigCallbackTypeCreate:
			cbList = &watcher.createTypeCallbacks
			break
		case ConfigCallbackTypeUpdate:
			cbList = &watcher.updateTypeCallbacks
			break
		case ConfigCallbackTypeRemove:
			cbList = &watcher.removeTypeCallbacks
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

type IConfigHandler interface {
	EncodeConfig(data []byte) error
	OnUpdate()
	GetConfigData() ([]byte, error)
}

type ConfigWatcherMgr struct {
	watcherMap map[string]*ConfigWatcher
}

func (t *ConfigWatcherMgr) initialize(workPath string, configInfoList []*configInfoModel, enabledDevMode bool) error {
	t.watcherMap = make(map[string]*ConfigWatcher)

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

		watcher := &ConfigWatcher{}
		if err := watcher.initialize(info.Key, info.Path, regInfo); err != nil {
			return fmt.Errorf("unable to initialize ConfigWatcher %v, Err: %v", info.Key, err)
		}

		t.watcherMap[watcher.GetKey()] = watcher
	}

	return nil
}

func (t *ConfigWatcherMgr) start() {
	for _, watcher := range t.watcherMap {
		if err := watcher.start(); err != nil {
			getLoggerInst().WarningF("failed to start ConfigWatcher %v, Err: %v", watcher.GetKey(), err)
		}
	}
}

func (t *ConfigWatcherMgr) stop() {
	for _, watcher := range t.watcherMap {
		if err := watcher.stop(); err != nil {
			getLoggerInst().WarningF("failed to stop ConfigWatcher %v, Err: %v", watcher.GetKey(), err)
		}
	}
}

func (t *ConfigWatcherMgr) GetConfigWatcherListInfo() (retList []ConfigWatcherInfo) {
	for _, watcher := range t.watcherMap {
		retList = append(retList, watcher.GetInfo())
	}

	return
}

type ConfigWatcher struct {
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
	confHandler         IConfigHandler
	updateTypeCallbacks []ConfigCallback
	createTypeCallbacks []ConfigCallback
	removeTypeCallbacks []ConfigCallback
}

func (t *ConfigWatcher) initialize(key, filePath string, regInfo *ConfigRegInfo) error {
	t.key = key
	t.dir, t.fileName = path.Split(filePath)
	t.path = path.Join(t.dir, t.fileName)
	t.confHandler = regInfo.NewConfigHandlerFunc()

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
		if t.enableWatchLog {
			getLoggerInst().WarningF("Unable to watch path %v for configuration %v, %v", t.dir, t.key, err)
		}
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
	if loadErr != nil && t.enableWatchLog {
		getLoggerInst().WarningF("Failed to load configuration %v from path %s, %v", t.key, t.path, loadErr)
	}
	if regInfo.MustLoad && loadErr != nil {
		return loadErr
	}

	return nil
}

func (t *ConfigWatcher) start() error {
	go t.loopWatch()
	return nil
}

func (t *ConfigWatcher) stop() error {
	return t.watcher.Close()
}

func (t *ConfigWatcher) loadFiled() error {
	data, readErr := os.ReadFile(t.path)
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
	if err := t.confHandler.EncodeConfig(data); err != nil {
		return err
	}
	t.version += 1
	t.updateTimestamp = ctime.CurrentTimestamp()
	getLoggerInst().InfoF("Updated the configuration %v from path %v, Version: %v", t.key, t.path, t.version)
	if t.enableWatchLog {
		getLoggerInst().InfoF("The content of configuration %v in version %v is as follows", t.key, t.version)
		fmt.Println(string(data))
	}

	updateCallbacks := t.updateTypeCallbacks
	for _, f := range updateCallbacks {
		f()
	}

	return nil
}

func (t *ConfigWatcher) loopWatch() {
	if !t.watched {
		getLoggerInst().InfoF("The configuration %v dose not watch successfully, start timing check operation", t.key)

		for {
			if t.enableWatchLog {
				getLoggerInst().DebugF("The configuration %v dose not watch successfully. "+
					"Wait for %d seconds before check and watch the config", t.key, t.retryWatchIntervalSec)
			}
			time.Sleep(time.Second * time.Duration(t.retryWatchIntervalSec))
			if t.watched {
				break
			}
		}
		if t.enableWatchLog {
			getLoggerInst().InfoF("The configuration %v exits the timed check operation", t.key)
		}

		if err := t.loadFiled(); err != nil {
			//return err
			getLoggerInst().WarningF("Failed to load configuration %v, %v", t.key, err)
		}
	}

	for {
		if t.watch() {
			break
		}
	}

	getLoggerInst().InfoF("ConfigWatcher.loopWatch quit, Key: %v, Path: %v", t.key, t.path)
}

func (t *ConfigWatcher) watch() bool {
	select {
	case err, ok := <-t.watcher.Errors:
		if !ok {
			getLoggerInst().Warning("watcher.Errors not ok")
			return true
		}
		getLoggerInst().WarningF("ConfigWatcher %v has failed to watch, Err: %v", t.key, err)
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
			getLoggerInst().WarningF("Failed to load configuration %v from path %s, %v", t.key, t.path, err)
		}
	}

	return false
}

func (t *ConfigWatcher) GetVersion() int {
	return t.version
}

func (t *ConfigWatcher) GetKey() string {
	return t.key
}

func (t *ConfigWatcher) GetPath() string {
	return t.path
}

func (t *ConfigWatcher) intervalRetryWatchPath(dirPath string) {
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

type ConfigWatcherInfo struct {
	Version         int
	UpdateTimestamp int64
	Key             string
	Path            string
	Dir             string
	FileName        string
	//HashVal         string
	Watched    bool
	ConfigData string
}

func (t *ConfigWatcher) GetInfo() (retInfo ConfigWatcherInfo) {
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

	if t.confHandler == nil {
		return
	}

	cfgData, getCfgDataErr := t.confHandler.GetConfigData()
	if getCfgDataErr != nil {
		getLoggerInst().WarningF("Failed to get the data of the configuration %v, %v", t.key, getCfgDataErr)
		return
	}

	retInfo.ConfigData = string(cfgData)
	return
}
