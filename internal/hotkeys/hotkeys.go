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
	ctrl          *controller.Controller
	cfg           *config.Config
	hkUp          *hotkey.Hotkey
	hkDown        *hotkey.Hotkey
	hkPlayPause   *hotkey.Hotkey
	mu            sync.Mutex
	stopUp        chan struct{}
	stopDown      chan struct{}
	stopPlayPause chan struct{}
}

// NewManager creates a new hotkey manager.
func NewManager(ctrl *controller.Controller, cfg *config.Config) *Manager {
	return &Manager{
		ctrl: ctrl,
		cfg:  cfg,
	}
}

// Register registers global hotkeys for playback control.
func (m *Manager) Register() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopUp = make(chan struct{})
	m.stopDown = make(chan struct{})
	m.stopPlayPause = make(chan struct{})

	go m.registerVolumeUp()
	go m.registerVolumeDown()
	go m.registerPlayPause()
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

// registerPlayPause sets up the play/pause hotkey.
func (m *Manager) registerPlayPause() {
	modifiers := parseModifiers(m.cfg.PlayPauseHotkey.Modifiers)
	key := parseKey(m.cfg.PlayPauseHotkey.Key)

	if key == 0 {
		slog.Warn("Invalid play/pause key", "key", m.cfg.PlayPauseHotkey.Key)
		return
	}

	m.hkPlayPause = hotkey.New(modifiers, key)

	if err := m.hkPlayPause.Register(); err != nil {
		slog.Warn("Failed to register play/pause hotkey", "error", err, "binding", m.cfg.PlayPauseHotkey.String())
		return
	}

	slog.Info("Registered play/pause hotkey", "binding", m.cfg.PlayPauseHotkey.String())

	for {
		select {
		case <-m.stopPlayPause:
			return
		case <-m.hkPlayPause.Keydown():
			state := m.ctrl.GetState()
			if !state.Connected {
				continue
			}

			wasPlaying := m.ctrl.IsPlaying()
			if err := m.ctrl.PlayPause(); err != nil {
				slog.Error("Failed to toggle play/pause via hotkey", "error", err)
			} else {
				if wasPlaying {
					slog.Info("Paused via hotkey")
				} else {
					slog.Info("Playing via hotkey")
				}
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
	if m.stopPlayPause != nil {
		close(m.stopPlayPause)
	}

	if m.hkUp != nil {
		_ = m.hkUp.Unregister()
		m.hkUp = nil
	}
	if m.hkDown != nil {
		_ = m.hkDown.Unregister()
		m.hkDown = nil
	}
	if m.hkPlayPause != nil {
		_ = m.hkPlayPause.Unregister()
		m.hkPlayPause = nil
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
	case "p":
		return hotkey.Key('P')
	case "s":
		return hotkey.Key('S')
	case ">":
		return hotkey.Key('>')
	case "<":
		return hotkey.Key('<')
	case ".":
		return hotkey.Key('.')
	case ",":
		return hotkey.Key(',')
	case "space":
		return hotkey.KeySpace
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
