package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/unquenchedservant/ChillClock/config"
	"github.com/unquenchedservant/ChillClock/internal/timer"
)

func main() {
	ssl := false
	getKey := flag.Bool("get-key", false, "Display the current API Key (generates one if none exists)")
	newKey := flag.Bool("new-key", false, "Generate and display a new API key")
	certFile := flag.String("cert", "", "Path to TLS Cert (Required w/ --key for https)")
	keyFile := flag.String("key", "", "Path to TLS Key File (Required w/ --cert for https)")
	flag.Parse()

	if *newKey {
		key, err := config.GenerateAPIKey()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to generate key:", err)
			os.Exit(1)
		}
		if err := config.SaveAPIKey(key); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save key:", err)
			os.Exit(1)
		}
		fmt.Println(key)
		return
	}

	if *getKey {
		key, err := config.LoadAPIKey()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to load key:", err)
			os.Exit(1)
		}
		if key == "" {
			key, err = config.GenerateAPIKey()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to generate key:", err)
				os.Exit(1)
			}
			if err := config.SaveAPIKey(key); err != nil {
				fmt.Fprintln(os.Stderr, "Failed to save key:", err)
				os.Exit(1)
			}
		}
		fmt.Println(key)
		return
	}

	if *certFile != "" || *keyFile != "" {
		if *certFile == "" || *keyFile == "" {
			fmt.Fprintln(os.Stderr, "both -cert and -key must be provided together")
			os.Exit(1)
		}
		if _, err := os.Stat(*certFile); err != nil {
			fmt.Fprintf(os.Stderr, "cert file not found: %s\n", *certFile)
			os.Exit(1)
		}
		if _, err := os.Stat(*keyFile); err != nil {
			fmt.Fprintf(os.Stderr, "key file not found: %s\n", *keyFile)
			os.Exit(1)
		}
		ssl = true
	}

	if err := config.EnsureConfigExists(); err != nil {
		panic(err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	apiKey, err := config.LoadAPIKey()
	if err != nil {
		panic(err)
	}
	if apiKey == "" {
		fmt.Println("Warning: No API key set. Run --get-key to enable auth.")
	}
	t := timer.New(cfg)

	go func() {
		ticker := time.NewTicker(time.Second / 10)
		defer ticker.Stop()
		for range ticker.C {
			t.Tick()
		}
	}()

	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		if apiKey == "" {
			return h
		}
		return requireAPIKey(apiKey, h)
	}

	http.HandleFunc("/config", wrap(handleConfig(t)))
	http.HandleFunc("/status", wrap(handleStatus(t)))
	http.HandleFunc("/toggle", wrap(handleToggle(t)))
	http.HandleFunc("/switch", wrap(handleSwitch(t)))
	http.HandleFunc("/events", wrap(handleSSE((t))))

	fmt.Println("ChillClock server running on 2420")
	if ssl {
		fmt.Println("SSL enabled")
		if err := http.ListenAndServeTLS(":2420", *certFile, *keyFile, nil); err != nil {
			fmt.Fprintln(os.Stderr, "ListenAndServeTLS:", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("SSL not enabled")
		if err := http.ListenAndServe(":2420", nil); err != nil {
			fmt.Fprintln(os.Stderr, "ListenAndServe:", err)
			os.Exit(1)
		}
	}
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

func handleSwitch(t *timer.Timer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		t.Switch()
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

func requireAPIKey(apiKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
