package config

import (
	"encoding/json"
	"os"

	"gopkg.in/guregu/null.v3"

	"telemak.org/internal/mlog"
	"telemak.org/internal/model"
)

type CommonConfig struct {
	LogFiles null.Int
}

type AppConfig struct {
	Common   CommonConfig
	MQTT     model.MQTTConfig
	Services []model.ServiceItem
}

var Data AppConfig

func (c *AppConfig) LoadConfig(cfgFile string) {
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		mlog.Panic("unable to read config file: " + err.Error())
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		mlog.Panic("unable to unmarshal config file:" + err.Error())
	}
}
