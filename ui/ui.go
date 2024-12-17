package ui

import (
	"github.com/trotttrotttrott/seq/midi"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	MIDI midi.MIDI
	err  error
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
	m, err := midi.NewMIDI()
	if err != nil {
		return nil, err
	}
	return &model{
		MIDI: *m,
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

			s := midi.Sequence{
				BPM:      150,
				Messages: midi.Messages(),
			}

			err := m.MIDI.Play(s)
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
