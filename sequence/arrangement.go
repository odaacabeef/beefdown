package sequence

type Arrangement struct {
	Name string

	metadata metadata
	StepData []string

	Parts [][]*Part
}

func (a *Arrangement) parseMetadata() {
	a.Name = a.metadata.Name()
}
