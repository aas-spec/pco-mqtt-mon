package strutil

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestChangeFileExt(t *testing.T) {
	require.Equal(t, "file.txt", ChangeFileExt("file.go", ".txt"))
	require.Equal(t, "file", ChangeFileExt("file.go", ""))
	require.Equal(t, "/path/to/file.log", ChangeFileExt("/path/to/file.go", ".log"))
}

func TestGenGUID(t *testing.T) {
	g1 := GenGUID()
	g2 := GenGUID()
	require.NotEqual(t, g1, g2)
	require.True(t, strings.HasPrefix(g1, "{"))
	require.True(t, strings.HasSuffix(g1, "}"))
	require.Len(t, g1, 38) // {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
}

func TestJSONPrettyPrint(t *testing.T) {
	out := JSONPrettyPrint(`{"a":1,"b":2}`)
	require.Contains(t, out, "\n")
	require.Contains(t, out, "\t")
	require.Equal(t, "not json", JSONPrettyPrint("not json"))
}

func TestReplaceLB(t *testing.T) {
	require.Equal(t, "hello world", ReplaceLB("hello\nworld"))
	require.Equal(t, "hello world", ReplaceLB("hello\r\nworld"))
	require.Equal(t, "hello world", ReplaceLB("hello\rworld"))
	require.Equal(t, "no change", ReplaceLB("no change"))
}

func TestDetermineMQTTTopicInWildcard(t *testing.T) {
	srcTopic := "PCO/Export/Telegram/Events/#"
	trueTopic := "PCO/Export/Telegram/Events/Test/Test"
	falseTopic := "PCO/Export/Telegram/Event/sTest/Test"
	resTrue := DetermineMQTTTopicInWildcard(trueTopic, srcTopic)
	resFalse := DetermineMQTTTopicInWildcard(falseTopic, srcTopic)
	require.True(t, resTrue)
	require.False(t, resFalse)
}
func TestDetermineMQTTTopicInWildcardPlus(t *testing.T) {
	srcTopic := "PCO/Export/Telegram/+/Events"
	trueTopic := "PCO/Export/Telegram/User1/Events"
	falseTopic := "PCO/Export/Telegram/Event/Test"
	resTrue := DetermineMQTTTopicInWildcard(trueTopic, srcTopic)
	resFalse := DetermineMQTTTopicInWildcard(falseTopic, srcTopic)
	require.True(t, resTrue)
	require.False(t, resFalse)
}

func TestDetermineMQTTTopicInWildcardMultiple(t *testing.T) {
	require.True(t, DetermineMQTTTopicInWildcard("a/b/c/d", "a/+/+/d"))
	require.False(t, DetermineMQTTTopicInWildcard("a/b/c/d/e", "a/+/+/d"))
}

func TestDetermineMQTTTopicInWildcardExact(t *testing.T) {
	require.True(t, DetermineMQTTTopicInWildcard("a/b/c", "a/b/c"))
	require.False(t, DetermineMQTTTopicInWildcard("a/b/x", "a/b/c"))
}
