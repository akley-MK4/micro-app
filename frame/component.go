package frame

import (
	"encoding/json"
	"fmt"
)

type ComponentType string
type ComponentID string
type ComponentStatus uint16

const (
	ComponentBaseInitStatus = iota
	ComponentInitStatus
	ComponentStartStatus
	ComponentStopStatus
)

const (
	ComponentPriorityLow = iota
	ComponentPriorityGeneral
	ComponentPriorityHigh
)

type NewComponent func() IComponent

type IComponentKW interface{}

type NewComponentKW func() IComponentKW

type RegComponentInfo struct {
	Priority       int
	Tpy            ComponentType
	NewComponent   NewComponent
	NewComponentKW NewComponentKW
}

type IComponent interface {
	baseInitialize(index int, tpy ComponentType) error
	Initialize(kw IComponentKW) error
	GetIndex() int
	GetID() ComponentID
	GetType() ComponentType
	Start() error
	Stop() error
}

var (
	regComponentInfoMap = make(map[ComponentType]*RegComponentInfo)
)

func RegisterComponentInfo(priority int, tpy ComponentType, newComponent NewComponent, newComponentKW NewComponentKW) {
	regComponentInfoMap[tpy] = &RegComponentInfo{
		Priority:       priority,
		Tpy:            tpy,
		NewComponent:   newComponent,
		NewComponentKW: newComponentKW,
	}
}

func createAndInitializeComponent(componentIndex int, cfg componentConfigModel) (retComponent IComponent, retErr error) {
	if cfg.Disable {
		return
	}

	tpy := ComponentType(cfg.ComponentType)
	regInfo, exist := regComponentInfoMap[tpy]
	if !exist {
		retErr = fmt.Errorf("component type %v dose not exist", tpy)
		return
	}

	defer func() {
		if retErr != nil {
			retErr = fmt.Errorf("component type %v error, %v", tpy, retErr)
			return
		}
	}()

	retComponent = regInfo.NewComponent()
	if err := retComponent.baseInitialize(componentIndex, tpy); err != nil {
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

	if err := retComponent.Initialize(kw); err != nil {
		retErr = err
		return
	}

	return
}

type BaseComponent struct {
	index  int
	tpy    ComponentType
	id     ComponentID
	status ComponentStatus
}

func (t *BaseComponent) baseInitialize(index int, tpy ComponentType) error {
	t.index = index
	t.tpy = tpy
	t.id = ComponentID(fmt.Sprintf("%v_%d", tpy, index))

	return nil
}

func (t *BaseComponent) Initialize(args ...interface{}) error {
	return nil
}

func (t *BaseComponent) GetType() ComponentType {
	return t.tpy
}

func (t *BaseComponent) GetIndex() int {
	return t.index
}

func (t *BaseComponent) GetID() ComponentID {
	return t.id
}

func (t *BaseComponent) GetStatus() ComponentStatus {
	return t.status
}

func (t *BaseComponent) Start() error {
	return nil
}

func (t *BaseComponent) Stop() error {
	return nil
}
