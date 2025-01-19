package sequence

import "time"

type Playable interface {
	Name() string
	Group() string
	Title() string
	Steps() string
	CurrentStep() *int
	UpdateStep(int)
	ClearStep()
	Warnings() []string
	calcDuration(float64)
	Duration() time.Duration
}
