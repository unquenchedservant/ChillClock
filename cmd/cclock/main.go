package main

import (
	"fmt"
	"os"
	"time"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/unquenchedservant/ChillClock/config"
)
var version = getVersion()

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
	}
	return "Didn't Work"
}

type model struct {
	width         int
	height        int
	config        config.Config
	timerRunning  bool
	timerStart    time.Time
	timerElapsed  time.Duration
	currentPhase  timerPhase
	timer         int
	timerDefault  int
	configPage    int
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
	CFG_PAGE_1 = 0
	CFG_PAGE_2 = 1
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
		timerDefault: TIMER_1,
		configPage: CFG_PAGE_1,
	}

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
