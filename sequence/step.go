package sequence

import (
	"regexp"
	"strconv"
	"strings"
)

type step string

func (s *step) mult() (*int64, error) {
	match := regexp.MustCompile(reMult).FindStringSubmatch(string(*s))
	var m int64 = 1
	var err error
	if len(match) > 0 {
		m, err = strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}

func (s *step) names() []string {
	var n []string
	for _, f := range strings.Fields(string(*s)) {
		if !regexp.MustCompile(`^([0-9A-Za-z'_-]+)`).MatchString(f) {
			continue
		}
		n = append(n, f)
	}
	return n
}
