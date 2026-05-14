package strutil

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nu7hatch/gouuid"

	"telemak.org/internal/mlog"
)

func ChangeFileExt(filePath string, ext string) string {
	return filePath[:len(filePath)-len(filepath.Ext(filePath))] + ext
}

func ParamStr(idx int) string {
	return os.Args[idx]
}

func GenGUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		mlog.Logln(err)
		panic(err)
	}
	return "{" + u.String() + "}"
}

func JSONPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

func ReplaceLB(in string) string {
	replacement := " "
	replacer := strings.NewReplacer(
		"\r\n", replacement,
		"\r", replacement,
		"\n", replacement,
		"\v", replacement,
		"\f", replacement,
		"\u0085", replacement,
		"\u2028", replacement,
		"\u2029", replacement,
	)
	return replacer.Replace(in)
}

func preQuoteRegexp(exp string) string {
	res := ""
	specChars := `\\.+*?[^]$(){}|-`
	for _, c := range exp {
		sc := string(c)
		if strings.Contains(specChars, sc) {
			res = res + "\\"
		} else if c == ' ' {
			res = res + "\\s"
			continue
		}
		res = res + sc
	}
	return res
}

func DetermineMQTTTopicInWildcard(topic string, wildcard string) bool {
	specCharsArr := make([]rune, 0)
	for _, r := range wildcard {
		if r == '#' || r == '+' {
			specCharsArr = append(specCharsArr, r)
		}
	}
	expression := preQuoteRegexp(wildcard)
	expression = strings.ReplaceAll(expression, "#", "(.*)")
	expression = strings.ReplaceAll(expression, "\\+", "(.*?)")

	exp, _ := regexp.Compile("^" + expression + "$")
	arr := exp.FindStringSubmatch(topic)
	if len(arr) != len(specCharsArr)+1 {
		return false
	}
	for idx, s := range arr[1:] {
		if specCharsArr[idx] == '+' && strings.Contains(s, "/") {
			return false
		}
	}
	return true
}
