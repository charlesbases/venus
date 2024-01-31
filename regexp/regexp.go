package regexp

import (
	"regexp"
	"strconv"
)

// compile .
type compile struct {
	r *regexp.Regexp
}

// FindSubString .
func (c compile) FindSubString(str string) string {
	if subs := c.r.FindStringSubmatch(str); len(subs) > 1 {
		return subs[1]
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
	return c.r.MatchString(str)
}

// New .
func New(rp string) compile {
	return compile{r: regexp.MustCompile(rp)}
}
