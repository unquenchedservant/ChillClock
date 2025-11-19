package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/unquenchedservant/ChillClock/config"
	"github.com/unquenchedservant/ChillClock/utilities"
)

func (m model) handleConfigInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingField {
		switch msg.String() {
		case "enter", "esc":
			if m.inputBuffer == ""{
				m.setFieldValue(m.previousValue)
			} else if val := m.parseInput(); val >= 0 {
				m.setFieldValue(val)
				config.SaveConfig(m.config)
			} else {
				m.setFieldValue(m.previousValue)
			}
			m.editingField = false
			m.inputBuffer = ""
		case "backspace": 
			if len(m.inputBuffer) > 0 {
				m.inputBuffer = m.inputBuffer[:len(m.inputBuffer) - 1]
			}
		case "up", "k":
			m.saveAndExitField()
			if m.selectedField > 0 {
				m.selectedField--
			}
		case "down", "j":
			m.saveAndExitField()
			if m.selectedField < fieldMax - 1 {
				m.selectedField++
			}
		default:
			if len(msg.String()) == 1 && msg.String()[0] >= '0' && msg.String()[0] <= '9' {
				m.inputBuffer += msg.String()
			}
		}
	} else {
		switch msg.String() {
		case "esc", "q", "?": 
			m.mode = viewClock
		case "up", "k":
			if m.selectedField > 0 {
				m.selectedField--
			}
		case "down", "j":
			if m.selectedField < fieldMax - 1 {
				m.selectedField++
			}
		case "enter", " ":
			m.previousValue = m.getFieldValue()
			m.editingField = true
			m.inputBuffer = ""
		}
	}
	return m, nil
}

func (m *model) saveAndExitField() {
	if m.inputBuffer == "" {
		m.setFieldValue(m.previousValue)
	} else if val := m.parseInput(); val >= 0 {
		m.setFieldValue(val)
		config.SaveConfig(m.config)
	} else {
		m.setFieldValue(m.previousValue)
	}
	m.editingField = false
	m.inputBuffer = ""
}

func (m model) getFieldValue() int {
	switch m.selectedField {
	case fieldPhase1Duration:
		return m.config.Timer.Phase1Duration
	case fieldPhase2Duration:
		return m.config.Timer.Phase2Duration
	case fieldPhase3Duration:
		return m.config.Timer.Phase3Duration
	case fieldPhase1Temp:
		return m.config.Timer.Phase1Temp
	case fieldPhase2Temp:
		return m.config.Timer.Phase2Temp
	case fieldPhase3Temp:
		return m.config.Timer.Phase3Temp
	}
	return 0
}

func (m *model) setFieldValue(val int) {
	switch m.selectedField {
	case fieldPhase1Duration:
		m.config.Timer.Phase1Duration = val
	case fieldPhase2Duration:
		m.config.Timer.Phase2Duration = val
	case fieldPhase3Duration:
		m.config.Timer.Phase3Duration = val
	case fieldPhase1Temp:
		m.config.Timer.Phase1Temp = val
	case fieldPhase2Temp:
		m.config.Timer.Phase2Temp = val
	case fieldPhase3Temp:
		m.config.Timer.Phase3Temp = val
	}
}

func (m model) parseInput() int {
	var val int
	if _, err := fmt.Sscanf(m.inputBuffer, "%d", &val); err == nil {
		return val
	}
	return -1
}

func (m model) renderConfigView() string {
	var output strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	editingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)

	output.WriteString("\n")
	output.WriteString(utilities.CenterText(titleStyle.Render("Configuration"), m.width))
	output.WriteString("\n\n")

	fields := []struct {
		name string
		field configField
		unit string
	}{
		{"Phase 1 Duration", fieldPhase1Duration, " minutes"},
		{"Phase 2 Duration", fieldPhase2Duration, " minutes"},
		{"Phase 3 Duration", fieldPhase3Duration, " minutes"},
		{"Phase 1 Temperature", fieldPhase1Temp, "°"},
		{"Phase 2 Temperature", fieldPhase2Temp, "°"},
		{"Phase 3 Temperatuer", fieldPhase3Temp, "°"},
	}

	for _, f := range fields {
		var line string
		value := m.getFieldValue()

		if f.field == m.selectedField {
			if m.editingField {
				displayValue := m.inputBuffer
				if displayValue == "" {
					displayValue = "_"
				}
				line = fmt.Sprintf("  ▶ %s: %s%s", f.name, displayValue, f.unit)
				line = editingStyle.Render(line)
			} else {
				line = fmt.Sprintf("  ▶ %s: %d%s", f.name, value, f.unit)
                line = selectedStyle.Render(line)
			}
		} else {
			switch f.field{
			case fieldPhase1Duration:
				value = m.config.Timer.Phase1Duration
			case fieldPhase2Duration:
				value = m.config.Timer.Phase2Duration
			case fieldPhase3Duration:
				value = m.config.Timer.Phase3Duration
			case fieldPhase1Temp:
				value = m.config.Timer.Phase1Temp
			case fieldPhase2Temp:
				value = m.config.Timer.Phase2Temp
			case fieldPhase3Temp:
				value = m.config.Timer.Phase3Temp
			}
			line = fmt.Sprintf("    %s: %d%s", f.name, value, f.unit)
			line = normalStyle.Render(line)
		}

		output.WriteString(utilities.CenterText(line, m.width))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	helpText := "↑/↓: Navigate | Enter: Edit | Esc/q/?: Exit"
    if m.editingField {
        helpText = "Type value | Enter: Save | Esc: Cancel"
    }
	
	output.WriteString(utilities.CenterText(normalStyle.Render(helpText), m.width))
	output.WriteString("\n")
	versionText := "v1.0.5"
	output.WriteString(utilities.CenterText(normalStyle.Render(versionText), m.width))

	return output.String()
}