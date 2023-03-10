package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type displayState struct {
	On bool `json:"on"`
}

type sharedDisplayState struct {
	sync.Mutex
	state displayState
}

func (s *sharedDisplayState) Get() displayState {
	s.Lock()
	current := s.state
	s.Unlock()

	return current
}

func (s *sharedDisplayState) Set(new displayState) {
	s.Lock()
	s.state = new
	s.Unlock()
}

var currentDisplayState sharedDisplayState

func turnOffDisplay(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "xrandr", "--display", ":0.0", "--output", "HDMI-1", "--off")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not turn off display: %v", err)
	}

	log.Println("Turned off display.")

	return nil
}

func turnOnDisplay(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "xrandr", "--display", ":0.0", "--output", "HDMI-1", "--auto", "--rotate", "left")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not turn on display: %v", err)
	}

	log.Println("Turned on display.")

	return nil
}

func displayStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getDisplayState(w, r)
	} else if r.Method == http.MethodPut {
		updateDisplayState(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getDisplayState(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(currentDisplayState.Get()); err != nil {
		log.Println("Failed to encode current display state:", err)
	}
}

func updateDisplayState(w http.ResponseWriter, r *http.Request) {

	var newState displayState
	if err := json.NewDecoder(r.Body).Decode(&newState); err != nil {
		log.Println("Could not decode new display state:", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var updateFunc func(context.Context) error
	if newState.On {
		updateFunc = turnOnDisplay
	} else {
		updateFunc = turnOffDisplay
	}

	if err := updateFunc(r.Context()); err != nil {
		log.Println("Failed to update display state:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currentDisplayState.Set(newState)

	if err := json.NewEncoder(w).Encode(newState); err != nil {
		log.Println("Failed to write display state response:", err)
	}
}

func main() {
	// We assume that the display starts in the "on" state. This is easier than
	// parsing the `xrandr` output and will be mostly correct. It also gets
	// fixed after the first display state update.
	currentDisplayState.Set(displayState{On: true})

	handler := http.NewServeMux()
	handler.HandleFunc("/display-state", displayStateHandler)

	s := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting HTTP server")
	if err := s.ListenAndServe(); err != nil {
		log.Fatalln("HTTP server failed:", err)
	}
}
