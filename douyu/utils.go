package douyu

import (
	"strings"
)

func Escaped(v string) string {
	vv := strings.Replace(v, "@", "@A", -1)
	vv = strings.Replace(vv, "/", "@S", -1)
	return vv
}

func Unescape(v string) string {
	return v
}
