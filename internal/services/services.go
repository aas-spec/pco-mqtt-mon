package services

import (
	"encoding/json"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"telemak.org/internal/events"
	"telemak.org/internal/mlog"
	"telemak.org/internal/model"
	"telemak.org/internal/strutil"
)

func decodePCOJsonPayload(service model.ServiceItem, topic string, data []byte, v interface{}) error {
	var content []byte
	if service.SrcCodePage != "" {
		content = strutil.DecodeTo(service.SrcCodePage, data)
	} else {
		content = data
	}
	s := string(content)
	content = []byte(strutil.ReplaceLB(s))
	mlog.Outf(service.Service, "Recv from: %s\n%s", topic, string(content))
	return json.Unmarshal(content, v)
}

func HandleMessage(c mqtt.Client, service model.ServiceItem, m mqtt.Message) {
	// Смотрю мой ли топик
	if !strutil.DetermineMQTTTopicInWildcard(m.Topic(), service.SrcTopic) {
		// Топик не мой, ничего не делаю
		return
	}

	var event events.PCOEvent
	err := decodePCOJsonPayload(service, m.Topic(), m.Payload(), &event)
	if err != nil {
		mlog.Outf(service.Service, "Unable to unmarshal event: %s", err)
		return
	}
	go processEvent(c, service, event)
}

func processEvent(c mqtt.Client, service model.ServiceItem, event events.PCOEvent) {
	pcoMsgArr, topic, ok := events.ConvertEvent(service, event)
	if !ok {
		return
	}
	for _, pcoMsg := range pcoMsgArr {
		resultTopic := topic
		if topic == "" { // Если topic пустой, то шлю в Alarm и Event
			if strings.EqualFold(pcoMsg.EventType, "Alert") || strings.EqualFold(pcoMsg.EventType, "Alarm") {
				resultTopic += service.DstTopicBase + "/Alarm"
			} else {
				resultTopic += service.DstTopicBase + "/Event"
			}
		}
		data, _ := json.MarshalIndent(pcoMsg, "", "  ")

		if service.DstCodePage != "" {
			mlog.Outf(service.Service, "Encode %s and Send to: %s %s\n%s", service.DstCodePage, service.Service, resultTopic, string(data))
			data = strutil.EncodeTo(service.DstCodePage, data)
		} else {
			mlog.Outf(service.Service, "Send to: %s %s\n%s", service.Service, resultTopic, string(data))
		}

		token := c.Publish(resultTopic, 0, false, data)
		token.Wait()
	}
}
