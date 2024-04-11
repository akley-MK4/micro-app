package main

import (
	"errors"
	"github.com/akley-MK4/micro-app/frame"
	"runtime"
	"time"
)

func init() {
	frame.RegisterComponentInfo(frame.ComponentPriorityGeneral, "HTTPAPIServer", func() frame.IComponent {
		return &HTTPAPIServerComponent{}
	}, func() frame.IComponentKW {
		return &HTTPAPIServerComponentKW{}
	})

	frame.RegisterComponentInfo(frame.ComponentPriorityGeneral, "StaticResourceServer", func() frame.IComponent {
		return &MonitorComponent{}
	}, func() frame.IComponentKW {
		return &MonitorComponentKW{}
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

	getGlobalLoggerInstance().InfoF("HTTPAPIServer Initialize KWArgs: %v", kwArgs)
	if !frame.RegisterConfigCallback(frame.ConfigCallbackTypeUpdate, GetConfigHandler(), onUpdateConfigEvent) {
		return errors.New("unable to register configuration callback function")
	}

	return nil
}

func (t *HTTPAPIServerComponent) Start() error {
	return nil
}

type MonitorComponentKW struct {
	ServerAddr string `json:"server_addr"`
}

type MonitorComponent struct {
	frame.BaseComponent
}

func (t *MonitorComponent) Initialize(kw frame.IComponentKW) error {
	kwArgs := kw.(*MonitorComponentKW)
	getGlobalLoggerInstance().InfoF("MonitorComponent Initialize KWArgs: %v", kwArgs)

	go func() {
		for {
			time.Sleep(time.Second * 5)
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			//kbSize := 1024
			//mbSize := kbSize * 1024
			getGlobalLoggerInstance().DebugF("MemStats Bytes, Alloc: %v, TotalAlloc: %v, Sys: %v, HeapAlloc: %v, HeapIdle: %v, "+
				"HeapReleased: %v, NextGC: %v, LastGC: %v, NumGC: %v",
				ms.Alloc, ms.TotalAlloc, ms.Sys, ms.HeapAlloc, ms.HeapIdle, ms.HeapReleased, ms.NextGC, ms.LastGC, ms.NumGC)
		}
	}()
	return nil
}

func onUpdateConfigEvent() {
	getGlobalLoggerInstance().DebugF("Updated the configuration, addr: %s", GetConfig().Addr)
}
