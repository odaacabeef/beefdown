package ui

import (
	"fmt"

	"github.com/trotttrotttrott/seq/device"
	"github.com/trotttrotttrott/seq/sequence"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	bpm      float64
	device   device.Device
	sequence sequence.Sequence
	err      error
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

	s, err := sequence.New("_test/test.md")
	if err != nil {
		return nil, err
	}

	return &model{
		bpm:      120,
		device:   *d,
		sequence: *s,
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

		case "enter":
			for _, p := range m.sequence.Parts {
				err := m.device.Play(m.bpm, p)
				if err != nil {
					m.err = err
				}
			}
		}
	}

	return m, nil
}

func (m model) View() string {

	s := ""

	if m.err != nil {
		s += fmt.Sprintf("\n%s\n", m.err.Error())
	}

	var parts []string
	for _, p := range m.sequence.Parts {
		part := fmt.Sprintf("\n%s\n\n", p.Name)
		for i, step := range p.StepData {
			part += fmt.Sprintf("%d  %s\n", i+1, step)
		}
		parts = append(parts, lipgloss.NewStyle().
			PaddingRight(5).
			PaddingLeft(2).
			BorderLeft(true).
			BorderStyle(lipgloss.Border{Left: "|"}).
			Render(part))
	}

	s += lipgloss.JoinHorizontal(lipgloss.Bottom, parts...)

	return s
}
