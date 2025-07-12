package ui

import (
	"github.com/odaacabeef/beefdown/device"

	tea "github.com/charmbracelet/bubbletea"
)

func StartWithOutput(sequencePath string, midiOutput string) error {

	m, err := initialModelWithOutput(sequencePath, midiOutput)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func initialModelWithOutput(sequencePath string, midiOutput string) (*model, error) {
	d, err := device.New(midiOutput)
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
