package main

import (
	"encoding/json"
	"github.com/akley-MK4/micro-app/frame"
)

var (
	configHandler = &ConfigHandler{
		cfg: &SidecarConfig{},
	}

	regConfigInfoMap = map[string]frame.ConfigRegInfo{
		"http_api_routes": {
			Suffix: "json", MustLoad: false, NewConfigHandlerFunc: func() frame.IConfigHandler {
				return GetConfigHandler()
			},
			RetryWatchIntervalSec: 5,
		},
	}
)

func registerConfigs() {
	for key, info := range regConfigInfoMap {
		info.Key = key
		frame.RegisterConfigInfo(info)
	}
}

type ConfigHandler struct {
	cfg *SidecarConfig
}

func (t *ConfigHandler) OnUpdate() {

}

func (t *ConfigHandler) EncodeConfig(data []byte) error {
	v := &SidecarConfig{}
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}

	t.cfg = v
	return nil
}

func (t *ConfigHandler) GetConfigData() ([]byte, error) {
	if t.cfg == nil {
		return []byte{}, nil
	}

	return json.Marshal(t.cfg)
}

type SidecarConfig struct {
	Addr string `json:"addr,omitempty"`
}

func GetConfigHandler() *ConfigHandler {
	return configHandler
}

func GetConfig() *SidecarConfig {
	return configHandler.cfg
}
