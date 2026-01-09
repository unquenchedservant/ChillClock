package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/unquenchedservant/ChillClock/config"
)

type model struct {
	width         int
	height        int
	config        config.Config
	timerRunning  bool
	timerStart    time.Time
	timerElapsed  time.Duration
	currentPhase  timerPhase
	timer         int
	lastPhase     timerPhase // Track last phase for ding detection
	mode          viewMode
	selectedField configField
	editingField  bool
	inputBuffer   string
	previousValue int // Store previous value to restore if input is blank
}

const (
	TIMER_1 = 1
	TIMER_2 = 2
	TIMER_DEFAULT = 1
)

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), watchForFileClick(), tea.EnterAltScreen)
}

func main() {
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
		selectedField: fieldPhase1DurationT1,
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
