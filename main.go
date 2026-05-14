// PCO-MQTT-Mon (Пункт централизованного наблюдения)
package main

import (
	"flag"
	"os"
	"path"
	"time"

	"telemak.org/internal/config"
	"telemak.org/internal/mlog"
	"telemak.org/internal/mqttclient"
)

const (
	DefaultLoggerStoreDays = 10
	MainLoggerStoreDays    = 10
)

var (
	logLevel = flag.Int("logLevel", 5, "Logger Level")
	debug    = flag.Bool("debug", false, "Debug Console")
	cfgPath  = flag.String("config", "", "Path to config file (default: <exe dir>/PCO-MQTT-Mon.cfg)")
)

func main() {
	flag.Parse()
	mlog.SetLogLevel(mlog.DefLoggerID, *logLevel)
	mlog.SetStoreDays(mlog.DefLoggerID, MainLoggerStoreDays)
	cfgFile := *cfgPath
	if cfgFile == "" {
		cfgFile = path.Dir(os.Args[0]) + string(os.PathSeparator) + "PCO-MQTT-Mon.cfg"
	}
	mlog.StdPrintf("PCO MQTT Monitor 2.1 (с) 2021 Telemak")
	enableStdout := *debug
	config.Data.LoadConfig(cfgFile)

	if !config.Data.Common.LogFiles.Valid { // Если не указано количество дней, то по дефолту
		config.Data.Common.LogFiles.SetValid(DefaultLoggerStoreDays)
	}

	os.Setenv("DisableStdout", "0")
	if !enableStdout {
		os.Setenv("DisableStdout", "1")
	}

	pcoStartedMsg := "PCO MQTT Monitor Started"
	mlog.Print(pcoStartedMsg)
	if !enableStdout {
		mlog.StdPrintf(pcoStartedMsg)
	} else {
		mlog.StdPrintf("Console Debug mode On")
	}

	mlog.LPrintf(8, "Log Store Days: %d", config.Data.Common.LogFiles.Int64)

	for _, service := range config.Data.Services {
		serviceName := service.Service
		if !service.IsActive() {
			serviceName = service.Service[2:]
		}
		mlog.SetLogLevel(serviceName, *logLevel)
		mlog.SetStoreDays(serviceName, int(config.Data.Common.LogFiles.Int64))
		mlog.SetLogUseOwnDir(serviceName, true)

		if !service.IsActive() {
			msg := "Service \"%s\" not Started due to Disabling in Config"
			mlog.Outf(serviceName, msg, serviceName)
			mlog.Printf(msg, serviceName)
			if !enableStdout {
				mlog.StdPrintf(msg, serviceName)
			}
		} else {
			msg := "Service \"%s\" Started"
			mlog.Outf(serviceName, msg, serviceName)
			mlog.Printf(msg, serviceName)
			if !enableStdout {
				mlog.StdPrintf(msg, serviceName)
			}
		}
	}
	mqttclient.MQTTConnect(config.Data.MQTT, enableStdout)
	for {
		time.Sleep(5 * time.Second)
	}
}
