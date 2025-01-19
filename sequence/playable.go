package sequence

import "time"

type Playable interface {
	Name() string
	Group() string
	Title(float64) string
	Steps() string
	CurrentStep() *int
	UpdateStep(int)
	ClearStep()
	Warnings() []string
	duration(float64) time.Duration
}
