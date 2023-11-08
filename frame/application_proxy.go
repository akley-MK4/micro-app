package frame

import (
	"errors"
	"github.com/akley-MK4/pubsub"
)

var (
	appProxy = &ApplicationProxy{}
)

func GetAppProxy() *ApplicationProxy {
	return appProxy
}

func setAppProxy(inst IApplication) {
	appProxy.appInstance = inst
}

type ApplicationProxy struct {
	appInstance IApplication
}

func (t *ApplicationProxy) Pub(topic TopicType, args ...interface{}) error {
	if t.appInstance == nil {
		return errors.New("app instance is a nil pointer")
	}
	return t.appInstance.Pub(topic, args...)
}

func (t *ApplicationProxy) Sub(topic TopicType, subKey TopicSubKey,
	handle pubsub.TopicFunc, preArgs ...interface{}) error {
	if t.appInstance == nil {
		return errors.New("app instance is a nil pointer")
	}
	return t.appInstance.Sub(topic, subKey, handle, preArgs...)
}

func (t *ApplicationProxy) GetProcessType() ProcessType {
	if t.appInstance == nil {
		return 0
	}
	return t.appInstance.GetProcessType()
}

func (t *ApplicationProxy) IsMainProcess() bool {
	if t.appInstance == nil {
		return false
	}
	return t.appInstance.GetProcessType() == MainProcessType
}
