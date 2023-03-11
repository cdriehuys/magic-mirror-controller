package magicmirror

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type DisplayState struct {
	On bool `json:"on"`
}

type SharedDisplayState struct {
	sync.Mutex
	state DisplayState
}

func (s *SharedDisplayState) State() DisplayState {
	s.Lock()
	defer s.Unlock()

	return s.state
}

func (s *SharedDisplayState) SetState(new DisplayState) {
	s.Lock()
	defer s.Unlock()

	s.state = new
}

func TurnOff(ctx context.Context, config Config) error {
	cmd := exec.CommandContext(
		ctx,
		"xrandr",
		"--display", config.DisplayIdentifier,
		"--output", config.OutputIdentifier,
		"--off",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not turn off display: %v", err)
	}

	log.Println("Turned off display.")

	return nil
}

type rotationIdentifier string

const (
	RotationNormal   rotationIdentifier = "normal"
	RotationInverted rotationIdentifier = "inverted"
	RotationLeft     rotationIdentifier = "left"
	RotationRight    rotationIdentifier = "right"
)

func TurnOn(ctx context.Context, config Config) error {
	cmd := exec.CommandContext(
		ctx,
		"xrandr",
		"--display", config.DisplayIdentifier,
		"--output", config.OutputIdentifier,
		"--auto",
		"--rotate", string(config.Rotation),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not turn on display: %v", err)
	}

	log.Println("Turned on display.")

	return nil
}

func Refresh(ctx context.Context, config Config) error {
	findWindowCmd := exec.CommandContext(
		ctx,
		"xdotool",
		"search",
		"--name", config.WindowName,
	)
	findWindowCmd.Env = append(findWindowCmd.Env, fmt.Sprintf("DISPLAY=%s", config.DisplayIdentifier))

	windowNameOutput, err := findWindowCmd.Output()
	if err != nil {
		return fmt.Errorf("could not find a window named %q: %v", config.WindowName, err)
	}

	windowNames := strings.Split("\n", strings.TrimSpace(string(windowNameOutput)))
	if len(windowNames) != 1 {
		return fmt.Errorf("expected to find exactly one window identifier, but got: %s", strings.Join(windowNames, ", "))
	}

	targetWindow := strings.TrimSpace(windowNames[0])

	refreshCmd := exec.CommandContext(
		ctx,
		"xdotool",
		"key",
		"--window", targetWindow,
		"F5",
	)
	refreshCmd.Env = append(refreshCmd.Env, fmt.Sprintf("DISPLAY=%s", config.DisplayIdentifier))

	if err := refreshCmd.Run(); err != nil {
		return fmt.Errorf("failed to send refresh command: %v", err)
	}

	log.Println("Refreshed mirror window.")

	return nil
}
