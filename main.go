package main

import (
	"context"
	"fmt"
	"log"
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

func main() {
	if err := turnOffDisplay(context.Background()); err != nil {
		log.Fatalln("Error turning off display:", err)
	}
	time.Sleep(10 * time.Second)

	if err := turnOnDisplay(context.Background()); err != nil {
		log.Fatalln("Error turning on display:", err)
	}
}
