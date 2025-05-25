package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
)

const serviceName = "gousbnotify.service"
const programName = "gousbnotify"

const systemdServiceTemplate = `[Unit]
Description={{.ProgramName}} - USB device notifier
After=sound.target

[Service]
ExecStart={{.ExecutablePath}} {{.Args}}
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
`

type ServiceTemplateData struct {
	ProgramName    string
	ExecutablePath string
	Args           string
}

func getServiceFilePath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("could not get current user: %w", err)
	}
	return filepath.Join(
		currentUser.HomeDir,
		".config",
		"systemd",
		"user",
		serviceName,
	), nil
}

func runSystemctlCommand(args ...string) error {
	cmdArgs := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Running: systemctl %s\n", strings.Join(cmdArgs, " "))
	return cmd.Run()
}

func installService(
	connectSound string,
	disconnectSound string,
	nonotif bool,
	nosound bool,
) error {
	fmt.Println("Installing systemd user service...")

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}

	serviceFilePath, err := getServiceFilePath()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(serviceFilePath), 0750); err != nil {
		return fmt.Errorf("could not create systemd user directory: %w", err)
	}

	// Prepare arguments for the service,
	// only include sound paths if they are set
	var serviceArgs []string
	if connectSound != "" {
		// If connectSound is a relative path, make it absolute for the service
		absConnectSound, err := filepath.Abs(connectSound)
		if err == nil {
			serviceArgs = append(
				serviceArgs,
				fmt.Sprintf("--connect-sound=\"%s\"", absConnectSound),
			)
		} else {
			fmt.Printf(
				"Warning: could not get absolute path for '%s', using as is.\n",
				connectSound,
			)
			serviceArgs = append(
				serviceArgs,
				fmt.Sprintf("--connect-sound=\"%s\"", connectSound),
			)
		}
	}
	if disconnectSound != "" {
		absDisconnectSound, err := filepath.Abs(disconnectSound)
		if err == nil {
			serviceArgs = append(
				serviceArgs,
				fmt.Sprintf("--disconnect-sound=\"%s\"", absDisconnectSound),
			)
		} else {
			fmt.Printf(
				"Warning: could not get absolute path for '%s', using as is.\n",
				disconnectSound,
			)
			serviceArgs = append(
				serviceArgs,
				fmt.Sprintf("--disconnect-sound=\"%s\"", disconnectSound),
			)
		}
	}

	if nonotif {
		serviceArgs = append(
			serviceArgs,
			"-nonotif",
		)
	}

	if nosound {
		serviceArgs = append(
			serviceArgs,
			"-nosound",
		)
	}

	data := ServiceTemplateData{
		ProgramName:    programName,
		ExecutablePath: execPath,
		Args:           strings.Join(serviceArgs, " "),
	}

	tmpl, err := template.New("service").Parse(systemdServiceTemplate)
	if err != nil {
		return fmt.Errorf("could not parse service template: %w", err)
	}

	file, err := os.Create(serviceFilePath)
	if err != nil {
		return fmt.Errorf(
			"could not create service file %s: %w",
			serviceFilePath,
			err,
		)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("could not write service file: %w", err)
	}
	fmt.Printf("Service file created at %s\n", serviceFilePath)

	if err := runSystemctlCommand("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd daemons: %w", err)
	}
	if err := runSystemctlCommand("enable", serviceName); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}
	if err := runSystemctlCommand("restart", serviceName); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Printf("Service %s installed, enabled and started.\n", serviceName)
	fmt.Println("You can check status: systemctl --user status", serviceName)
	fmt.Println("And view logs with: journalctl --user -u", serviceName)
	return nil
}

func uninstallService() error {
	fmt.Println("Uninstalling systemd user service...")
	serviceFilePath, err := getServiceFilePath()
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(serviceFilePath); os.IsNotExist(statErr) {
		fmt.Printf(
			"Service file %s does not exist. Nothing to do.\n",
			serviceFilePath,
		)
		return nil
	}

	if err := runSystemctlCommand("stop", serviceName); err != nil {
		fmt.Printf(
			"Warning: failed to stop service (it might not be running): %v\n",
			err,
		)
	}
	if err := runSystemctlCommand("disable", serviceName); err != nil {
		fmt.Printf(
			"Warning: failed to disable service (it might not be enabled): %v\n",
			err,
		)
	}

	if err := os.Remove(serviceFilePath); err != nil {
		return fmt.Errorf(
			"failed to remove service file %s: %w",
			serviceFilePath,
			err,
		)
	}
	fmt.Printf("Service file %s removed.\n", serviceFilePath)

	if err := runSystemctlCommand("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd daemons: %w", err)
	}
	// Good practice after removing units
	if err := runSystemctlCommand("reset-failed"); err != nil {
		fmt.Printf(
			"Warning: failed to reset-failed state for systemd user units: %v\n",
			err,
		)
	}

	fmt.Printf("Service %s uninstalled.\n", serviceName)
	return nil
}
