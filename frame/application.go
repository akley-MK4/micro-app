package frame

import (
	"encoding/json"
	"fmt"
	"github.com/akley-MK4/go-tools-box/ctime"
	ossignal "github.com/akley-MK4/go-tools-box/signal"
	"github.com/akley-MK4/pubsub"
	"os"
	"syscall"
)

const (
	APPDefaultSignChanSize = 1
)

const (
	MainProcessType ProcessType = iota + 1
	SubProcessType
)

type ApplicationID string
type NewApplication func() IApplication
type ProcessType int
type TopicType uint16
type TopicSubKey interface{}

type IApplication interface {
	baseInitialize(appID ApplicationID, processType ProcessType) error
	Initialize(args ...interface{}) error
	GetID() ApplicationID
	GetProcessType() ProcessType
	importComponents(infoList []componentConfigModel) error
	Pub(topic TopicType, args ...interface{}) error
	Sub(topic TopicType, subKey TopicSubKey, handle pubsub.TopicFunc, preArgs ...interface{}) error
	start() error
	stop() error
	forever()
}

type BaseApplication struct {
	id            ApplicationID
	processType   ProcessType
	signalHandler *ossignal.Handler
	ob            *pubsub.ObServer
	componentMap  map[ComponentID]IComponent
}

func (t *BaseApplication) baseInitialize(id ApplicationID, processType ProcessType) error {
	t.id = id
	t.processType = processType
	t.componentMap = make(map[ComponentID]IComponent)

	t.signalHandler = &ossignal.Handler{}
	if err := t.signalHandler.InitSignalHandler(APPDefaultSignChanSize); err != nil {
		return err
	}

	ob, newObErr := pubsub.NewObServer(true)
	if newObErr != nil {
		return newObErr
	}
	t.ob = ob

	for _, sig := range []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT} {
		t.signalHandler.RegisterSignal(sig, func() {
			t.signalHandler.CloseSignalHandler()
		})
	}

	return nil
}

func (t *BaseApplication) Initialize(args ...interface{}) error {
	return nil
}

func (t *BaseApplication) GetID() ApplicationID {
	return t.id
}

func (t *BaseApplication) GetProcessType() ProcessType {
	return t.processType
}

func (t *BaseApplication) Pub(topic TopicType, args ...interface{}) error {
	return t.ob.Publish(topic, false, args...)
}

func (t *BaseApplication) Sub(topic TopicType, subKey TopicSubKey,
	handle pubsub.TopicFunc, preArgs ...interface{}) error {

	return t.ob.Subscribe(topic, subKey, handle, preArgs...)
}

func (t *BaseApplication) start() error {
	for componentID, component := range t.componentMap {
		getLoggerInst().InfoF("Starting component %v", componentID)
		if err := component.Start(); err != nil {
			return fmt.Errorf("component %v has failed to start, %v", componentID, err)
		}
		component.setStartTimestamp(ctime.CurrentTimestamp())
		getLoggerInst().InfoF("Component %v has started", componentID)
	}

	getLoggerInst().Info("All components have been started")
	return nil
}

func (t *BaseApplication) AfterStart() error {
	return nil
}

func (t *BaseApplication) stop() error {
	for componentID, component := range t.componentMap {
		getLoggerInst().InfoF("Stopping component %v", componentID)
		if err := component.Stop(); err != nil {
			return fmt.Errorf("failed to stop component %v, %v", componentID, err)
		}
		getLoggerInst().InfoF("Component %v has Stopped", componentID)
	}
	return nil
}

func (t *BaseApplication) StopBefore() error {

	return nil
}

func (t *BaseApplication) forever() {
	t.signalHandler.ListenSignal()
}

func (t *BaseApplication) importComponents(cfgList []componentConfigModel) (retErr error) {
	componentNum := 1
	var compTpy ComponentType

	defer func() {
		if retErr != nil {
			retErr = fmt.Errorf("component type %v not imported, %v", compTpy, retErr)
		}
	}()

	for _, cfg := range cfgList {
		if cfg.Disable {
			continue
		}

		compTpy = ComponentType(cfg.ComponentType)
		regInfo, exist := regComponentInfoMap[compTpy]
		getLoggerInst().InfoF("Import component type %v", compTpy)
		if !exist {
			retErr = fmt.Errorf("component type %v dose not exist", compTpy)
			return
		}

		component := regInfo.NewComponent()
		if err := component.baseInitialize(componentNum, compTpy); err != nil {
			retErr = err
			return
		}

		kwData, msErr := json.Marshal(cfg.Kw)
		if msErr != nil {
			retErr = msErr
			return
		}
		kw := regInfo.NewComponentKW()
		if kw != nil {
			if err := json.Unmarshal(kwData, kw); err != nil {
				retErr = err
				return
			}
		}

		if err := component.Initialize(kw); err != nil {
			retErr = err
			return
		}

		t.componentMap[component.GetID()] = component
		getLoggerInst().InfoF("Component %v with type %v has imported", component.GetID(), compTpy)
	}

	return
}
