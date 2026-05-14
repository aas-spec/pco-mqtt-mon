package events

import (
	"regexp"
	"strconv"
	"strings"

	"telemak.org/internal/mlog"
	"telemak.org/internal/model"
	"telemak.org/internal/strutil"
)

type PCOEvent struct {
	ServiceID     string
	EventID       string
	EventTime     string
	EventCategory string
	PultNumber    string
	DeviceID      string
	DeviceInfo    string
	ObjectName    string
	ObjectAddr    string
	EventType     string
	EventFlags    string
	EventCode     string
	EventText     string
	EventNum      string
	ZoneUser      string
	ZoneInfo      string
	UserName      string
	EventInfo     string
	JournalID     string
}

type PCOMonEvent struct {
	EventID   string
	EventTime string
	CodeSet   string
	EventCode int
	ZoneUser  int
	UserName  string
	Info      string
	JournalID int
}

type PCOMonMessage struct {
	Message    string
	MessageID  string
	EventType  string
	PultNumber string
	UvoNumber  string `json:"UVO_Number,omitempty"`
	ServiceID  string `json:"ServiceID,omitempty"`
	Event      PCOMonEvent
}

func checkName(rusRes string, evName string, en string, ru string) string {
	if strings.Contains(evName, en) {
		if rusRes != "" {
			rusRes += "_"
		}
		rusRes += ru
	}
	return rusRes
}

func GetFlagName(evFlags string) string {
	f := checkName("", evFlags, "Check", "Проверка")
	if f == "" {
		f = checkName("", evFlags, "Ignore", "Игнор")
	}
	if f == "" {
		f = checkName("", evFlags, "Test", "Тест")
	}
	if f == "" {
		f = checkName("", evFlags, "Off", "Отключен")
	}
	if f == "" {
		f = checkName("", evFlags, "NoObject", "Другое")
	}
	return f
}

func GetRusEventName(evName string, evFlags string) string {
	f := GetFlagName(evFlags)
	if f != "" {
		l := len(strings.Split(evName, "_"))
		res := ""
		for i := 0; i < l; i++ {
			if i != 0 {
				res += "_"
			}
			res += f
		}
		return res
	}
	r := checkName("", evName, "Alert", "Тревога")
	r = checkName(r, evName, "Alarm", "Тревога")
	r = checkName(r, evName, "Failure", "Неисправность")
	r = checkName(r, evName, "Fault", "Неисправность")
	r = checkName(r, evName, "Open", "Снятие")
	r = checkName(r, evName, "Close", "Взятие")
	r = checkName(r, evName, "Repair", "Восстановление")
	r = checkName(r, evName, "Test", "Тест")
	r = checkName(r, evName, "Other", "Прочее")
	r = checkName(r, evName, "Info", "Инфо")
	return r
}

func ContainService(services string, serviceID string) bool {
	if serviceID == "" || serviceID == "*" {
		return true
	}
	serviceID = strings.TrimSpace(serviceID)
	serviceList := strings.Split(services, ",")
	for _, service := range serviceList {
		if strings.TrimSpace(service) == serviceID {
			return true
		}
	}
	return false
}

func ConvertEvent(service model.ServiceItem, event PCOEvent) ([]PCOMonMessage, string, bool) {
	if !ContainService(event.ServiceID, service.ServiceID) {
		mlog.Outf(service.Service, "Ignored by ServiceID, Event.ServiceID: %s, Service.ServiceID: %s", event.ServiceID, service.ServiceID)
		return nil, "", false
	}

	eventCode, _ := strconv.Atoi(event.EventCode)
	eventType := event.EventType
	topic := ""

	changed := false
	for _, t := range service.CodeTranslations {
		if eventCode != t.CodeFrom {
			continue
		}
		if t.CodeTo.Int64 == int64(0) {
			mlog.Outf(service.Service, "Block EventCode by Translation: %s -> 0", event.EventCode)
			return nil, "", false
		}
		eventCode = int(t.CodeTo.Int64)
		changed = true
		mlog.Outf(service.Service, "Change EventCode by Translation: %s -> %d", event.EventCode, eventCode)
		if t.TypeTo != "" {
			eventType = t.TypeTo
			mlog.Outf(service.Service, "Change EventType by Translation: %s -> %s", event.EventType, eventType)
		}
		if t.AddTopic.Valid {
			topic = service.DstTopicBase
			if t.AddTopic.String != "" {
				topic = topic + "/" + t.AddTopic.String
			}
			mlog.Outf(service.Service, "Change DstTopic by Translation: %s", topic)
		}
		break
	}

	if !changed {
		if service.DefaultTranslation.CodeTo.Int64 != int64(0) {
			eventCode = int(service.DefaultTranslation.CodeTo.Int64)
			changed = true
			mlog.Outf(service.Service, "Change EventCode by Default Translation: %s -> %d", event.EventCode, eventCode)
			if service.DefaultTranslation.TypeTo != "" {
				eventType = service.DefaultTranslation.TypeTo
				mlog.Outf(service.Service, "Change EventType by Default Translation: %s -> %s", event.EventType, eventType)
			}
			if service.DefaultTranslation.AddTopic.Valid {
				topic = service.DstTopicBase
				if service.DefaultTranslation.AddTopic.String != "" {
					topic = topic + "/" + service.DefaultTranslation.AddTopic.String
				}
				mlog.Outf(service.Service, "Change DstTopic by Default Translation: %s", topic)
			}
		}
	}

	if !changed {
		if !service.PassAll {
			mlog.Outf(service.Service, "Block EventCode by Service Settings: %s", event.EventCode)
			return nil, "", false
		}
		if service.DefaultAddTopic.Valid {
			topic = service.DstTopicBase
			if service.DefaultAddTopic.String != "" {
				topic = topic + "/" + service.DefaultAddTopic.String
			}
			mlog.Outf(service.Service, "Change DstTopic by Default AddTopic: %s", topic)
		}
	}

	event.EventType = eventType
	event.EventCode = strconv.Itoa(eventCode)

	if event.EventType == "" {
		event.EventType = "Info"
	}

	curRusName := GetRusEventName(event.EventType, event.EventFlags)
	evRu := strings.Split(curRusName, "_")
	evEn := strings.Split(event.EventType, "_")
	count := len(evEn)
	if len(evRu) < count {
		count = len(evRu)
	}
	res := make([]PCOMonMessage, count)

	for i := 0; i < count; i++ {
		res[i].Message = "Event"
		res[i].MessageID = strutil.GenGUID()
		res[i].EventType = evEn[i]

		res[i].PultNumber = event.PultNumber
		s := strings.Replace(strings.ToUpper(event.DeviceInfo), "УВО=", "UVO=", -1)
		exp, _ := regexp.Compile("UVO=(.*?)(\\s|$|,)")
		arr := exp.FindStringSubmatch(s)
		if len(arr) > 1 {
			res[i].UvoNumber = arr[1]
		}
		res[i].ServiceID = event.ServiceID
		res[i].Event.EventID = event.EventID
		res[i].Event.EventTime = event.EventTime
		res[i].Event.CodeSet = "PCO"

		curEventCode, err := strconv.Atoi(event.EventCode)
		if err != nil {
			curEventCode = 0
		}
		res[i].Event.EventCode = curEventCode

		curZoneUser, err := strconv.Atoi(event.ZoneUser)
		if err != nil {
			curZoneUser = 0
		}
		res[i].Event.ZoneUser = curZoneUser
		if !service.SmartEventInfo.Valid || service.SmartEventInfo.Bool {
			info := evRu[i] + ": " + event.EventText + " (" + strconv.Itoa(curEventCode) + ")"
			infoHandled := false
			if (evEn[i] == "Alert") || (evEn[i] == "Repair") {
				if curZoneUser != 0 {
					info += ", Зона (" + strconv.Itoa(curZoneUser) + ")"
					if event.EventInfo != "" {
						info += " " + event.EventInfo
					}
					infoHandled = true
				}
			} else if (evEn[i] == "Open") || (evEn[i] == "Close") {
				if curZoneUser != 0 {
					info += ", Пользователь (" + strconv.Itoa(curZoneUser) + ")"
					if event.EventInfo != "" {
						res[i].Event.UserName = event.EventInfo
						info += " " + event.EventInfo
					}
					infoHandled = true
				}
			}

			if !infoHandled && event.EventInfo != "" {
				info += ", " + event.EventInfo
			}
			res[i].Event.Info = info
		} else {
			res[i].Event.Info = event.EventInfo
		}

		res[i].Event.JournalID, _ = strconv.Atoi(event.JournalID)
		switch {
		case strings.EqualFold(res[i].EventType, "Alert"):
			res[i].EventType = "Alarm"
		case strings.EqualFold(res[i].EventType, "Failure"):
			res[i].EventType = "Fault"
		}
	}
	return res, topic, true
}
