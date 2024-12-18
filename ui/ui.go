package ui

import (
	"github.com/trotttrotttrott/seq/device"
	"github.com/trotttrotttrott/seq/sequence"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	bpm       float64
	device    device.Device
	sequences []sequence.Sequence
	err       error
}

func Start() error {

	m, err := initialModel()
	if err != nil {
		return err
	}

	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}

func initialModel() (*model, error) {
	d, err := device.New()
	if err != nil {
		return nil, err
	}

	return &model{
		bpm:       120,
		device:    *d,
		sequences: sequence.List(),
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":

			return m, tea.Quit

		case "enter", " ":

			s := m.sequences[0]

			err := m.device.Play(m.bpm, s)
			if err != nil {
				m.err = err
			}
		}
	}

	return m, nil
}

func (m model) View() string {

	s := "Press enter to play some notes\n\n"

	if m.err != nil {
		s += m.err.Error()
	}

	s += "\n\nPress q to quit.\n"

	return s
}
