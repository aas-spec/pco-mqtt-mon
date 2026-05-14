package mqttclient

import (
	"os"
	"path/filepath"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"telemak.org/internal/config"
	"telemak.org/internal/mlog"
	"telemak.org/internal/model"
	"telemak.org/internal/services"
	"telemak.org/internal/strutil"
)

var MQTTEnableStdout = false

func MQTTConnect(cfg model.MQTTConfig, enableStdout bool) {
	MQTTEnableStdout = enableStdout
	mlog.Logf("Connecting to MQTT Server %s:%d ", cfg.Host, cfg.Port)
	if !enableStdout {
		mlog.StdPrintf("Connecting to MQTT Server %s:%d ", cfg.Host, cfg.Port)
	}

	opts := mqtt.NewClientOptions().SetAutoReconnect(true)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectionLostHandler
	opts.AddBroker("tcp://" + cfg.Host + ":" + strconv.Itoa(cfg.Port))
	opts.ClientID = getClientName()
	opts.Username = cfg.User
	opts.Password = cfg.Password
	if cfg.KeepAlive != 0 {
		opts.KeepAlive = cfg.KeepAlive
	}
	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		mlog.Logf("MQTT Connect error: %s", token.Error())
	}
}

func getClientName() string {
	h, err := os.Hostname()
	if err != nil {
		mlog.Logln("Error: Unable to get Hostname")
		h = "localhost"
	}
	curPath := os.Args[0]
	return strutil.ChangeFileExt(filepath.Base(curPath), "") + "-" + h
}

func subscribeHandler(c mqtt.Client, m mqtt.Message) {
	mlog.Log("Handle Event")
	for _, service := range config.Data.Services {
		if !service.IsActive() {
			continue
		}
		services.HandleMessage(c, service, m)
	}
}

func connectHandler(c mqtt.Client) {
	msg := "MQTT Connected"
	mlog.Log(msg)
	if !MQTTEnableStdout {
		mlog.StdPrintf(msg)
	}

	// Подписываюсь на все топики
	for _, service := range config.Data.Services {
		if !service.IsActive() {
			continue
		}
		c.Subscribe(service.SrcTopic, 0, subscribeHandler)

		msgSubscribed := "Subscribed to Topic: %v"
		mlog.Outf(service.Service, msgSubscribed, service.SrcTopic)
		mlog.Printf(msgSubscribed, service.SrcTopic)
		if !MQTTEnableStdout {
			mlog.StdPrintf(msgSubscribed, service.SrcTopic)
		}
	}
}

func connectionLostHandler(c mqtt.Client, e error) {
	mlog.Log("MQTT Disconnected " + e.Error())
}
