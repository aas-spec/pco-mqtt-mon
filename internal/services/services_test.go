package services

import (
	"testing"

	"github.com/stretchr/testify/require"
	"telemak.org/internal/events"
	"telemak.org/internal/model"
)

func TestDecodePCOJsonPayload(t *testing.T) {
	service := model.ServiceItem{Service: "TEST"}
	payload := []byte(`{"ServiceID":"SVC1","EventType":"Alert","EventCode":"131","ZoneUser":"4"}`)

	var event events.PCOEvent
	err := decodePCOJsonPayload(service, "test/topic", payload, &event)

	require.NoError(t, err)
	require.Equal(t, "SVC1", event.ServiceID)
	require.Equal(t, "Alert", event.EventType)
	require.Equal(t, "131", event.EventCode)
	require.Equal(t, "4", event.ZoneUser)
}

func TestDecodePCOJsonPayloadInvalid(t *testing.T) {
	service := model.ServiceItem{Service: "TEST"}
	var event events.PCOEvent
	err := decodePCOJsonPayload(service, "test/topic", []byte("not json"), &event)
	require.Error(t, err)
}

func TestDecodePCOJsonPayloadWithCodePage(t *testing.T) {
	service := model.ServiceItem{Service: "TEST", SrcCodePage: "UTF-8"}
	payload := []byte(`{"ServiceID":"SVC1","EventType":"Info"}`)

	var event events.PCOEvent
	err := decodePCOJsonPayload(service, "test/topic", payload, &event)

	require.NoError(t, err)
	require.Equal(t, "SVC1", event.ServiceID)
}
