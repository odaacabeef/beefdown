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
	errs     []error
	errCh    chan (error)
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
		errCh:    make(chan error),
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

			// TODO: blocks if attempted while device playing
			return m, tea.Quit

		case "enter":
			a := m.sequence.Arrangements[0]
			m.device.Play(m.bpm, a)

		case " ":
			if m.device.Playing() {
				m.device.Stop <- true
			}
		}
	}

	return m, nil
}

func (m model) View() string {

	s := "\n"

	s += fmt.Sprintf("  state: %s\n", m.device.State())

	for _, err := range m.errs {
		s += fmt.Sprintf("\n%s\n", err.Error())
	}

	var parts []string
	for _, p := range m.sequence.Parts {
		part := fmt.Sprintf("\n%s (ch:%d)\n\n", p.Name, p.Channel)
		for i, step := range p.StepData {
			part += fmt.Sprintf("%d  %s\n", i+1, step)
		}
		parts = append(parts, lipgloss.NewStyle().
			PaddingRight(5).
			PaddingLeft(2).
			Render(part))
	}

	s += lipgloss.JoinHorizontal(lipgloss.Bottom, parts...)

	var arrangements []string
	for _, a := range m.sequence.Arrangements {
		arrangement := fmt.Sprintf("\n%s\n\n", a.Name)
		for i, step := range a.StepData {
			arrangement += fmt.Sprintf("%d  %s\n", i+1, step)
		}
		arrangements = append(arrangements, lipgloss.NewStyle().
			PaddingRight(5).
			PaddingLeft(2).
			Render(arrangement))
	}

	s += lipgloss.JoinVertical(lipgloss.Bottom, arrangements...)

	return s
}
