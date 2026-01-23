package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/unquenchedservant/ChillClock/config"
	util "github.com/unquenchedservant/ChillClock/utilities"
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
			minField := fieldPhase1DurationT1
			if m.configPage == CFG_PAGE_2 {
				minField = fieldPhase1DurationT2
			}
			if m.selectedField > minField {
				m.selectedField--
			}
		case "down", "j":
			m.saveAndExitField()
			maxField := fieldPhase3TempT1
			if m.configPage == CFG_PAGE_2 {
				maxField = fieldPhase3TempT2
			}
			if m.selectedField < maxField {
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
			minField := fieldPhase1DurationT1
			if m.configPage == CFG_PAGE_2 {
				minField = fieldPhase1DurationT2
			}
			if m.selectedField > minField {
				m.selectedField--
			}
		case "left", "h":
			if m.configPage == CFG_PAGE_2 {
				m.selectedField -= 6
			}
			m.configPage = CFG_PAGE_1
		case "right", "l":
			if m.configPage == CFG_PAGE_1 {
				m.selectedField += 6
			}
			m.configPage = CFG_PAGE_2
		case "down", "j":
			maxField := fieldPhase3TempT1
			if m.configPage == CFG_PAGE_2 {
				maxField = fieldPhase3TempT2
			}
			if m.selectedField < maxField {
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
	case fieldPhase1DurationT1:
		return m.config.Timer.Phase1Duration_Timer1
	case fieldPhase2DurationT1:
		return m.config.Timer.Phase2Duration_Timer1
	case fieldPhase3DurationT1:
		return m.config.Timer.Phase3Duration_Timer1
	case fieldPhase1TempT1:
		return m.config.Timer.Phase1Temp_Timer1
	case fieldPhase2TempT1:
		return m.config.Timer.Phase2Temp_Timer1
	case fieldPhase3TempT1:
		return m.config.Timer.Phase3Temp_Timer1
	case fieldPhase1DurationT2:
		return m.config.Timer.Phase1Duration_Timer2
	case fieldPhase2DurationT2:
		return m.config.Timer.Phase2Duration_Timer2
	case fieldPhase3DurationT2:
		return m.config.Timer.Phase3Duration_Timer2
	case fieldPhase1TempT2:
		return m.config.Timer.Phase1Temp_Timer2
	case fieldPhase2TempT2:
		return m.config.Timer.Phase2Temp_Timer2
	case fieldPhase3TempT2:
		return m.config.Timer.Phase3Temp_Timer2
	}
	return 0
}

func (m *model) setFieldValue(val int) {
	switch m.selectedField {
	case fieldPhase1DurationT1:
		m.config.Timer.Phase1Duration_Timer1 = val
	case fieldPhase2DurationT1:
		m.config.Timer.Phase2Duration_Timer1 = val
	case fieldPhase3DurationT1:
		m.config.Timer.Phase3Duration_Timer1 = val
	case fieldPhase1TempT1:
		m.config.Timer.Phase1Temp_Timer1 = val
	case fieldPhase2TempT1:
		m.config.Timer.Phase2Temp_Timer1 = val
	case fieldPhase3TempT1:
		m.config.Timer.Phase3Temp_Timer1 = val
	case fieldPhase1DurationT2:
		m.config.Timer.Phase1Duration_Timer2 = val
	case fieldPhase2DurationT2:
		m.config.Timer.Phase2Duration_Timer2 = val
	case fieldPhase3DurationT2:
		m.config.Timer.Phase3Duration_Timer2 = val
	case fieldPhase1TempT2:
		m.config.Timer.Phase1Temp_Timer2 = val
	case fieldPhase2TempT2:
		m.config.Timer.Phase2Temp_Timer2 = val
	case fieldPhase3TempT2:
		m.config.Timer.Phase3Temp_Timer2 = val
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
	fields := []struct {
		name string 
		field configField
		unit string
	}{}
	minField := fieldPhase1DurationT1
	maxField := fieldPhase3TempT1
	if m.configPage == CFG_PAGE_2 {
		minField = fieldPhase1DurationT2
		maxField = fieldPhase3TempT2
	}
	output.WriteString("\n")
	output.WriteString(util.CenterText(util.GetYellowStyle().Bold(true).Render(fmt.Sprintf("    Timer %d Configuration", m.configPage+1)), m.width))
	output.WriteString("\n\n")
	if m.configPage == CFG_PAGE_1 {
		fields = []struct {
			name string
			field configField
			unit string
		}{
			{"Phase 1 Duration", fieldPhase1DurationT1, " minutes"},
			{"Phase 2 Duration", fieldPhase2DurationT1, " minutes"},
			{"Phase 3 Duration", fieldPhase3DurationT1, " minutes"},
			{"Phase 1 Temperature", fieldPhase1TempT1, "°"},
			{"Phase 2 Temperature", fieldPhase2TempT1, "°"},
			{"Phase 3 Temperature", fieldPhase3TempT1, "°"},
		}
	} else {
		fields = []struct {
			name string
			field configField
			unit string
		}{
			{"Phase 1 Duration", fieldPhase1DurationT2, " minutes"},
			{"Phase 2 Duration", fieldPhase2DurationT2, " minutes"},
			{"Phase 3 Duration", fieldPhase3DurationT2, " minutes"},
			{"Phase 1 Temperature", fieldPhase1TempT2, "°"},
			{"Phase 2 Temperature", fieldPhase2TempT2, "°"},
			{"Phase 3 Temperature", fieldPhase3TempT2, "°"},
		}
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
				line = util.GetEditingStyle().Render(line)
			} else {
				line = fmt.Sprintf("  ▶ %s: %d%s", f.name, value, f.unit)
                line = util.GetGreenStyle().Bold(true).Render(line)
			}
		} else {
			switch f.field{
			case fieldPhase1DurationT1:
				value = m.config.Timer.Phase1Duration_Timer1
			case fieldPhase2DurationT1:
				value = m.config.Timer.Phase2Duration_Timer1
			case fieldPhase3DurationT1:
				value = m.config.Timer.Phase3Duration_Timer1
			case fieldPhase1TempT1:
				value = m.config.Timer.Phase1Temp_Timer1
			case fieldPhase2TempT1:
				value = m.config.Timer.Phase2Temp_Timer1
			case fieldPhase3TempT1:
				value = m.config.Timer.Phase3Temp_Timer1
			case fieldPhase1DurationT2:
				value = m.config.Timer.Phase1Duration_Timer2
			case fieldPhase2DurationT2:
				value = m.config.Timer.Phase2Duration_Timer2
			case fieldPhase3DurationT2:
				value = m.config.Timer.Phase3Duration_Timer2
			case fieldPhase1TempT2:
				value = m.config.Timer.Phase1Temp_Timer2
			case fieldPhase2TempT2:
				value = m.config.Timer.Phase2Temp_Timer2
			case fieldPhase3TempT2:
				value = m.config.Timer.Phase3Temp_Timer2
			}
			line = fmt.Sprintf("    %s: %d%s", f.name, value, f.unit)
			line = util.GetNormalStyle().Render(line)
		}

		output.WriteString(util.CenterText(line, m.width))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	navigate_page := ""
	up_down := ""
	if m.configPage == CFG_PAGE_1 {
		navigate_page = "→: Next Timer | "
	}else if m.configPage == CFG_PAGE_2 {
		navigate_page = "←: Prv. Timer | "
	}
	if m.selectedField == minField {
		up_down = "↓: Navigate | "
	}else if m.selectedField == maxField {
		up_down ="↑: Navigate | "
	}else {
		up_down = "↑/↓: Navigate | "
	}
	helpText := fmt.Sprintf("%s%sEnter: Edit | Esc/q/?: Exit", navigate_page, up_down)
    if m.editingField {
        helpText = "Type value | Enter: Save | Esc: Cancel"
    }
	
	output.WriteString(util.CenterText(util.GetNormalStyle().Render(helpText), m.width))
	output.WriteString("\n")
	versionText := version
	output.WriteString(util.CenterText(util.GetNormalStyle().Render(versionText), m.width))

	return output.String()
}