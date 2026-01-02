// Package hotkeys provides global keyboard shortcut handling.
package hotkeys

import (
	"log/slog"
	"strings"
	"sync"

	"github/com/inquire/kefbar-go/internal/config"
	"github/com/inquire/kefbar-go/internal/controller"
	"golang.design/x/hotkey"
)

// Manager handles global hotkey registration.
type Manager struct {
	ctrl   *controller.Controller
	cfg    *config.Config
	hkUp   *hotkey.Hotkey
	hkDown *hotkey.Hotkey
	mu     sync.Mutex
	stopUp   chan struct{}
	stopDown chan struct{}
}

// NewManager creates a new hotkey manager.
func NewManager(ctrl *controller.Controller, cfg *config.Config) *Manager {
	return &Manager{
		ctrl: ctrl,
		cfg:  cfg,
	}
}

// Register registers global hotkeys for volume control.
func (m *Manager) Register() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopUp = make(chan struct{})
	m.stopDown = make(chan struct{})

	go m.registerVolumeUp()
	go m.registerVolumeDown()
}

// Reregister unregisters and re-registers hotkeys with new config.
func (m *Manager) Reregister() {
	m.Unregister()
	m.Register()
}

// registerVolumeUp sets up the volume up hotkey.
func (m *Manager) registerVolumeUp() {
	modifiers := parseModifiers(m.cfg.VolumeUpHotkey.Modifiers)
	key := parseKey(m.cfg.VolumeUpHotkey.Key)

	if key == 0 {
		slog.Warn("Invalid volume up key", "key", m.cfg.VolumeUpHotkey.Key)
		return
	}

	m.hkUp = hotkey.New(modifiers, key)

	if err := m.hkUp.Register(); err != nil {
		slog.Warn("Failed to register volume up hotkey", "error", err, "binding", m.cfg.VolumeUpHotkey.String())
		return
	}

	slog.Info("Registered volume up hotkey", "binding", m.cfg.VolumeUpHotkey.String())

	for {
		select {
		case <-m.stopUp:
			return
		case <-m.hkUp.Keydown():
			state := m.ctrl.GetState()
			if !state.Connected {
				continue
			}

			oldVol := state.Volume
			if err := m.ctrl.VolumeUp(); err != nil {
				slog.Error("Failed to increase volume via hotkey", "error", err)
			} else {
				newState := m.ctrl.GetState()
				slog.Info("Volume changed via hotkey", "old", oldVol, "new", newState.Volume)
			}
		}
	}
}

// registerVolumeDown sets up the volume down hotkey.
func (m *Manager) registerVolumeDown() {
	modifiers := parseModifiers(m.cfg.VolumeDownHotkey.Modifiers)
	key := parseKey(m.cfg.VolumeDownHotkey.Key)

	if key == 0 {
		slog.Warn("Invalid volume down key", "key", m.cfg.VolumeDownHotkey.Key)
		return
	}

	m.hkDown = hotkey.New(modifiers, key)

	if err := m.hkDown.Register(); err != nil {
		slog.Warn("Failed to register volume down hotkey", "error", err, "binding", m.cfg.VolumeDownHotkey.String())
		return
	}

	slog.Info("Registered volume down hotkey", "binding", m.cfg.VolumeDownHotkey.String())

	for {
		select {
		case <-m.stopDown:
			return
		case <-m.hkDown.Keydown():
			state := m.ctrl.GetState()
			if !state.Connected {
				continue
			}

			oldVol := state.Volume
			if err := m.ctrl.VolumeDown(); err != nil {
				slog.Error("Failed to decrease volume via hotkey", "error", err)
			} else {
				newState := m.ctrl.GetState()
				slog.Info("Volume changed via hotkey", "old", oldVol, "new", newState.Volume)
			}
		}
	}
}

// Unregister unregisters all hotkeys.
func (m *Manager) Unregister() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopUp != nil {
		close(m.stopUp)
	}
	if m.stopDown != nil {
		close(m.stopDown)
	}

	if m.hkUp != nil {
		_ = m.hkUp.Unregister()
		m.hkUp = nil
	}
	if m.hkDown != nil {
		_ = m.hkDown.Unregister()
		m.hkDown = nil
	}
}

// parseModifiers converts a modifier string to hotkey modifiers.
func parseModifiers(s string) []hotkey.Modifier {
	var mods []hotkey.Modifier
	s = strings.ToLower(s)

	if strings.Contains(s, "cmd") {
		mods = append(mods, hotkey.ModCmd)
	}
	if strings.Contains(s, "ctrl") {
		mods = append(mods, hotkey.ModCtrl)
	}
	if strings.Contains(s, "alt") {
		mods = append(mods, hotkey.ModOption)
	}
	if strings.Contains(s, "shift") {
		mods = append(mods, hotkey.ModShift)
	}

	return mods
}

// parseKey converts a key string to a hotkey key.
func parseKey(s string) hotkey.Key {
	switch strings.ToLower(s) {
	case "up":
		return hotkey.KeyUp
	case "down":
		return hotkey.KeyDown
	case "left":
		return hotkey.KeyLeft
	case "right":
		return hotkey.KeyRight
	case "f1":
		return hotkey.KeyF1
	case "f2":
		return hotkey.KeyF2
	case "f3":
		return hotkey.KeyF3
	case "f4":
		return hotkey.KeyF4
	case "f5":
		return hotkey.KeyF5
	case "f6":
		return hotkey.KeyF6
	case "f7":
		return hotkey.KeyF7
	case "f8":
		return hotkey.KeyF8
	case "f9":
		return hotkey.KeyF9
	case "f10":
		return hotkey.KeyF10
	case "f11":
		return hotkey.KeyF11
	case "f12":
		return hotkey.KeyF12
	case "[":
		return hotkey.Key('[')
	case "]":
		return hotkey.Key(']')
	case "=":
		return hotkey.Key('=')
	case "-":
		return hotkey.Key('-')
	default:
		return 0
	}
}
