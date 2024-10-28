package frame

import (
	"errors"

	"github.com/akley-MK4/pubsub"
)

type EventType uint16
type EventSubKey interface{}

const (
	EventAPPStarted EventType = iota + 1
)

var (
	eventMessageMgrInst    *eventMessageMgr
	ErrInvalidEventMsgInst = errors.New("invalid eventMessageMgr instance")
)

type eventMessageMgr struct {
	obs *pubsub.ObServer
}

func initializeEventMessageMgr() error {
	obs, newObsErr := pubsub.NewObServer(true)
	if newObsErr != nil {
		return newObsErr
	}

	eventMessageMgrInst = &eventMessageMgr{
		obs: obs,
	}

	return nil
}

func PublishEventMessage(event EventType, args ...interface{}) error {
	if eventMessageMgrInst == nil {
		return ErrInvalidEventMsgInst
	}

	return eventMessageMgrInst.obs.Publish(event, false, args...)
}

func SubscribeEventMessage(event EventType, subKey EventSubKey, handle pubsub.TopicFunc, preArgs ...interface{}) error {
	if eventMessageMgrInst == nil {
		return ErrInvalidEventMsgInst
	}

	return eventMessageMgrInst.obs.Subscribe(event, subKey, handle, preArgs...)
}
