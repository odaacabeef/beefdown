package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/trotttrotttrott/seq/device"
	"github.com/trotttrotttrott/seq/sequence"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	bpm      float64
	device   *device.Device
	sequence *sequence.Sequence

	ctx    context.Context
	cancel context.CancelFunc

	errs []error
}

func Start() error {

	m, err := initialModel()
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
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

	ctx, cancel := context.WithCancel(context.Background())

	return &model{
		bpm:      120,
		device:   d,
		sequence: s,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

type deviceStop <-chan struct{}

func listenForDeviceStop(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		return deviceStop(ctx.Done())
	}
}

func (m model) Init() tea.Cmd {
	return listenForDeviceStop(m.ctx)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case deviceStop:
		return m, listenForDeviceStop(m.ctx)

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case " ":
			switch {
			case m.device.Stopped():

				// a := m.sequence.Arrangements[0]
				a := m.sequence.Parts[0]

				m.ctx, m.cancel = context.WithCancel(context.Background())
				m.device.Play(m.ctx, m.bpm, a)
			case m.device.Playing():
				m.cancel()
			}
		}
	}

	return m, nil
}

func (m model) View() string {

	st := style(lipgloss.NewStyle())

	s := ""

	s += st.state().Render(fmt.Sprintf("state: %s", m.device.State()))

	if len(m.errs) > 0 {
		s += st.errors().Render(fmt.Sprintf("%v", m.errs))
	}

	var blocks []string

	for _, p := range m.sequence.Parts {
		part := fmt.Sprintf("%s (ch:%d)\n\n", p.Name, p.Channel)
		var steps []string
		for i, step := range p.StepData {
			steps = append(steps, fmt.Sprintf("%d  %s", i+1, step))
		}
		part += strings.Join(steps, "\n")
		blocks = append(blocks, st.block().Render(part))
	}

	for _, a := range m.sequence.Arrangements {
		arrangement := fmt.Sprintf("%s\n\n", a.Name)
		var steps []string
		for i, step := range a.StepData {
			steps = append(steps, fmt.Sprintf("%d  %s", i+1, step))
		}
		arrangement += strings.Join(steps, "\n")
		blocks = append(blocks, st.block().Render(arrangement))
	}

	s += lipgloss.JoinHorizontal(lipgloss.Top, blocks...)

	return s
}
