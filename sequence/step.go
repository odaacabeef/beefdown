package sequence

import (
	"regexp"
	"strconv"
	"strings"
)

type step string

// mult returns the multiplication factor and modulo factor for a step
// Format: *N%M where N is the multiplication factor and M is the modulo factor
// If no modulo is specified, modulo factor is 0 (no modulo)
func (s *step) mult() (*int64, *int64, error) {
	// Match complete patterns: *N or *N%M (but not *N%)
	// First try to match *N%M pattern
	match := regexp.MustCompile(`\*([[:digit:]]+)%([[:digit:]]+)`).FindStringSubmatch(string(*s))
	var m int64 = 1
	var modulo int64 = 0
	var err error
	
	if len(match) > 0 {
		// Found *N%M pattern
		m, err = strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		modulo, err = strconv.ParseInt(match[2], 10, 64)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// Try to match just *N pattern (without modulo)
		match = regexp.MustCompile(`\*([[:digit:]]+)$`).FindStringSubmatch(string(*s))
		if len(match) > 0 {
			m, err = strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return &m, &modulo, nil
}

// multFactor returns just the multiplication factor (for backward compatibility)
func (s *step) multFactor() (*int64, error) {
	m, _, err := s.mult()
	return m, err
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
