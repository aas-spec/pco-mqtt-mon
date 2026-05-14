package strutil

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEncodeTo(t *testing.T) {
	s := "Привет"
	s1 := EncodeTo("windows-1251", []uint8(s))
	print(string(s1))
	s2 := DecodeTo("windows-1251	", []uint8(s1))
	print(string(s2))
	require.NotEqual(t, s, string(s1))
	require.NotEqual(t, string(s1), string(s2))
	require.Equal(t, string(s2), s)
}
