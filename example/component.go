package example

import (
	"github.com/akley-MK4/micro-app/frame"
	"github.com/akley-MK4/rstlog"
)

func init() {
	frame.RegisterComponentInfo(frame.ComponentPriorityGeneral, "LoginMgr", func() frame.IComponent {
		return &LoginMgrComponent{}
	}, func() frame.IComponentKW {
		return &LoginMgrComponentKW{}
	})

	frame.RegisterComponentInfo(frame.ComponentPriorityGeneral, "PayMgr", func() frame.IComponent {
		return &PayMgrComponent{}
	}, func() frame.IComponentKW {
		return &PayMgrComponentKW{}
	})
}

type LoginMgrComponentKW struct {
	ServerAddr string `json:"server_addr"`
}

type LoginMgrComponent struct {
	frame.BaseComponent
}

func (t *LoginMgrComponent) Initialize(kw frame.IComponentKW) error {
	kwArgs := kw.(*LoginMgrComponentKW)

	rstlog.GetDefaultLogger().InfoF("LoginMgrComponent Initialize KWArgs: %v", kwArgs)
	return nil
}

type PayMgrComponentKW struct {
	ServerAddr string `json:"server_addr"`
}

type PayMgrComponent struct {
	frame.BaseComponent
}

func (t *PayMgrComponent) Initialize(kw frame.IComponentKW) error {
	kwArgs := kw.(*PayMgrComponentKW)

	rstlog.GetDefaultLogger().InfoF("PayMgrComponent Initialize KWArgs: %v", kwArgs)
	return nil
}
