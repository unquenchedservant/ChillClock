package timer

import (
	"fmt"
	"sync"
	"time"

	"github.com/unquenchedservant/ChillClock/config"
)

type Phase int

const (
	PhaseNotStarted Phase = iota
	Phase1
	Phase2
	Phase3
	PhaseCompleted
)

type State struct {
	Text            string `json:"text"`
	Class           string `json:"class"`
	Phase           Phase  `json:"phase"`
	Running         bool   `json:"running"`
	RemainingString string `json:"remainingString"`
	RemainingInt    int64  `json:"remainingInt"`
	Temp            int    `json:"temp"`
	Elapsed         int64  `json:"elapsed"`
	ActiveTimer     int    `json:"activeTimer"`
}

type Timer struct {
	mu           sync.Mutex
	running      bool
	startTime    time.Time
	elapsed      time.Duration
	currentPhase Phase
	activeTimer  int
	config       config.Config
	temp         int
}

func New(cfg config.Config) *Timer {
	return &Timer{
		config:       cfg,
		currentPhase: PhaseNotStarted,
	}
}

func (t *Timer) Toggle(which int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		t.running = true
		t.startTime = time.Now()
		t.elapsed = 0
		t.currentPhase = Phase1
		t.activeTimer = which
	} else {
		t.running = false
		t.elapsed = 0
		t.currentPhase = PhaseNotStarted
	}
}

func (t *Timer) Switch() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running {
		switch t.activeTimer {
		case 1:
			t.activeTimer = 2
		case 2:
			t.activeTimer = 1
		}
	}
}

func (t *Timer) GetConfig() config.Config {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.config
}

func (t *Timer) UpdateConfig(cfg config.Config) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config = cfg
}

func (t *Timer) Tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return
	}

	t.elapsed = time.Since(t.startTime)

	var p1, p2, p3 time.Duration
	if t.activeTimer == 1 {
		p1 = time.Duration(t.config.Timer.Phase1Duration_Timer1) * time.Minute
		p2 = time.Duration(t.config.Timer.Phase2Duration_Timer1) * time.Minute
		p3 = time.Duration(t.config.Timer.Phase3Duration_Timer1) * time.Minute
	} else {
		p1 = time.Duration(t.config.Timer.Phase1Duration_Timer2) * time.Minute
		p2 = time.Duration(t.config.Timer.Phase2Duration_Timer2) * time.Minute
		p3 = time.Duration(t.config.Timer.Phase3Duration_Timer2) * time.Minute
	}

	total := p1 + p2 + p3
	switch {
	case t.elapsed >= total:
		t.currentPhase = PhaseCompleted
		t.running = false
	case t.elapsed >= p1+p2:
		t.currentPhase = Phase3
	case t.elapsed >= p1:
		t.currentPhase = Phase2
	default:
		t.currentPhase = Phase1
	}
}

func (t *Timer) State() State {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running && (t.currentPhase == PhaseNotStarted || t.currentPhase == PhaseCompleted) {
		return State{Text: "0:00", Class: "white", Phase: t.currentPhase}
	}

	var p1, p2, p3 time.Duration
	if t.activeTimer == 1 {
		p1 = time.Duration(t.config.Timer.Phase1Duration_Timer1) * time.Minute
		p2 = time.Duration(t.config.Timer.Phase2Duration_Timer1) * time.Minute
		p3 = time.Duration(t.config.Timer.Phase3Duration_Timer1) * time.Minute
	} else {
		p1 = time.Duration(t.config.Timer.Phase1Duration_Timer2) * time.Minute
		p2 = time.Duration(t.config.Timer.Phase2Duration_Timer2) * time.Minute
		p3 = time.Duration(t.config.Timer.Phase3Duration_Timer2) * time.Minute
	}

	total := p1 + p2 + p3
	remaining := total - t.elapsed

	rm := int(remaining.Minutes())
	rs := int(remaining.Seconds()) % 60

	minutes := int(t.elapsed.Minutes())
	seconds := int(t.elapsed.Seconds()) % 60

	var class string
	switch t.currentPhase {
	case Phase1:
		class = "green"
	case Phase2:
		class = "yellow"
	case Phase3:
		class = "red"
	default:
		class = "white"
	}

	var temp int
	if t.activeTimer == 1 {
		switch t.currentPhase {
		case Phase1:
			temp = t.config.Timer.Phase1Temp_Timer1
		case Phase2:
			temp = t.config.Timer.Phase2Temp_Timer1
		case Phase3:
			temp = t.config.Timer.Phase3Temp_Timer1
		}
	} else {
		switch t.currentPhase {
		case Phase1:
			temp = t.config.Timer.Phase1Temp_Timer2
		case Phase2:
			temp = t.config.Timer.Phase2Temp_Timer2
		case Phase3:
			temp = t.config.Timer.Phase3Temp_Timer2
		}
	}

	return State{
		Text:            fmt.Sprintf("%d:%02d", minutes, seconds),
		Class:           class,
		Phase:           t.currentPhase,
		Running:         t.running,
		RemainingInt:    int64(remaining.Milliseconds()),
		Elapsed:         int64(t.elapsed.Milliseconds()),
		RemainingString: fmt.Sprintf("%d:%02d", rm, rs),
		Temp:            temp,
		ActiveTimer:     t.activeTimer,
	}
}
