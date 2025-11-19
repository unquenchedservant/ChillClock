package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/unquenchedservant/ChillClock/config"
)

// Digit definitions - each digit is 7 runes wide + 1 space trailing = 8 total
var digits = map[rune][]string{
	'0': {
		" ██████ ",
		" ██  ██ ",
		" ██  ██ ",
		" ██  ██ ",
		" ██████ ",
	},
	'1': {
		" ████   ",
		"   ██   ",
		"   ██   ",
		"   ██   ",
		" ██████ ",
	},
	'2': {
		" ██████ ",
		"     ██ ",
		" ██████ ",
		" ██     ",
		" ██████ ",
	},
	'3': {
		" ██████ ",
		"     ██ ",
		" ██████ ",
		"     ██ ",
		" ██████ ",
	},
	'4': {
		" ██   ██ ",
		" ██   ██ ",
		" ███████ ",
		"      ██ ",
		"      ██ ",
	},
	'5': {
		" ███████ ",
		" ██      ",
		" ███████ ",
		"      ██ ",
		" ███████ ",
	},
	'6': {
		" ███████ ",
		" ██      ",
		" ███████ ",
		" ██   ██ ",
		" ███████ ",
	},
	'7': {
		" ███████ ",
		"      ██ ",
		"      ██ ",
		"      ██ ",
		"      ██ ",
	},
	'8': {
		" ███████ ",
		" ██   ██ ",
		" ███████ ",
		" ██   ██ ",
		" ███████ ",
	},
	'9': {
		" ███████ ",
		" ██   ██ ",
		" ███████ ",
		"      ██ ",
		" ███████ ",
	},
	':': {
		"      ",
		"  ██  ",
		"      ",
		"  ██  ",
		"      ",
	},
	' ': {
		"     ",
		"     ",
		"     ",
		"     ",
		"     ",
	},
}

type timerPhase int

const (
	phaseNotStarted timerPhase = iota
	phase1
	phase2
	phase3
	phaseCompleted
)

// TimerOutput represents the JSON structure for the timer state file
type TimerOutput struct {
	Text  string `json:"text"`
	Class string `json:"class"`
}

type viewMode int

const (
	viewClock viewMode = iota
	viewConfig
)

type configField int

const (
	fieldPhase1Duration configField = iota
	fieldPhase2Duration
	fieldPhase3Duration
	fieldPhase1Temp
	fieldPhase2Temp
	fieldPhase3Temp
	fieldMax
)

type model struct {
	width           int
	height          int
	config          config.Config
	timerRunning    bool
	timerStart      time.Time
	timerElapsed    time.Duration
	currentPhase    timerPhase
	lastPhase       timerPhase // Track last phase for ding detection
	mode            viewMode
	selectedField   configField
	editingField    bool
	inputBuffer     string
	previousValue   int // Store previous value to restore if input is blank
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type dingMsg struct{}

func dingCmd(phase timerPhase, temp int) tea.Cmd {
	return func() tea.Msg {
		playBeep()
		sendNotification(phase, temp)
		return dingMsg{}
	}
}

type fileClickMsg struct{}

func watchForFileClick() tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil
		}

		clickFile := filepath.Join(homeDir, "dhv_timer_click1")

		// Check if file exists
		if _, err := os.Stat(clickFile); err == nil {
			// File exists, delete it and return click message
			os.Remove(clickFile)
			return fileClickMsg{}
		}

		return nil
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), watchForFileClick(), tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle config mode
		if m.mode == viewConfig {
			return m.handleConfigInput(msg)
		}

		// Handle clock mode
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
		case "enter", " ":
			// Toggle timer
			if !m.timerRunning {
				m.timerRunning = true
				m.timerStart = time.Now()
				m.timerElapsed = 0
				m.currentPhase = phase1
				m.lastPhase = phaseNotStarted
			} else {
				// Reset timer
				m.timerRunning = false
				m.timerElapsed = 0
				m.currentPhase = phaseNotStarted
				m.lastPhase = phaseNotStarted
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case fileClickMsg:
		// Toggle timer (same as Enter/Space)
		if !m.timerRunning {
			m.timerRunning = true
			m.timerStart = time.Now()
			m.timerElapsed = 0
			m.currentPhase = phase1
			m.lastPhase = phaseNotStarted
		} else {
			// Reset timer
			m.timerRunning = false
			m.timerElapsed = 0
			m.currentPhase = phaseNotStarted
			m.lastPhase = phaseNotStarted
		}
		return m, watchForFileClick()
	case tickMsg:
		if m.timerRunning {
			m.timerElapsed = time.Since(m.timerStart)

			// Calculate total duration and determine phase
			phase1Dur := time.Duration(m.config.Timer.Phase1Duration) * time.Minute
			phase2Dur := time.Duration(m.config.Timer.Phase2Duration) * time.Minute
			phase3Dur := time.Duration(m.config.Timer.Phase3Duration) * time.Minute
			totalDur := phase1Dur + phase2Dur + phase3Dur

			// Determine current phase
			oldPhase := m.currentPhase
			if m.timerElapsed >= totalDur {
				m.currentPhase = phaseCompleted
				m.timerRunning = false
			} else if m.timerElapsed >= phase1Dur+phase2Dur {
				m.currentPhase = phase3
			} else if m.timerElapsed >= phase1Dur {
				m.currentPhase = phase2
			} else {
				m.currentPhase = phase1
			}

			// Write timer state to file
			writeTimerState(m)

			// Check for phase transition and play ding with notification
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
					temp = 0 // Not used for completion
				}
				return m, tea.Batch(tickCmd(), dingCmd(m.currentPhase, temp))
			}
		} else {
			// Write empty state when timer is not running
			writeTimerState(m)
		}
		return m, tea.Batch(tickCmd(), watchForFileClick())
	case dingMsg:
		// Ding completed, nothing to do
		return m, nil
	}
	return m, nil
}

func (m model) handleConfigInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingField {
		// Handle input while editing
		switch msg.String() {
		case "enter":
			// Save the value if valid, otherwise restore previous value
			if m.inputBuffer == "" {
				// Restore previous value if input is blank
				m.setFieldValue(m.previousValue)
			} else if val := m.parseInput(); val >= 0 {
				m.setFieldValue(val)
				config.SaveConfig(m.config)
			} else {
				// Invalid input, restore previous value
				m.setFieldValue(m.previousValue)
			}
			m.editingField = false
			m.inputBuffer = ""
		case "esc":
			// Cancel editing and restore previous value
			m.setFieldValue(m.previousValue)
			m.editingField = false
			m.inputBuffer = ""
		case "backspace":
			if len(m.inputBuffer) > 0 {
				m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
			}
		case "up", "k":
			// Moving away from field - restore previous value if blank
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
			if m.selectedField > 0 {
				m.selectedField--
			}
		case "down", "j":
			// Moving away from field - restore previous value if blank
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
			if m.selectedField < fieldMax-1 {
				m.selectedField++
			}
		default:
			// Only accept digits
			if len(msg.String()) == 1 && msg.String()[0] >= '0' && msg.String()[0] <= '9' {
				m.inputBuffer += msg.String()
			}
		}
	} else {
		// Navigate fields
		switch msg.String() {
		case "esc", "q", "?":
			m.mode = viewClock
		case "up", "k":
			if m.selectedField > 0 {
				m.selectedField--
			}
		case "down", "j":
			if m.selectedField < fieldMax-1 {
				m.selectedField++
			}
		case "enter", " ":
			// Store the current value and clear the input buffer
			m.previousValue = m.getFieldValue()
			m.editingField = true
			m.inputBuffer = ""
		}
	}
	return m, nil
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

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.mode == viewConfig {
		return m.renderConfigView()
	}

	return m.renderClockView()
}

func (m model) renderConfigView() string {
	var output strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	editingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)

	// Title
	output.WriteString("\n")
	output.WriteString(centerText(titleStyle.Render("Configuration"), m.width))
	output.WriteString("\n\n")

	fields := []struct {
		name  string
		field configField
		unit  string
	}{
		{"Phase 1 Duration", fieldPhase1Duration, " minutes"},
		{"Phase 2 Duration", fieldPhase2Duration, " minutes"},
		{"Phase 3 Duration", fieldPhase3Duration, " minutes"},
		{"Phase 1 Temperature", fieldPhase1Temp, "°"},
		{"Phase 2 Temperature", fieldPhase2Temp, "°"},
		{"Phase 3 Temperature", fieldPhase3Temp, "°"},
	}

	for _, f := range fields {
		var line string
		value := m.getFieldValue()

		if f.field == m.selectedField {
			if m.editingField {
				// Show input buffer or placeholder if empty
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
			if f.field == fieldPhase1Duration {
				value = m.config.Timer.Phase1Duration
			} else if f.field == fieldPhase2Duration {
				value = m.config.Timer.Phase2Duration
			} else if f.field == fieldPhase3Duration {
				value = m.config.Timer.Phase3Duration
			} else if f.field == fieldPhase1Temp {
				value = m.config.Timer.Phase1Temp
			} else if f.field == fieldPhase2Temp {
				value = m.config.Timer.Phase2Temp
			} else if f.field == fieldPhase3Temp {
				value = m.config.Timer.Phase3Temp
			}
			line = fmt.Sprintf("    %s: %d%s", f.name, value, f.unit)
			line = normalStyle.Render(line)
		}

		output.WriteString(centerText(line, m.width))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	helpText := "↑/↓: Navigate | Enter: Edit | Esc/q/?: Exit"
	if m.editingField {
		helpText = "Type value | Enter: Save | Esc: Cancel"
	}
	output.WriteString(centerText(normalStyle.Render(helpText), m.width))

	return output.String()
}

func (m model) renderClockView() string {
	now := time.Now()
	timeStr := now.Format("15:04:05")
	dateStr := now.Format("2006-01-02")

	// Render the clock
	clockLines := renderLargeText(timeStr)

	// Style definitions
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	// Build output
	var output strings.Builder

	// Calculate vertical centering
	totalLines := 1 + len(clockLines) + 3 // date + clock + spacing + timer
	topPadding := (m.height - totalLines) / 2

	// Add top padding
	for i := 0; i < topPadding; i++ {
		output.WriteString("\n")
	}

	// Add centered date in yellow
	output.WriteString(centerText(yellowStyle.Render(dateStr), m.width))
	output.WriteString("\n\n")

	// Add centered clock in green
	for _, line := range clockLines {
		styledLine := greenStyle.Render(line)
		output.WriteString(centerText(styledLine, m.width))
		output.WriteString("\n")
	}

	// Add timer display
	output.WriteString("\n")
	timerText, timerStyle := m.getTimerDisplay()
	output.WriteString(centerText(timerStyle.Render(timerText), m.width))

	return output.String()
}

func (m model) getTimerDisplay() (string, lipgloss.Style) {
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	if !m.timerRunning && m.currentPhase == phaseNotStarted {
		return "Press Enter or Space to start timer, '?' for config", whiteStyle
	}

	if m.currentPhase == phaseCompleted {
		return "Timer completed! Press Enter or Space to restart, '?' for config", whiteStyle
	}

	// Calculate remaining time
	
	phase1Temp := m.config.Timer.Phase1Temp
	phase2Temp := m.config.Timer.Phase2Temp
	phase3Temp := m.config.Timer.Phase3Temp

	elapsed := m.timerElapsed

	// Format time as MM:SS
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60

	timerText := fmt.Sprintf("Timer: %02d:%02d", minutes, seconds)

	// Determine color based on phase
	var style lipgloss.Style
	switch m.currentPhase {
	case phase1:
		style = greenStyle
		timerText += fmt.Sprintf(" Temp: %d°", phase1Temp)
	case phase2:
		style = yellowStyle
		timerText += fmt.Sprintf(" Temp: %d°", phase2Temp)
	case phase3:
		style = redStyle
		timerText += fmt.Sprintf(" Temp: %d°", phase3Temp)
	default:
		style = whiteStyle
	}

	return timerText, style
}

func renderLargeText(text string) []string {
	var lines [5]strings.Builder

	for _, char := range text {
		digitLines, exists := digits[char]
		if !exists {
			digitLines = digits[' ']
		}

		for i, line := range digitLines {
			lines[i].WriteString(line)
		}
	}

	result := make([]string, 5)
	for i, line := range lines {
		result[i] = line.String()
	}

	return result
}

func centerText(text string, width int) string {
	// Use lipgloss width calculation to handle ANSI codes
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}

	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

func playBeep() {
	// Play a system beep sound
	switch runtime.GOOS {
	case "linux":
		// Try paplay (PulseAudio) first, fall back to speaker-test
		cmd := exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
		if err := cmd.Run(); err != nil {
			// Fallback to beep command or speaker-test
			exec.Command("speaker-test", "-t", "sine", "-f", "1000", "-l", "1").Run()
		}
	case "darwin":
		// macOS
		exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()
	case "windows":
		// Windows - use rundll32 to play system sound
		exec.Command("rundll32", "user32.dll,MessageBeep").Run()
	default:
		// Fallback: print bell character
		fmt.Print("\a")
	}
}

func sendNotification(phase timerPhase, temp int) {
	var title, body string

	switch phase {
	case phase1:
		title = "Phase 1"
	case phase2:
		title = "Phase 2"
	case phase3:
		title = "Phase 3"
	case phaseCompleted:
		title = "Timer Complete"
		body = "All phases finished!"
	default:
		return
	}

	if phase != phaseCompleted {
		body = fmt.Sprintf("%d°", temp)
	}

	switch runtime.GOOS {
	case "linux":
		// Use notify-send for desktop notifications
		exec.Command("notify-send", "-u", "normal", "-t", "5000", title, body).Run()
	case "darwin":
		// macOS - use osascript to display notification
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
		exec.Command("osascript", "-e", script).Run()
	case "windows":
		// Windows - use PowerShell to show toast notification
		script := fmt.Sprintf(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; $Template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02); $RawXml = [xml] $Template.GetXml(); ($RawXml.toast.visual.binding.text|where {$_.id -eq "1"}).AppendChild($RawXml.CreateTextNode("%s")) > $null; ($RawXml.toast.visual.binding.text|where {$_.id -eq "2"}).AppendChild($RawXml.CreateTextNode("%s")) > $null; $SerializedXml = New-Object Windows.Data.Xml.Dom.XmlDocument; $SerializedXml.LoadXml($RawXml.OuterXml); $Toast = [Windows.UI.Notifications.ToastNotification]::new($SerializedXml); $Toast.Tag = "ChillClock"; $Toast.Group = "ChillClock"; $Notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("ChillClock"); $Notifier.Show($Toast);`, title, body)
		exec.Command("powershell", "-Command", script).Run()
	}
}

func writeTimerState(m model) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	timerFile := filepath.Join(homeDir, "dhv_timer.txt")

	var output TimerOutput

	if !m.timerRunning && m.currentPhase == phaseNotStarted {
		output = TimerOutput{
			Text:  "0:00",
			Class: "white",
		}
	} else if m.currentPhase == phaseCompleted {
		output = TimerOutput{
			Text:  "0:00",
			Class: "white",
		}
	} else {
		// Format time as M:SS (no leading zero for minutes)
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

		output = TimerOutput{
			Text:  timerText,
			Class: class,
		}
	}

	data, err := json.Marshal(output)
	if err != nil {
		return err
	}

	return os.WriteFile(timerFile, data, 0644)
}

func main() {
	// Ensure config exists and load it
	if err := config.EnsureConfigExists(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create initial model with config
	initialModel := model{
		config:        cfg,
		currentPhase:  phaseNotStarted,
		lastPhase:     phaseNotStarted,
		mode:          viewClock,
		selectedField: fieldPhase1Duration,
		editingField:  false,
		inputBuffer:   "",
		previousValue: 0,
	}

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
