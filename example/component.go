package example

import (
	"github.com/akley-MK4/micro-app/frame"
	"github.com/akley-MK4/rstlog"
)

func init() {
	frame.RegisterComponentInfo(frame.ComponentPriorityGeneral, "HTTPAPIServer", func() frame.IComponent {
		return &HTTPAPIServerComponent{}
	}, func() frame.IComponentKW {
		return &HTTPAPIServerComponentKW{}
	})

	frame.RegisterComponentInfo(frame.ComponentPriorityGeneral, "StaticResourceServer", func() frame.IComponent {
		return &StaticResourceServerComponent{}
	}, func() frame.IComponentKW {
		return &StaticResourceServerComponentKW{}
	})
}

type HTTPAPIServerComponentKW struct {
	ServerAddr string `json:"server_addr"`
}

type HTTPAPIServerComponent struct {
	frame.BaseComponent
}

func (t *HTTPAPIServerComponent) Initialize(kw frame.IComponentKW) error {
	kwArgs := kw.(*HTTPAPIServerComponentKW)

	rstlog.GetDefaultLogger().InfoF("HTTPAPIServer Initialize KWArgs: %v", kwArgs)
	return nil
}

type StaticResourceServerComponentKW struct {
	ServerAddr string `json:"server_addr"`
}

type StaticResourceServerComponent struct {
	frame.BaseComponent
}

func (t *StaticResourceServerComponent) Initialize(kw frame.IComponentKW) error {
	kwArgs := kw.(*StaticResourceServerComponentKW)

	rstlog.GetDefaultLogger().InfoF("StaticResourceServer Initialize KWArgs: %v", kwArgs)
	return nil
}
