package types

import (
	"regexp"
	"strconv"
	"strings"
)

// compile .
type compile struct {
	Regexp *regexp.Regexp
}

// FindSubString .
func (c compile) FindSubString(str string) string {
	if subs := c.Regexp.FindStringSubmatch(str); len(subs) > 1 {
		return subs[1]
	}
	return ""
}

// FinsSubStringBeforePrefix .
func (c compile) FinsSubStringBeforePrefix(str string, prefix string) string {
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, prefix) {
		return c.FindSubString(str)
	}
	return ""
}

// FindSubNumber .
func (c compile) FindSubNumber(str string) int64 {
	var subint int64
	if substr := c.FindSubString(str); len(substr) != 0 {
		subint, _ = strconv.ParseInt(substr, 10, 64)
	}
	return subint
}

// MatchString .
func (c compile) MatchString(str string) bool {
	return c.Regexp.MatchString(str)
}

// MustCompile .
func MustCompile(rp string) compile {
	return compile{Regexp: regexp.MustCompile(rp)}
}
