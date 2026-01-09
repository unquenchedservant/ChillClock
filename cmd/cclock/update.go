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
		return m.handleTimerToggle(TIMER_1), watchForFileClick()
	case fileClickMsg2:
		return m.handleTimerToggle(TIMER_2), watchForFileClick()
	case tickMsg:
		return m.handleTick()
	case dingMsg:
		return m, nil
	}
	return m, nil
}

func (m model) handleClockInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "?":
		if !m.timerRunning {
			m.mode = viewConfig
			m.selectedField = fieldPhase1DurationT1
			m.editingField = false
			m.inputBuffer = ""
		}
	case "r":
		if m.timerRunning {
			if m.timer == TIMER_1 {
				m.timer = TIMER_2
			}else if m.timer == TIMER_2 {
				m.timer = TIMER_1
			}
		}
	case "d":
		if !m.timerRunning {
			if m.timerDefault == TIMER_1 {
				m.timerDefault = TIMER_2
			}else if m.timerDefault == TIMER_2 {
				m.timerDefault = TIMER_1
			}
		}
	case "1":
		return m.handleTimerToggle(TIMER_1), nil
	case "2":
		return m.handleTimerToggle(TIMER_2), nil
	case "enter", "":
		return m.handleTimerToggle(m.timerDefault), nil
	}
	return m, nil
}

func (m model) handleTimerToggle(timer int) model {
	if !m.timerRunning {
		m.timerRunning = true
		m.timerStart = time.Now()
		m.timerElapsed = 0
		m.lastPhase = phaseNotStarted
		m.timer = timer
	} else {
		m.timerRunning = false
		m.timerElapsed = 0
		m.currentPhase = phaseNotStarted
		m.lastPhase = phaseNotStarted
	}
	return m
}