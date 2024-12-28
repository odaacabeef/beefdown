package sequence

import (
	"regexp"
	"strconv"
)

type metadata string

func (m metadata) name() string {
	re := regexp.MustCompile(`name:([0-9A-Za-z_-]+)`)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		return match[1]
	}
	return ""
}

func (m metadata) channel() (uint8, error) {
	re := regexp.MustCompile(`ch:([0-9]+)`)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		num, err := strconv.ParseUint(match[1], 10, 8)
		if err != nil {
			return 1, err
		}
		return uint8(num), nil
	}
	return 1, nil
}
