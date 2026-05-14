package events

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"telemak.org/internal/model"
	"testing"
)

func TestGetRusEventNameAlarm(t *testing.T) {
	res := GetRusEventName("Alarm", "")
	require.Equal(t,  "Тревога", res)
}
func TestGetRusEventNameAlert(t *testing.T) {
	res := GetRusEventName("Alert", "")
	require.Equal(t, "Тревога",  res)
}
func TestGetRusEventNameFault(t *testing.T) {
	res := GetRusEventName("Fault", "")
	require.Equal(t,  "Неисправность", res)
}
func TestGetRusEventNameFailure(t *testing.T) {
	res := GetRusEventName("Failure", "")
	require.Equal(t, "Неисправность",  res)
}

const testConfig = `
	{
		"Service": "TEST",
		"ServiceID": "TestID1",
		"SrcTopic":     "PCO/Export/Telegram/Events/#",
		"DstTopicBase": "PCO/Monitor/Users/TEST/Devices",
		"SrcCodePage": "UTF-8",
		"DstCodePage": "UTF-8",
		"CodeTranslations": [
			{"CodeFrom": 5320, "CodeTo": 130, "TypeTo": "Alarm"},
			{"CodeFrom": 5321, "CodeTo": 120, "TypeTo": "Alarm"},
			{"CodeFrom": 99,  "CodeTo": 150, "TypeTo": "Alarm"},
			{"CodeFrom": 99,  "CodeTo": 150, "TypeTo": "Alarm", "AddTopic": "Test"},
			{"CodeFrom": 100,  "CodeTo": 160, "TypeTo": "Alarm", "AddTopic": "Test"},
			{"CodeFrom": 101,  "CodeTo": 160, "TypeTo": "Alarm", "AddTopic": ""},
			{"CodeFrom": 102,  "CodeTo": 0, "TypeTo": "Alarm", "AddTopic": ""}
		],
		"DefaultTranslation": {"CodeTo": 0, "TypeTo": "TestType", "AddTopic": "Test"},
		"PassAll": true,
		"DefaultAddTopic": ""
		}`


const testEvent = `{
	   "ServiceID":"TestID1",
	   "EventID":"849454813",
	   "EventTime":"2021-06-17T00:31:54",
	   "EventCategory":"Event",
	   "PultNumber":"986372",
	   "DeviceID":"1802",
	   "DeviceInfo":"",
	   "ObjectName":"АВТОМАТИЧЕСКАЯ ПЕРЕДАЧА В УВО!!!ООО \"ГЕРМЕТ ПЛЮС\" (продуктовый магазин)",
	   "ObjectAddr":"М.О., Люберецкий р-н, г.п. Красково, Школьная ул 6",
	   "EventType":"Alert",
	   "EventFlags":"",
	   "EventCode":"131",
	   "EventText":"Тревога по периметру",
	   "EventNum":"4",
	   "ZoneUser":"4",
	   "ZoneInfo":"Объем торговый зал и вход",
	   "UserName":"",
	   "EventInfo":"Объем торговый зал и вход",
	   "JournalID":"5714060"
		}
	`

func getTestParam() (service model.ServiceItem, event PCOEvent, err error) {
	err = json.Unmarshal( []byte (testConfig), &service)
	if err != nil {
		return
	}
	err = json.Unmarshal([] byte(testEvent), &event )
	if err != nil {
		return
	}
	return
}
func TestConvertEventServiceID(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	// Проверяю известный ServiceID
	event.ServiceID = "TestID1 "
	_, _, ok := ConvertEvent(service, event)
	require.True(t, ok)

	event.ServiceID = "TestID2 , TestID1"
	_, _, ok = ConvertEvent(service, event)
	require.True(t, ok)

	event.ServiceID = "TestID1, TestID3 , TestID2"
	_, _, ok = ConvertEvent(service, event)
	require.True(t, ok)

	event.ServiceID = " TestID4 , TestID1 "
	_, _, ok = ConvertEvent(service, event)
	require.True(t, ok)

	// Проверяю неизвестный ServiceID
	event.ServiceID = "Unknown Service ID, TestID2, TestID3"
	_, _, ok = ConvertEvent(service, event)
	require.False(t, ok)

	// Проверяю все
	service.ServiceID = "*"
	_, _, ok = ConvertEvent(service, event)
	require.True(t, ok)
}

func TestConvertEventUVONumRus(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	// Проверяю в Translation без AddTopic
	event.DeviceInfo="увО=77771111"
	res, _, ok := ConvertEvent(service, event)

	// data, _ := json.MarshalIndent(res, "", "  ")
	// mlog.Printf("Topic: %s, Accepted: %v, Event %s",  "xxx", ok, data)

	require.True(t, ok)
	require.Equal(t, "77771111", res[0].UvoNumber)
}

func TestConvertEventUVONumEn(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	// Проверяю в Translation без AddTopic
	event.DeviceInfo="uVo=77771111"
	res, _, ok := ConvertEvent(service, event)

	// data, _ := json.MarshalIndent(res, "", "  ")
	//mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)

	require.True(t, ok)
	require.Equal(t, "77771111", res[0].UvoNumber)
}


func TestConvertEventWithoutTranslation(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	// Проверяю в Translation без AddTopic
	_, topic, ok := ConvertEvent(service, event)

	// data, _ := json.MarshalIndent(res, "", "  ")
	//mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)

	require.True(t, ok)
	require.Equal(t, service.DstTopicBase, topic)
}

func TestConvertEventSimpleTranslation(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "5320"

	// Проверяю в Translation без AddTopic
	res, topic, ok := ConvertEvent(service, event)

	//data, _ := json.MarshalIndent(res, "", "  ")
	//mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)

	require.Equal(t, 1, len(res)) // Одно значение
	require.Equal(t, 130, res[0].Event.EventCode) // Изменился код на 130
	require.Equal(t, "Alarm", res[0].EventType) // Изменилось на Alarm
	require.True(t, ok)
	require.Equal(t, "", topic)
}


func TestConvertEventSimpleTranslationAddTopic(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "100"

	// Проверяю в Translation без AddTopic
	res, topic, ok := ConvertEvent(service, event)

	//data, _ := json.MarshalIndent(res, "", "  ")
	//mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)

	require.Equal(t, 1, len(res) ) // Одно значение
	require.Equal(t, 160,  res[0].Event.EventCode) // Изменился код на 160
	require.Equal(t,"Alarm", res[0].EventType ) // Изменилось на Alarm
	require.True(t, ok)
	require.Equal(t, service.DstTopicBase + "/Test", topic)
}

func TestConvertEventSimpleTranslationAddTopicEmpty(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "101"

	// Проверяю в Translation без AddTopic
	res, topic, ok := ConvertEvent(service, event)

	// data, _ := json.MarshalIndent(res, "", "  ")
	//mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)

	require.Equal(t, 1, len(res) ) // Одно значение
	require.Equal(t, 160,  res[0].Event.EventCode) // Изменился код на 160
	require.Equal(t,"Alarm", res[0].EventType ) // Изменилось на Alarm
	require.True(t, ok)
	require.Equal(t, service.DstTopicBase, topic) // AddТопик пустой, поэтому Dst
}

func TestConvertEventSimpleTranslationCode0(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "102"

	// Проверяю в Translation без AddTopic
	_,_, ok := ConvertEvent(service, event)
	 require.False(t, ok)
}

func TestConvertEventSimpleTranslationCode0NotPassAll(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "111"
	service.PassAll = false

	// Проверяю в Translation без AddTopic
	_, _, ok := ConvertEvent(service, event)

//	data, _ := json.MarshalIndent(res, "", "  ")
//	mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)
	require.False(t, ok)
}

func TestConvertEventDefTranslationCode0(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "144"

	res, topic, ok := ConvertEvent(service, event)

	// data, _ := json.MarshalIndent(res, "", "  ")
	// mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)

	require.Equal(t, 1, len(res) ) // Одно значение
	require.Equal(t, 144,  res[0].Event.EventCode) // Остался код
	require.Equal(t,"Alarm", res[0].EventType ) // Изменилось на Alarm
	require.True(t, ok)
	require.Equal(t, service.DstTopicBase, topic) // AddТопик пустой, поэтому Default
}

func TestConvertEventDefTranslationCode130(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "144"
	service.DefaultTranslation.CodeTo.SetValid(130)

	res, topic, ok := ConvertEvent(service, event)

	// data, _ := json.MarshalIndent(res, "", "  ")
	// mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)

	require.Equal(t, 1, len(res) ) // Одно значение
	require.Equal(t, 130,  res[0].Event.EventCode) // 130
	require.Equal(t,"TestType", res[0].EventType ) // Изменилось на TestType
	require.True(t, ok)
	require.Equal(t, service.DstTopicBase+ "/Test", topic)
}

func TestConvertEventDefAddTopic(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "144"

	res, topic, ok := ConvertEvent(service, event)

	// data, _ := json.MarshalIndent(res, "", "  ")
	// mlog.Printf("Topic: %s, Accepted: %v, Event %s",  topic, ok, data)

	require.Equal(t, 1, len(res) ) // Одно значение
	require.Equal(t, 144,  res[0].Event.EventCode) // 144
	require.Equal(t,"Alarm", res[0].EventType ) // Изменилось на Alarm
	require.True(t, ok)
	require.Equal(t, service.DstTopicBase, topic)
}

func TestConvertEventDefAddTopicWithValue(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "144"
	service.DefaultAddTopic.SetValid( "DefAddTopic")

	res, topic, ok := ConvertEvent(service, event)

	require.Equal(t, 1, len(res) ) // Одно значение
	require.Equal(t, 144,  res[0].Event.EventCode) // 144
	require.Equal(t,"Alarm", res[0].EventType ) // Изменилось на Alarm
	require.True(t, ok)
	require.Equal(t, service.DstTopicBase +"/DefAddTopic", topic)
}


func TestSmartEventInfo(t *testing.T) {
	service, event, err := getTestParam()
	require.Nil(t, err)

	event.EventCode = "120"
	service.DefaultTranslation.CodeTo.SetValid(130)

	res, _, _ := ConvertEvent(service, event)

	require.NotEqual(t, res[0].Event.Info, event.EventInfo)
	service.SmartEventInfo.SetValid(false)
	res, _, _ = ConvertEvent(service, event)
	require.Equal(t, res[0].Event.Info, event.EventInfo)
}

func TestGetFlagNameCheck(t *testing.T) {
	require.Equal(t, "Проверка", GetFlagName("Check"))
}

func TestGetFlagNameIgnore(t *testing.T) {
	require.Equal(t, "Игнор", GetFlagName("Ignore"))
}

func TestGetFlagNameTest(t *testing.T) {
	require.Equal(t, "Тест", GetFlagName("Test"))
}

func TestGetFlagNameOff(t *testing.T) {
	require.Equal(t, "Отключен", GetFlagName("Off"))
}

func TestGetFlagNameNoObject(t *testing.T) {
	require.Equal(t, "Другое", GetFlagName("NoObject"))
}

func TestGetFlagNameEmpty(t *testing.T) {
	require.Equal(t, "", GetFlagName(""))
}

func TestGetRusEventNameOpen(t *testing.T) {
	require.Equal(t, "Снятие", GetRusEventName("Open", ""))
}

func TestGetRusEventNameClose(t *testing.T) {
	require.Equal(t, "Взятие", GetRusEventName("Close", ""))
}

func TestGetRusEventNameRepair(t *testing.T) {
	require.Equal(t, "Восстановление", GetRusEventName("Repair", ""))
}

func TestGetRusEventNameInfo(t *testing.T) {
	require.Equal(t, "Инфо", GetRusEventName("Info", ""))
}

func TestGetRusEventNameWithFlag(t *testing.T) {
	require.Equal(t, "Тест_Тест", GetRusEventName("Alert_Alarm", "Test"))
}

func TestContainServiceAll(t *testing.T) {
	require.True(t, ContainService("any", "*"))
	require.True(t, ContainService("any", ""))
}

func TestContainServiceFound(t *testing.T) {
	require.True(t, ContainService("svc1, svc2, svc3", "svc2"))
	require.True(t, ContainService(" svc1 , svc2 ", "svc1"))
}

func TestContainServiceNotFound(t *testing.T) {
	require.False(t, ContainService("svc1, svc2", "svc3"))
}

