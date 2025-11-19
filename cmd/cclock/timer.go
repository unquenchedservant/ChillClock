package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) handleTick() (tea.Model, tea.Cmd) {
	if m.timerRunning {
		m.timerElapsed = time.Since(m.timerStart)

		phase1Dur := time.Duration(m.config.Timer.Phase1Duration) * time.Minute
		phase2Dur := time.Duration(m.config.Timer.Phase2Duration) * time.Minute
		phase3Dur := time.Duration(m.config.Timer.Phase3Duration) * time.Minute
		totalDur := phase1Dur + phase2Dur + phase3Dur

		oldPhase := m.currentPhase
		if m.timerElapsed >= totalDur {
			m.currentPhase = phaseCompleted
			m.timerRunning = false
		} else if m.timerElapsed >= phase1Dur+phase2Dur {
			m.currentPhase = phase3
		} else if m.timerElapsed >= phase1Dur {
			m.currentPhase = phase2
		} else {
			m.currentPhase = phase 1
		}

		writeTimerState(m)

		if oldPhase != m.currentPhase && m.currentPhase != phaseNotStarted {
			var temp int
			switch m.currentPhase {
			case phase1:
				temp = m.config.Timer.Phase1Temp
			case phase2:
				temp = m.config.Timer.Phase2Temp
			case phase3:
				temp = m.config.Timer.Phase3Temp
			case phaseCompleted:
				temp = 0
			}
			return m, tea.Batch(tickCmd(), dingCmd(m.currentPhase, temp))
		}
	} else {
		writeTimerState(m)
	}
	return m, tea.Batch(tickCmd(), watchForFileClick())
}

func (m model) getTimerDisplay() (string, lipgloss.Style) {
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	if (!m.timerRunning && m.currentPhase == phaseNotStarted) || m.currentPhase == phaseCompleted {
		return "Press Enter or Space to start timer, '?' for config", whiteStyle
	}

	elapsed := m.timerElapsed
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	timerText := fmt.Sprintf("Timer: %d:%02d", minutes, seconds)

	var style lipgloss.Style
	var temp int
	switch m.currentPhase {
	case phase1:
		style = greenStyle
		temp = m.config.Timer.Phase1Temp
	case phase2:
		style = yellowStyle
		temp = m.config.Timer.Phase2Temp
	case phase3:
		style = redStyle
		temp = m.config.Timer.Phase3Temp
	default:
		style = whiteStyle
	}
	timerText += fmt.Sprintf(" Temp: %d°", temp)
	return timerText, style
}

func writeTimerState(m model) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	timerFile := filepath.Join(homeDir, "dhv_timer.txt")
	var output TimerOutput

	if (!m.timerRunning && m.currentPhase == phaseNotStarted) || m.currentPhase == phaseCompleted {
		output = TimerOutput{Text: "0.00", Class: "white"}
	} else {
		minutes := int(m.timerElapsed.Minutes())
		seconds := int(m.timerElapsed.Seconds()) % 60
		timerText := fmt.Sprintf("%d:%02d", minutes, seconds)

		var class string
		switch m.currentPhase {
		case phase1:
			class = "green"
		case phase2:
			class = "yellow"
		case phase3:
			class = "red"
		default:
			class = "white"
		}

		output = TimerOutput{Text: timerText, Class: class}
	}

	data, err := json.Marshal(output)
	if err != nil {
		return err
	}

	return os.WriteFile(timerFile, data, 0644)
}

func watchForFileClick() tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil
		}

		clickFile := filepath.Join(homeDir, "dhv_timer_click1")
		if _, err := os.Stat(clickFile); err == nil {
			os.remove(clickFile)
			return fileClickMsg{}
		}
		return nil
	}
}