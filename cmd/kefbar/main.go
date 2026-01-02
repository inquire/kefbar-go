// KEF Bar - macOS menu bar application for controlling KEF speakers
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github/com/inquire/kefbar-go/internal/config"
	"github/com/inquire/kefbar-go/internal/controller"
	"github/com/inquire/kefbar-go/internal/hotkeys"
	"github/com/inquire/kefbar-go/internal/ui"
)

func main() {
	// Setup structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("KEF Bar starting...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Warn("Failed to load config", "error", err)
		cfg = config.New()
	}

	// Create controller
	ctrl := controller.New(cfg)
	defer ctrl.Close()

	// Auto-connect if we have a saved IP
	if cfg.SpeakerIP != "" {
		slog.Info("Loading saved IP", "ip", cfg.SpeakerIP)
		ctrl.SetIP(cfg.SpeakerIP)

		go func() {
			if err := ctrl.Connect(); err != nil {
				slog.Warn("Failed to connect to saved IP", "ip", cfg.SpeakerIP, "error", err)
			} else {
				slog.Info("Connected to speaker", "ip", cfg.SpeakerIP)
			}
		}()
	}

	// Register global hotkeys
	hotkeyMgr := hotkeys.NewManager(ctrl, cfg)
	hotkeyMgr.Register()
	defer hotkeyMgr.Unregister()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received interrupt signal, quitting...")
		os.Exit(0)
	}()

	// Create and run the systray app
	app := ui.NewApp(ctrl, cfg)

	// Set callback to re-register hotkeys when settings change
	app.SetHotkeyUpdateCallback(func() {
		slog.Info("Re-registering hotkeys after settings change")
		hotkeyMgr.Reregister()
	})

	onExit := func() {
		slog.Info("KEF Bar shutting down...")
		hotkeyMgr.Unregister()
		ctrl.Close()
		os.Exit(0)
	}

	app.Run(onExit)
}
