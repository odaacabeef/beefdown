package ui

import (
	"context"

	"github.com/odaacabeef/beefdown/device"
	"github.com/odaacabeef/beefdown/sequence"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	device   *device.Device
	sequence *sequence.Sequence

	groupNames []string
	groups     map[string][]sequence.Playable
	groupX     map[string]int

	selected coordinates
	playing  *coordinates

	viewport *viewport

	stop context.CancelFunc

	errs []error
}

type coordinates struct {
	x, y int
}

func Start(sequencePath string) error {

	m, err := initialModel(sequencePath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func initialModel(sequencePath string) (*model, error) {
	d, err := device.New()
	if err != nil {
		return nil, err
	}

	m := model{
		device:   d,
		viewport: &viewport{},
	}

	err = m.loadSequence(sequencePath)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *model) loadSequence(sequencePath string) error {
	s, err := sequence.New(sequencePath)
	if err != nil {
		return err
	}
	m.sequence = s
	m.groups = map[string][]sequence.Playable{}
	m.groupNames = []string{}
	m.groupX = map[string]int{}
groupPlayables:
	for _, p := range m.sequence.Playable {
		m.groups[p.Group()] = append(m.groups[p.Group()], p)
		for _, name := range m.groupNames {
			if name == p.Group() {
				continue groupPlayables
			}
		}
		m.groupNames = append(m.groupNames, p.Group())
		m.groupX[p.Group()] = 0
	}
	return nil
}

type devicePlay struct{}

func listenForDevicePlay(c chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return devicePlay(<-c)
	}
}

type deviceStop struct{}

func listenForDeviceStop(c chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return deviceStop(<-c)
	}
}

type deviceClock struct{}

func listenForDeviceClock(c chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return deviceClock(<-c)
	}
}

type deviceError error

func listenForDeviceErrors(err chan error) tea.Cmd {
	return func() tea.Msg {
		return deviceError(<-err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		listenForDevicePlay(m.device.PlayCh()),
		listenForDeviceErrors(m.device.ErrorsCh()),
	)
}
