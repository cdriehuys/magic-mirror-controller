# Magic Mirror Controller

A simple REST controller for managing the display of a magic mirror. The API
allows for turning the mirror on and off, as well as querying its current state.

## Building

```bash
GOOS=linux GOARCH=arm64 go build -o build/magic-mirror-controller
```

## Display Power

This program is essentially a wrapper around the following two commands which
turn the display on and off:
```bash
xrandr --display :0.0 --output HDMI-1 --off
xrandr --display :0.0 --output HDMI-1 --auto --rotate left
```
