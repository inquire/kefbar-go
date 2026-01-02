// Package kef provides shared types for KEF speaker control.
package kef

// PlaybackInfo contains information about the currently playing track.
type PlaybackInfo struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	AlbumArt string `json:"album_art"`
	Duration int    `json:"duration"`
	Position int    `json:"position"`
	State    string `json:"state"`
}

// SpeakerState represents the current state of a KEF speaker.
type SpeakerState struct {
	IPAddress    string
	Port         int
	Connected    bool
	Volume       int
	PlaybackInfo *PlaybackInfo
	IsPoweredOn  bool
	Error        string
	Model        string // Speaker model (e.g., "LSXII", "LS50WII")
}

// Speaker defines the interface for controlling a KEF speaker.
type Speaker interface {
	// Connection
	SetIP(ip string)
	Connect() error
	Close()
	GetState() SpeakerState

	// Volume
	GetVolume() (int, error)
	SetVolume(level int) error

	// Playback
	GetPlaybackInfo() (*PlaybackInfo, error)
	NextTrack() error
	PreviousTrack() error

	// Info
	GetSpeakerModel() (string, error)
}
