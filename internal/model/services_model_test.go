package model

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsActive(t *testing.T) {
	require.True(t, ServiceItem{Service: "MyService"}.IsActive())
	require.True(t, ServiceItem{Service: "-MyService"}.IsActive())
	require.False(t, ServiceItem{Service: "--MyService"}.IsActive())
	require.False(t, ServiceItem{Service: "--"}.IsActive())
}
