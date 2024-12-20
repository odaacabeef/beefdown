package sequence

import "regexp"

type metadata string

func (m metadata) Name() string {
	re := regexp.MustCompile(`name:(.*)`)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		return match[1]
	}
	return ""
}
