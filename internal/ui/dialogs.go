package ui

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"

	"github/com/inquire/kefbar-go/internal/config"
	"github/com/inquire/kefbar-go/internal/controller"
)

// ShowSettingsDialog displays a native macOS dialog to enter speaker IP.
func ShowSettingsDialog(ctrl *controller.Controller) {
	state := ctrl.GetState()
	currentIP := state.IPAddress
	if currentIP == "" {
		currentIP = "192.168.1.100"
	}

	script := fmt.Sprintf(`
		set dialogResult to display dialog "Enter KEF Speaker IP Address:" default answer "%s" buttons {"Cancel", "Connect"} default button "Connect" with title "KEF Bar Settings"
		if button returned of dialogResult is "Connect" then
			return text returned of dialogResult
		else
			return ""
		end if
	`, currentIP)

	go func() {
		cmd := exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err != nil {
			slog.Debug("Settings dialog cancelled or error", "error", err)
			return
		}

		ip := strings.TrimSpace(string(output))
		if ip == "" {
			return
		}

		slog.Info("Connect requested via settings", "ip", ip)
		ctrl.SetIP(ip)
		_ = config.SaveIP(ip)

		if err := ctrl.Connect(); err != nil {
			slog.Error("Connection failed", "error", err)
			ShowAlert("Connection Failed", fmt.Sprintf("Could not connect to %s: %v", ip, err))
		} else {
			slog.Info("Connected to speaker", "ip", ip)
			ShowAlert("Connected", fmt.Sprintf("Successfully connected to %s", ip))
		}
	}()
}

// ShowVolumeDialog displays a native macOS dialog to set volume.
func ShowVolumeDialog(ctrl *controller.Controller) {
	state := ctrl.GetState()
	if !state.Connected {
		ShowAlert("Not Connected", "Please connect to a speaker first.")
		return
	}

	currentVol := state.Volume

	script := fmt.Sprintf(`
		set dialogResult to display dialog "Enter volume (0-100):" default answer "%d" buttons {"Cancel", "Set Volume"} default button "Set Volume" with title "KEF Bar Volume"
		if button returned of dialogResult is "Set Volume" then
			return text returned of dialogResult
		else
			return ""
		end if
	`, currentVol)

	go func() {
		cmd := exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err != nil {
			slog.Debug("Volume dialog cancelled or error", "error", err)
			return
		}

		volStr := strings.TrimSpace(string(output))
		if volStr == "" {
			return
		}

		vol, err := strconv.Atoi(volStr)
		if err != nil {
			ShowAlert("Invalid Volume", "Please enter a number between 0 and 100.")
			return
		}

		if vol < 0 || vol > 100 {
			ShowAlert("Invalid Volume", "Volume must be between 0 and 100.")
			return
		}

		oldVol := ctrl.GetState().Volume
		if err := ctrl.SetVolume(vol); err != nil {
			slog.Error("Failed to set volume", "error", err)
			ShowAlert("Error", fmt.Sprintf("Could not set volume: %v", err))
		} else {
			slog.Info("Volume changed via dialog", "old", oldVol, "new", vol)
		}
	}()
}

// ShowAlert displays a native macOS alert.
func ShowAlert(title, message string) {
	script := fmt.Sprintf(`display alert "%s" message "%s" as informational`, title, message)
	cmd := exec.Command("osascript", "-e", script)
	_ = cmd.Run()
}

// HotkeyCallback is called when hotkeys are updated.
type HotkeyCallback func()

// ShowHotkeySettingsDialog displays a dialog to configure hotkey bindings.
func ShowHotkeySettingsDialog(cfg *config.Config, onUpdate HotkeyCallback) {
	// Build modifier options string
	modifierOptions := strings.Join(config.AvailableModifiers, ", ")
	keyOptions := strings.Join(config.AvailableKeys, ", ")

	script := fmt.Sprintf(`
		set volumeUpMod to "%s"
		set volumeUpKey to "%s"
		set volumeDownMod to "%s"
		set volumeDownKey to "%s"

		-- Volume Up Modifiers
		set dialogResult to display dialog "Volume Up - Modifiers:" & return & return & "Options: %s" default answer volumeUpMod buttons {"Cancel", "Next"} default button "Next" with title "KEF Bar - Hotkey Settings (1/4)"
		if button returned of dialogResult is "Cancel" then
			return "CANCELLED"
		end if
		set volumeUpMod to text returned of dialogResult

		-- Volume Up Key
		set dialogResult to display dialog "Volume Up - Key:" & return & return & "Options: %s" default answer volumeUpKey buttons {"Cancel", "Next"} default button "Next" with title "KEF Bar - Hotkey Settings (2/4)"
		if button returned of dialogResult is "Cancel" then
			return "CANCELLED"
		end if
		set volumeUpKey to text returned of dialogResult

		-- Volume Down Modifiers
		set dialogResult to display dialog "Volume Down - Modifiers:" & return & return & "Options: %s" default answer volumeDownMod buttons {"Cancel", "Next"} default button "Next" with title "KEF Bar - Hotkey Settings (3/4)"
		if button returned of dialogResult is "Cancel" then
			return "CANCELLED"
		end if
		set volumeDownMod to text returned of dialogResult

		-- Volume Down Key
		set dialogResult to display dialog "Volume Down - Key:" & return & return & "Options: %s" default answer volumeDownKey buttons {"Cancel", "Save"} default button "Save" with title "KEF Bar - Hotkey Settings (4/4)"
		if button returned of dialogResult is "Cancel" then
			return "CANCELLED"
		end if
		set volumeDownKey to text returned of dialogResult

		return volumeUpMod & "|" & volumeUpKey & "|" & volumeDownMod & "|" & volumeDownKey
	`,
		cfg.VolumeUpHotkey.Modifiers,
		cfg.VolumeUpHotkey.Key,
		cfg.VolumeDownHotkey.Modifiers,
		cfg.VolumeDownHotkey.Key,
		modifierOptions,
		keyOptions,
		modifierOptions,
		keyOptions,
	)

	go func() {
		cmd := exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err != nil {
			slog.Debug("Hotkey settings cancelled or error", "error", err)
			return
		}

		result := strings.TrimSpace(string(output))
		if result == "" || result == "CANCELLED" {
			return
		}

		parts := strings.Split(result, "|")
		if len(parts) != 4 {
			slog.Error("Invalid hotkey settings result", "result", result)
			return
		}

		// Update config
		cfg.VolumeUpHotkey.Modifiers = strings.TrimSpace(parts[0])
		cfg.VolumeUpHotkey.Key = strings.TrimSpace(parts[1])
		cfg.VolumeDownHotkey.Modifiers = strings.TrimSpace(parts[2])
		cfg.VolumeDownHotkey.Key = strings.TrimSpace(parts[3])

		// Save config
		if err := cfg.Save(); err != nil {
			slog.Error("Failed to save hotkey settings", "error", err)
			ShowAlert("Error", "Failed to save hotkey settings.")
			return
		}

		slog.Info("Hotkey settings updated",
			"volumeUp", cfg.VolumeUpHotkey.String(),
			"volumeDown", cfg.VolumeDownHotkey.String())

		// Notify caller to re-register hotkeys
		if onUpdate != nil {
			onUpdate()
		}

		ShowAlert("Hotkeys Updated", fmt.Sprintf(
			"Volume Up: %s\nVolume Down: %s\n\nHotkeys will be re-registered.",
			cfg.VolumeUpHotkey.String(),
			cfg.VolumeDownHotkey.String()))
	}()
}

