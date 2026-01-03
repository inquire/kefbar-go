# 🔊 KEF Bar

> A sleek macOS menu bar app to control your KEF wireless speakers

<!-- Update the Build badge URL to match your GitHub repository -->
[![Build][![Build](https://github.com/inquire/kefbar-go/actions/workflows/build.yaml/badge.svg)](https://github.com/inquire/kefbar-go/actions/workflows/build.yaml)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![macOS](https://img.shields.io/badge/macOS-11.0+-000000?style=flat&logo=apple)
![License](https://img.shields.io/badge/license-MIT-green)

## ✨ What is this?

**KEF Bar** is a lightweight native macOS menu bar application that lets you control your KEF wireless speakers (LSX II, LS50 Wireless II, etc.) directly from your menu bar. No more reaching for your phone or the KEF app - just click the menu bar icon!

### 🎯 Key Features

| Feature | Description |
|---------|-------------|
| 🎚️ **Volume Control** | Adjust volume with customizable keyboard shortcuts |
| ⌨️ **Configurable Hotkeys** | Set your own keyboard shortcuts for volume and play/pause |
| 📊 **Visual Volume Indicator** | Menu bar icon shows current volume level as a fill indicator |
| 🔍 **Auto-Discovery** | Automatically finds KEF speakers on your network |
| 🎵 **Now Playing** | See what's currently playing on your speaker |
| ⏯️ **Play/Pause** | Toggle playback with Cmd+Shift+Space |
| ⏭️ **Track Control** | Skip tracks without leaving your keyboard |
| 🏷️ **Model Detection** | Identifies your speaker model (LSX II, LS50W2, etc.) |

## 🖼️ How It Works

The app lives in your macOS menu bar and displays a KEF "K" logo that fills up based on your current volume level:

- **Empty outline** = Volume at 0% (muted)
- **Partially filled** = Volume somewhere in between
- **Fully filled** = Volume at 100%

Click the icon to see:
- 📡 Connection status with speaker model
- 🔊 Current volume percentage (clickable to set volume)
- 🎵 Now playing information
- ⏮️ ▶️/⏸️ ⏭️ Playback controls (previous, play/pause, next)
- 🔍 Speaker discovery
- ⚙️ Speaker settings
- ⌨️ Hotkey settings (with current bindings displayed)

## ⌨️ Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd + Shift + ↑` | Volume Up (+5%) |
| `Cmd + Shift + ↓` | Volume Down (-5%) |
| `Cmd + Shift + Space` | Play/Pause toggle |

### 🔧 Customizing Hotkeys

You can customize the keyboard shortcuts:

1. Click the menu bar icon
2. Select "⌨️ Hotkey Settings"
3. Configure your preferred modifiers and keys:
   - **Modifiers**: Cmd, Ctrl, Alt, Shift (or combinations like Cmd+Shift)
   - **Keys**: Arrow keys, > < . , P, S, Space, F1-F12, or [ ] = -

Settings are saved to `~/.kefbar.json` and persist across restarts.

## 🚀 Getting Started

### Requirements

- 🍎 macOS 11.0 (Big Sur) or later
- 🔧 Go 1.21 or later
- 🔈 KEF wireless speaker (LSX II, LS50 Wireless II, etc.) on the same network

### Building

```bash
# Clone and build
cd kefbar-go
go mod tidy

# Build with Makefile (recommended)
make build

# Or build directly
go build -o build/kefbar ./cmd/kefbar
```

### Running

```bash
# Using Makefile
make run

# Or run directly
./build/kefbar

# For development (no build step)
make dev
```

The app will:
1. 🔍 Automatically search for KEF speakers on your network
2. 🔗 Connect to the first speaker found
3. 📊 Display the volume indicator in your menu bar

### First Time Setup

If auto-discovery doesn't find your speaker:
1. Click the menu bar icon
2. Select "⚙️ Speaker Settings"
3. Enter your speaker's IP address manually

## 📁 Configuration

All settings are saved to `~/.kefbar.json`:

```json
{
  "speaker_ip": "192.168.1.100",
  "port": 80,
  "volume_step": 5,
  "volume_up_hotkey": {
    "modifiers": "Cmd+Shift",
    "key": "Up"
  },
  "volume_down_hotkey": {
    "modifiers": "Cmd+Shift",
    "key": "Down"
  }
}
```

| Setting | Description | Default |
|---------|-------------|---------|
| `speaker_ip` | Your KEF speaker's IP address | - |
| `port` | HTTP API port | 80 |
| `volume_step` | Volume change per hotkey press | 5% |
| `volume_up_hotkey` | Keyboard shortcut for volume up | Cmd+Shift+Up |
| `volume_down_hotkey` | Keyboard shortcut for volume down | Cmd+Shift+Down |

## 🛠️ Technical Details

### Supported Speakers

- ✅ KEF LSX II
- ✅ KEF LS50 Wireless II
- ✅ Other KEF speakers with the same API

### API Communication

KEF Bar communicates with your speaker over HTTP using the KEF REST API:

| Endpoint | Purpose |
|----------|---------|
| `player:volume` | Get/Set volume level |
| `player:player/control` | Playback control (next/previous) |
| `player:player/data` | Now playing metadata |
| `settings:/deviceName` | Speaker name |
| `settings:/releasetext` | Speaker model & firmware |

Based on the excellent [pykefcontrol](https://github.com/N0ciple/pykefcontrol) Python library.

### Discovery Methods

1. **SSDP** - Multicast discovery protocol
2. **Network Scan** - Fallback scanning of local network

## 📂 Project Structure

```
kefbar-go/
├── cmd/
│   └── kefbar/
│       └── main.go              # 🚀 Entry point (~70 lines)
├── internal/
│   ├── api/
│   │   └── client.go            # 🌐 KEF HTTP API client
│   ├── config/
│   │   └── config.go            # ⚙️ Configuration management
│   ├── controller/
│   │   └── controller.go        # 🎛️ Business logic & state
│   ├── discovery/
│   │   ├── discovery.go         # 🔍 Discovery orchestration
│   │   ├── ssdp.go              # 📡 SSDP multicast discovery
│   │   └── scan.go              # 🔎 Network scan fallback
│   ├── hotkeys/
│   │   └── hotkeys.go           # ⌨️ Keyboard shortcuts
│   └── ui/
│       ├── systray.go           # 📊 Menu bar interface
│       ├── dialogs.go           # 💬 Native macOS dialogs
│       ├── icon.go              # 🎨 Dynamic volume icon
│       └── assets/
│           └── kef.png          # 🖼️ KEF K logo
├── pkg/
│   └── kef/
│       └── types.go             # 📦 Shared types & interfaces
├── icons/
│   └── kef.png                  # 🖼️ KEF K logo asset
├── go.mod                       # 📦 Go module definition
├── Makefile                     # 🔧 Build automation
└── README.md                    # 📖 You are here!
```

## 🤝 Contributing

Feel free to open issues or submit PRs! This is a personal project but contributions are welcome.

## 📄 License

MIT License - feel free to modify and distribute.

---

Made with ❤️ for KEF speaker owners who want quick volume control from their Mac.
