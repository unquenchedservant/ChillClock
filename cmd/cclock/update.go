package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == viewConfig{
			return m.handleConfigInput(msg)
		}
		return m.handleClockInput(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case fileClickMsg:
		return m.handleTimerToggle(), watchForFileClick()
	case tickMsg:
		return m.handleTick()
	case dingMsg:
		return m, nil
	}
	return m, nil
}

func (m model) handleClockInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "?":
		if !m.timerRunning {
			m.mode = viewConfig
			m.selectedField = fieldPhase1Duration
			m.editingField = false
			m.inputBuffer = ""
		}
	case "enter", "":
		return m.handleTimerToggle(), nil
	}
	return m, nil
}

func (m model) handleTimerToggle() model {
	if !m.timerRunning {
		m.timerRunning = true
		m.timerStart = time.Now()
		m.timerElapsed = 0
		m.currentPhase = phase1
		m.lastPhase = phaseNotStarted
	} else {
		m.timerRunning = false
		m.timerElapsed = 0
		m.currentPhase = phaseNotStarted
		m.lastPhase = phaseNotStarted
	}
	return m
}