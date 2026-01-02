// Package config handles application configuration.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Default configuration values.
const (
	DefaultPort         = 80
	DefaultVolumeStep   = 5
	DefaultPollInterval = 3 * time.Second
	DefaultTimeout      = 5 * time.Second
	DefaultUIInterval   = 1 * time.Second
	ConfigFileName      = ".kefbar.json"
	LegacyConfigFile    = ".kefbar_ip"
)

// Default hotkey bindings.
const (
	DefaultVolumeUpModifiers   = "Cmd+Shift"
	DefaultVolumeUpKey         = "Up"
	DefaultVolumeDownModifiers = "Cmd+Shift"
	DefaultVolumeDownKey       = "Down"
)

// HotkeyBinding represents a keyboard shortcut configuration.
type HotkeyBinding struct {
	Modifiers string `json:"modifiers"` // e.g., "Cmd+Shift", "Ctrl+Alt"
	Key       string `json:"key"`       // e.g., "Up", "Down", "F1"
}

// String returns a human-readable representation of the hotkey.
func (h HotkeyBinding) String() string {
	if h.Modifiers == "" {
		return h.Key
	}
	return h.Modifiers + "+" + h.Key
}

// Config holds the application configuration.
type Config struct {
	SpeakerIP        string        `json:"speaker_ip"`
	Port             int           `json:"port"`
	VolumeStep       int           `json:"volume_step"`
	VolumeUpHotkey   HotkeyBinding `json:"volume_up_hotkey"`
	VolumeDownHotkey HotkeyBinding `json:"volume_down_hotkey"`

	// Non-persisted runtime values
	PollInterval time.Duration `json:"-"`
	Timeout      time.Duration `json:"-"`
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{
		Port:         DefaultPort,
		VolumeStep:   DefaultVolumeStep,
		PollInterval: DefaultPollInterval,
		Timeout:      DefaultTimeout,
		VolumeUpHotkey: HotkeyBinding{
			Modifiers: DefaultVolumeUpModifiers,
			Key:       DefaultVolumeUpKey,
		},
		VolumeDownHotkey: HotkeyBinding{
			Modifiers: DefaultVolumeDownModifiers,
			Key:       DefaultVolumeDownKey,
		},
	}
}

// Load loads the configuration from disk.
func Load() (*Config, error) {
	cfg := New()

	path, err := configFilePath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Try legacy config file for backwards compatibility
		if ip, legacyErr := loadLegacyIP(); legacyErr == nil {
			cfg.SpeakerIP = ip
		}
		return cfg, nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	// Ensure runtime values are set
	cfg.PollInterval = DefaultPollInterval
	cfg.Timeout = DefaultTimeout

	return cfg, nil
}

// Save saves the configuration to disk.
func (c *Config) Save() error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// configFilePath returns the path to the config file.
func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigFileName), nil
}

// loadLegacyIP loads IP from the old config format.
func loadLegacyIP() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(filepath.Join(home, LegacyConfigFile))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// LoadSavedIP loads the saved speaker IP (for backwards compatibility).
func LoadSavedIP() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.SpeakerIP, nil
}

// SaveIP saves the speaker IP to disk.
func SaveIP(ip string) error {
	cfg, _ := Load()
	cfg.SpeakerIP = ip
	return cfg.Save()
}

// Available modifier options for the UI.
var AvailableModifiers = []string{
	"Cmd+Shift",
	"Cmd+Ctrl",
	"Cmd+Alt",
	"Ctrl+Shift",
	"Ctrl+Alt",
	"Alt+Shift",
	"Cmd",
	"Ctrl",
	"Alt",
	"Shift",
}

// Available key options for the UI.
var AvailableKeys = []string{
	"Up",
	"Down",
	"Left",
	"Right",
	"F1",
	"F2",
	"F3",
	"F4",
	"F5",
	"F6",
	"F7",
	"F8",
	"F9",
	"F10",
	"F11",
	"F12",
	"[",
	"]",
	"=",
	"-",
}
