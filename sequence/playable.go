package sequence

type Playable interface {
	Group() string
	Title() string
	Steps() string
	CurrentStep() *int
	UpdateStep(int)
	ClearStep()
	Warnings() []string
}
