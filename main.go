package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/gen2brain/beeep"
)

//go:embed sounds/*
var soundFiles embed.FS

func main() {
	// Define command-line flags
	connectSoundPath := flag.String(
		"connect-sound",
		"",
		"Path to custom connect sound file (can be relative or absolute)",
	)
	disconnectSoundPath := flag.String(
		"disconnect-sound",
		"",
		"Path to custom disconnect sound file (can be relative or absolute)",
	)
	installFlag := flag.Bool(
		"install",
		false,
		"Install the program as a systemd user service and exit.",
	)
	uninstallFlag := flag.Bool(
		"uninstall",
		false,
		"Uninstall the systemd user service and exit.",
	)
	nosoundFlag := flag.Bool(
		"nosound",
		false,
		"Disable sounds",
	)
	nonotifFlag := flag.Bool(
		"nonotif",
		false,
		"Disable notification pop-up.",
	)

	flag.Parse()

	if *installFlag && *uninstallFlag {
		fmt.Fprintln(
			os.Stderr,
			"Error: --install and --uninstall flags cannot be used together.",
		)
		os.Exit(1)
	}

	if *installFlag {
		// When installing, pass through the sound path flags if they were
		// provided on the command line.
		// If not provided, the service will use default embedded sounds.
		err := installService(
			*connectSoundPath,
			*disconnectSoundPath,
			*nonotifFlag,
			*nosoundFlag,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error installing service: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *uninstallFlag {
		err := uninstallService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uninstalling service: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	ctx := context.Background()
	devChan, err := monitorUSBEvents(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting monitor: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Monitoring USB devices... (Ctrl+C to stop)")
	for dev := range devChan {
		if dev.Event.Action == "add" || dev.Event.Action == "remove" {
			// Show desktop notification
			if !*nonotifFlag {
				actionStr := "USB Connected"
				if dev.Event.Action == "remove" {
					actionStr = "USB Disconnected"
				}
				if err := beeep.Notify(actionStr, dev.FriendlyName(), ""); err != nil {
					fmt.Printf("Error showing notification: %v\n", err)
				}
			}

			// Print to console if not running as a service
			// (or for debugging service)
			fmt.Println(dev.String())

			// Determine sound path based on action
			var soundPath string
			var isEmbedded bool
			if dev.Event.Action == "add" {
				soundPath, isEmbedded = getSoundPath(
					"connect.wav",
					*connectSoundPath,
					"sounds/connect.wav",
				)
			} else if dev.Event.Action == "remove" {
				soundPath, isEmbedded = getSoundPath(
					"disconnect.wav",
					*disconnectSoundPath,
					"sounds/disconnect.wav",
				)
			}

			// Play the selected sound
			if !*nosoundFlag {
				if err := playSound(soundPath, isEmbedded); err != nil {
					fmt.Printf("Error playing sound %s: %v\n", soundPath, err)
				}
			}
		}
	}
}
