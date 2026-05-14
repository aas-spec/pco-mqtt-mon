package model

import (
	"gopkg.in/guregu/null.v3"
	"strings"
)

type DefaultTranslation struct {
	CodeTo null.Int
	TypeTo string
	AddTopic null.String
}

type CodeTranslation struct {
	CodeFrom int
	DefaultTranslation
}

type ServiceItem struct {
	Service            string
	ServiceID          string
	SrcTopic           string
	DstTopicBase       string
	SrcCodePage        string
	DstCodePage        string
	CodeTranslations   []CodeTranslation
	DefaultTranslation DefaultTranslation
	PassAll            bool
	DefaultAddTopic    null.String
	SmartEventInfo     null.Bool
}

func (service ServiceItem) IsActive() bool {
	return !strings.HasPrefix(service.Service, "--")
}
