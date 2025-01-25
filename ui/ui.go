package ui

import (
	"github.com/odaacabeef/beefdown/device"

	tea "github.com/charmbracelet/bubbletea"
)

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
