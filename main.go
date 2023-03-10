package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

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

type displayState struct {
	On bool `json:"on"`
}

func displayStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

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

	if err := json.NewEncoder(w).Encode(newState); err != nil {
		log.Println("Failed to write display state response:", err)
	}
}

func main() {
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
