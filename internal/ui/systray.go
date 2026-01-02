package ui

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"fyne.io/systray"
	"github/com/inquire/kefbar-go/internal/config"
	"github/com/inquire/kefbar-go/internal/controller"
	"github/com/inquire/kefbar-go/internal/discovery"
)

// App represents the systray application.
type App struct {
	ctrl              *controller.Controller
	cfg               *config.Config
	lastVolume        int
	onHotkeyUpdate    func()
}

// NewApp creates a new systray application.
func NewApp(ctrl *controller.Controller, cfg *config.Config) *App {
	return &App{
		ctrl:       ctrl,
		cfg:        cfg,
		lastVolume: -1,
	}
}

// SetHotkeyUpdateCallback sets the callback for when hotkeys are updated.
func (a *App) SetHotkeyUpdateCallback(cb func()) {
	a.onHotkeyUpdate = cb
}

// Run starts the systray application.
func (a *App) Run(onExit func()) {
	systray.Run(a.onReady, onExit)
}

// onReady sets up the systray menu.
func (a *App) onReady() {
	systray.SetIcon(GenerateVolumeIcon(0))
	systray.SetTitle("")
	systray.SetTooltip("KEF Speaker Controller")

	// Menu items
	statusItem := systray.AddMenuItem("🔌 Not Connected", "")
	statusItem.Disable()

	volumeItem := systray.AddMenuItem("🔊 Volume: --", "")
	volumeItem.Disable()

	playbackItem := systray.AddMenuItem("🎵 No playback info", "")
	playbackItem.Disable()

	systray.AddSeparator()

	prevItem := systray.AddMenuItem("⏮️ Previous Track", "")
	nextItem := systray.AddMenuItem("⏭️ Next Track", "")

	systray.AddSeparator()

	discoverItem := systray.AddMenuItem("🔍 Discover Speaker", "")

	systray.AddSeparator()

	// Settings submenu
	settingsItem := systray.AddMenuItem("⚙️ Speaker Settings", "")
	hotkeyItem := systray.AddMenuItem("⌨️ Hotkey Settings", "")
	
	// Show current hotkey bindings
	hotkeyInfoItem := systray.AddMenuItem(
		fmt.Sprintf("   Vol+: %s  Vol-: %s", 
			a.cfg.VolumeUpHotkey.String(), 
			a.cfg.VolumeDownHotkey.String()),
		"")
	hotkeyInfoItem.Disable()

	systray.AddSeparator()

	quitItem := systray.AddMenuItem("🚪 Quit", "")

	// Start update loop
	go a.updateLoop(statusItem, volumeItem, playbackItem, hotkeyInfoItem)

	// Handle menu clicks
	go a.handleMenuClicks(
		prevItem, nextItem, discoverItem,
		settingsItem, hotkeyItem, volumeItem, quitItem,
	)
}

// updateLoop periodically updates the UI with current state.
func (a *App) updateLoop(statusItem, volumeItem, playbackItem, hotkeyInfoItem *systray.MenuItem) {
	ticker := time.NewTicker(config.DefaultUIInterval)
	defer ticker.Stop()

	for range ticker.C {
		state := a.ctrl.GetState()

		if state.Connected {
			statusText := "✅ Connected: " + state.IPAddress
			if state.Model != "" {
				statusText = "✅ " + state.Model + " (" + state.IPAddress + ")"
			}
			statusItem.SetTitle(statusText)
			volumeItem.SetTitle(fmt.Sprintf("🔊 Volume: %d%%", state.Volume))
			volumeItem.Enable()

			// Update icon if volume changed
			if state.Volume != a.lastVolume {
				systray.SetIcon(GenerateVolumeIcon(state.Volume))
				a.lastVolume = state.Volume
			}

			if state.PlaybackInfo != nil {
				info := state.PlaybackInfo
				title := "No title"
				if info.Title != "" {
					title = info.Title
				}
				if info.Artist != "" {
					title += " - " + info.Artist
				}
				playbackItem.SetTitle("🎵 " + title)
			} else {
				playbackItem.SetTitle("🎵 No playback info")
			}
		} else {
			statusItem.SetTitle("🔌 Not Connected")
			volumeItem.SetTitle("🔊 Volume: --")
			volumeItem.Disable()
			playbackItem.SetTitle("🎵 No playback info")

			if a.lastVolume != -1 {
				systray.SetIcon(GenerateVolumeIcon(0))
				a.lastVolume = -1
			}
		}

		if state.Error != "" {
			statusItem.SetTitle("❌ Error: " + state.Error)
		}

		// Update hotkey info display
		hotkeyInfoItem.SetTitle(fmt.Sprintf("   Vol+: %s  Vol-: %s",
			a.cfg.VolumeUpHotkey.String(),
			a.cfg.VolumeDownHotkey.String()))
	}
}

// handleMenuClicks processes menu item clicks.
func (a *App) handleMenuClicks(
	prevItem, nextItem, discoverItem,
	settingsItem, hotkeyItem, volumeItem, quitItem *systray.MenuItem,
) {
	for {
		select {
		case <-prevItem.ClickedCh:
			slog.Info("Previous track requested")
			if err := a.ctrl.PreviousTrack(); err != nil {
				slog.Error("Failed to skip previous", "error", err)
			}

		case <-nextItem.ClickedCh:
			slog.Info("Next track requested")
			if err := a.ctrl.NextTrack(); err != nil {
				slog.Error("Failed to skip next", "error", err)
			}

		case <-discoverItem.ClickedCh:
			go a.handleDiscovery(discoverItem)

		case <-settingsItem.ClickedCh:
			slog.Info("Speaker settings opened")
			ShowSettingsDialog(a.ctrl)

		case <-hotkeyItem.ClickedCh:
			slog.Info("Hotkey settings opened")
			ShowHotkeySettingsDialog(a.cfg, a.onHotkeyUpdate)

		case <-volumeItem.ClickedCh:
			slog.Info("Volume dialog opened")
			ShowVolumeDialog(a.ctrl)

		case <-quitItem.ClickedCh:
			slog.Info("Quit requested")
			systray.Quit()
			return
		}
	}
}

// handleDiscovery performs speaker discovery.
func (a *App) handleDiscovery(discoverItem *systray.MenuItem) {
	slog.Info("Starting discovery")
	discoverItem.SetTitle("🔄 Discovering...")
	discoverItem.Disable()

	ip, err := discovery.Discover(context.Background(), 10*time.Second)
	if err == nil {
		slog.Info("Discovery found speaker", "ip", ip)
		a.ctrl.SetIP(ip)
		_ = config.SaveIP(ip)

		if err := a.ctrl.Connect(); err != nil {
			slog.Error("Connection failed after discovery", "error", err)
		} else {
			slog.Info("Connected to discovered speaker", "ip", ip)
		}
	} else {
		slog.Warn("Discovery failed", "error", err)
	}

	discoverItem.SetTitle("🔍 Discover Speaker")
	discoverItem.Enable()
}
