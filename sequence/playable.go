package sequence

type Playable interface {
	Name() string
	Group() string
	Title() string
	Steps() string
	CurrentStep() *int
	UpdateStep(int)
	ClearStep()
	Warnings() []string
}
