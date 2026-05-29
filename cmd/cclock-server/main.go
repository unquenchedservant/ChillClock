package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/unquenchedservant/ChillClock/config"
	"github.com/unquenchedservant/ChillClock/internal/timer"
)

func main() {
	if err := config.EnsureConfigExists(); err != nil {
		panic(err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	t := timer.New(cfg)

	go func() {
		ticker := time.NewTicker(time.Second / 10)
		defer ticker.Stop()
		for range ticker.C {
			t.Tick()
		}
	}()
	http.HandleFunc("/config", handleConfig(t))
	http.HandleFunc("/status", handleStatus(t))
	http.HandleFunc("/toggle", handleToggle(t))
	http.HandleFunc("/events", handleSSE((t)))

	fmt.Println("ChillClock server running on 2420")
	http.ListenAndServe(":2420", nil)
}

func handleConfig(t *timer.Timer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(t.GetConfig())

		case http.MethodPut:
			var cfg config.Config
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				http.Error(w, "Invalid config", http.StatusBadRequest)
				return
			}
			t.UpdateConfig(cfg)
			config.SaveConfig(cfg)
			json.NewEncoder(w).Encode(cfg)
		case http.MethodPatch:
			var patch map[string]int
			if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			cfg := t.GetConfig()
			for key, val := range patch {
				switch key {
				case "default_timer":
					cfg.Timer.DefaultTimer = val
				case "phase1_timer1_duration_minutes":
					cfg.Timer.Phase1Duration_Timer1 = val
				case "phase2_timer1_duration_minutes":
					cfg.Timer.Phase2Duration_Timer1 = val
				case "phase3_timer1_duration_minutes":
					cfg.Timer.Phase3Duration_Timer1 = val
				case "phase1_timer2_duration_minutes":
					cfg.Timer.Phase1Duration_Timer2 = val
				case "phase2_timer2_duration_minutes":
					cfg.Timer.Phase2Duration_Timer2 = val
				case "phase3_timer2_duration_minutes":
					cfg.Timer.Phase3Duration_Timer2 = val
				case "phase1_timer1_temp":
					cfg.Timer.Phase1Temp_Timer1 = val
				case "phase2_timer1_temp":
					cfg.Timer.Phase2Temp_Timer1 = val
				case "phase3_timer1_temp":
					cfg.Timer.Phase3Temp_Timer1 = val
				case "phase1_timer2_temp":
					cfg.Timer.Phase1Temp_Timer2 = val
				case "phase2_timer2_temp":
					cfg.Timer.Phase2Temp_Timer2 = val
				case "phase3_timer2_temp":
					cfg.Timer.Phase3Temp_Timer2 = val
				}
			}
			t.UpdateConfig(cfg)
			config.SaveConfig(cfg)
			json.NewEncoder(w).Encode(t.GetConfig())
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleStatus(t *timer.Timer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t.State())
	}
}

func handleToggle(t *timer.Timer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		which, err := strconv.Atoi(r.URL.Query().Get("timer"))
		if err != nil || (which != 1 && which != 2) {
			http.Error(w, "timer must be 1 or 2", http.StatusBadRequest)
			return
		}
		t.Toggle(which)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t.State())
	}
}

func handleSSE(t *timer.Timer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ticker := time.NewTicker(time.Second / 10)
		defer ticker.Stop()
		for range ticker.C {
			data, _ := json.Marshal(t.State())
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		}
	}
}
