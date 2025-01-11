package sequence

import (
	"regexp"
	"strconv"
)

type metadata string

func (m metadata) name() string {
	re := regexp.MustCompile("name:" + reName)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		return match[1]
	}
	return ""
}

func (m metadata) group() string {
	re := regexp.MustCompile("group:" + reGroup)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		return match[1]
	}
	return "default"
}

func (m metadata) channel() (uint8, error) {
	re := regexp.MustCompile("ch:" + reChannel)
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

func (m metadata) bpm() (float64, error) {
	re := regexp.MustCompile("bpm:" + reBPM)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		num, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return 120, err
		}
		return num, nil
	}
	return 120, nil
}

func (m metadata) loop() bool {
	re := regexp.MustCompile("loop:" + reLoop)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		switch match[1] {
		case "true":
			return true
		case "false":
			return false
		}
	}
	return false
}

func (m metadata) sync() string {
	re := regexp.MustCompile("sync:" + reSync)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		switch match[1] {
		case "leader":
			return match[1]
		}
	}
	return "none"
}

func (m metadata) div() int {
	re := regexp.MustCompile("div:" + reDivision)
	match := re.FindStringSubmatch(string(m))
	if len(match) > 0 {
		switch match[1] {
		case "8th":
			return 12
		case "8th-triplet":
			return 8
		case "16th":
			return 6
		case "32nd":
			return 3
		}
	}
	return 24
}
