// Package controller provides the business logic for KEF speaker control.
package controller

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github/com/inquire/kefbar-go/internal/api"
	"github/com/inquire/kefbar-go/internal/config"
	"github/com/inquire/kefbar-go/pkg/kef"
)

// Controller manages the KEF speaker state and operations.
type Controller struct {
	client *api.Client
	state  *kef.SpeakerState
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.Config
}

// New creates a new Controller.
func New(cfg *config.Config) *Controller {
	ctx, cancel := context.WithCancel(context.Background())

	client := api.NewClient(cfg.SpeakerIP, cfg.Port, cfg.Timeout)
	client.SetContext(ctx)

	return &Controller{
		client: client,
		state: &kef.SpeakerState{
			Port: cfg.Port,
		},
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}
}

// SetIP sets the speaker IP address.
func (c *Controller) SetIP(ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state.IPAddress = ip
	c.state.Error = ""
	c.client.SetHost(ip)
}

// Connect establishes a connection to the speaker.
func (c *Controller) Connect() error {
	c.mu.RLock()
	ip := c.state.IPAddress
	c.mu.RUnlock()

	if ip == "" {
		return fmt.Errorf("no IP address set")
	}

	// Test connection by getting volume
	_, err := c.GetVolume()
	if err != nil {
		c.mu.Lock()
		c.state.Connected = false
		c.state.Error = err.Error()
		c.mu.Unlock()
		return err
	}

	// Get speaker model
	model, err := c.GetSpeakerModel()
	if err != nil {
		slog.Warn("Could not get speaker model", "error", err)
	} else {
		slog.Info("Speaker model detected", "model", model)
	}

	c.mu.Lock()
	c.state.Connected = true
	c.state.Error = ""
	c.mu.Unlock()

	// Start periodic updates
	go c.startPeriodicUpdates()

	return nil
}

// Close shuts down the controller.
func (c *Controller) Close() {
	c.cancel()
}

// GetState returns a copy of the current speaker state.
func (c *Controller) GetState() kef.SpeakerState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.state
}

// GetVolume retrieves the current volume level.
func (c *Controller) GetVolume() (int, error) {
	volume, err := c.client.GetInt("player:volume")
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	c.state.Volume = volume
	c.mu.Unlock()

	return volume, nil
}

// SetVolume sets the volume level (0-100).
func (c *Controller) SetVolume(level int) error {
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}

	err := c.client.SetInt("player:volume", level)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.state.Volume = level
	c.mu.Unlock()

	return nil
}

// VolumeUp increases volume by the configured step.
func (c *Controller) VolumeUp() error {
	c.mu.RLock()
	current := c.state.Volume
	c.mu.RUnlock()

	newVol := current + c.cfg.VolumeStep
	if newVol > 100 {
		newVol = 100
	}

	return c.SetVolume(newVol)
}

// VolumeDown decreases volume by the configured step.
func (c *Controller) VolumeDown() error {
	c.mu.RLock()
	current := c.state.Volume
	c.mu.RUnlock()

	newVol := current - c.cfg.VolumeStep
	if newVol < 0 {
		newVol = 0
	}

	return c.SetVolume(newVol)
}

// GetSpeakerModel retrieves the speaker model from firmware info.
func (c *Controller) GetSpeakerModel() (string, error) {
	releaseText, err := c.client.GetString("settings:/releasetext")
	if err != nil {
		return "", err
	}

	// Model is the first part before underscore (e.g., "LSXII_4.0.1")
	parts := strings.Split(releaseText, "_")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid release text format")
	}

	model := parts[0]

	c.mu.Lock()
	c.state.Model = model
	c.mu.Unlock()

	return model, nil
}

// NextTrack skips to the next track.
func (c *Controller) NextTrack() error {
	err := c.client.SetData("player:player/control", "activate", `{"control":"next"}`)
	if err != nil {
		return err
	}

	// Refresh playback info after a delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		_, _ = c.GetPlaybackInfo()
	}()

	return nil
}

// PreviousTrack skips to the previous track.
func (c *Controller) PreviousTrack() error {
	err := c.client.SetData("player:player/control", "activate", `{"control":"previous"}`)
	if err != nil {
		return err
	}

	// Refresh playback info after a delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		_, _ = c.GetPlaybackInfo()
	}()

	return nil
}

// GetPlaybackInfo retrieves current playback information.
func (c *Controller) GetPlaybackInfo() (*kef.PlaybackInfo, error) {
	result, err := c.client.GetData("player:player/data", "value")
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("empty playback response")
	}

	data, ok := result[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid playback response format")
	}

	info := &kef.PlaybackInfo{}

	// Extract state
	if state, ok := data["state"].(string); ok {
		info.State = state
	}

	// Extract duration from status
	if status, ok := data["status"].(map[string]interface{}); ok {
		if duration, ok := status["duration"].(float64); ok {
			info.Duration = int(duration)
		}
	}

	// Extract track info from trackRoles
	if trackRoles, ok := data["trackRoles"].(map[string]interface{}); ok {
		if title, ok := trackRoles["title"].(string); ok {
			info.Title = title
		}
		if icon, ok := trackRoles["icon"].(string); ok {
			info.AlbumArt = icon
		}

		// Extract metadata
		if mediaData, ok := trackRoles["mediaData"].(map[string]interface{}); ok {
			if metaData, ok := mediaData["metaData"].(map[string]interface{}); ok {
				if artist, ok := metaData["artist"].(string); ok {
					info.Artist = artist
				}
				if album, ok := metaData["album"].(string); ok {
					info.Album = album
				}
			}
		}
	}

	c.mu.Lock()
	c.state.PlaybackInfo = info
	c.mu.Unlock()

	return info, nil
}

// startPeriodicUpdates polls the speaker for state updates.
func (c *Controller) startPeriodicUpdates() {
	ticker := time.NewTicker(c.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			connected := c.state.Connected
			c.mu.RUnlock()

			if connected {
				_, _ = c.GetVolume()
				_, _ = c.GetPlaybackInfo()
			}
		}
	}
}
