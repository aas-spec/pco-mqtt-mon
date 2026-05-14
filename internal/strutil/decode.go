package strutil

import (
	"golang.org/x/text/encoding/ianaindex"
)

func DecodeTo(codePage string, ba []uint8) []uint8 {
	e, err := ianaindex.IANA.Encoding(codePage)
	if err != nil {
		e, err = ianaindex.MIME.Encoding(codePage)
	}
	if err != nil {
		return ba
	}
	out, _ := e.NewDecoder().Bytes(ba)
	return out
}

func EncodeTo(codePage string, ba []uint8) []uint8 {
	e, err := ianaindex.IANA.Encoding(codePage)
	if err != nil {
		e, err = ianaindex.MIME.Encoding(codePage)
	}
	if err != nil {
		return ba
	}
	out, _ := e.NewEncoder().Bytes(ba)
	return out
}
