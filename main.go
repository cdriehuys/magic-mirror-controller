package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	magicmirror "github.com/cdriehuys/magic-mirror-controller/internal"
)

func createDisplayStateHandler(config magicmirror.Config, state *magicmirror.SharedDisplayState) http.HandlerFunc {
	getDisplayState := createGetDisplayStateHandler(state)
	updateDisplayState := createUpdateDisplayState(config, state)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getDisplayState(w, r)
		} else if r.Method == http.MethodPut {
			updateDisplayState(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func createGetDisplayStateHandler(state *magicmirror.SharedDisplayState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(state.State()); err != nil {
			log.Println("Failed to encode current display state:", err)
		}
	}
}

func createUpdateDisplayState(config magicmirror.Config, state *magicmirror.SharedDisplayState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newState magicmirror.DisplayState
		if err := json.NewDecoder(r.Body).Decode(&newState); err != nil {
			log.Println("Could not decode new display state:", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		var updateFunc func(context.Context, magicmirror.Config) error
		if newState.On {
			updateFunc = magicmirror.TurnOn
		} else {
			updateFunc = magicmirror.TurnOff
		}

		if err := updateFunc(r.Context(), config); err != nil {
			log.Println("Failed to update display state:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		state.SetState(newState)

		if err := json.NewEncoder(w).Encode(newState); err != nil {
			log.Println("Failed to write display state response:", err)
		}
	}
}

func createRefreshHandler(config magicmirror.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := magicmirror.Refresh(r.Context(), config); err != nil {
			log.Println("Failed to refresh display:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	// We assume that the display starts in the "on" state. This is easier than
	// parsing the `xrandr` output and will be mostly correct. It also gets
	// fixed after the first display state update.
	var sharedDisplayState magicmirror.SharedDisplayState
	sharedDisplayState.SetState(magicmirror.DisplayState{On: true})

	config := magicmirror.Config{
		DisplayIdentifier: ":0.0",
		OutputIdentifier:  "HDMI-1",
		Rotation:          magicmirror.RotationLeft,
		WindowName:        "Mozilla Firefox",
		RefreshKey:        "F5",
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/display-state", createDisplayStateHandler(config, &sharedDisplayState))
	handler.HandleFunc("/refresh", createRefreshHandler(config))

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
