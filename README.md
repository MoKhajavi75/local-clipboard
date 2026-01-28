# Local Clipboard 📋

A simple, elegant local network clipboard/chat application. Share text and files between devices on your local network without internet.

## Features

- ✨ Beautiful, modern UI with gradient design
- 🔄 Real-time text sharing via WebSocket
- 📎 File sharing and download support
- 📱 Works on any device with a browser (desktop, mobile, tablet)
- 🌐 No internet required - works on local network only
- 💾 In-memory storage (clears when server stops)
- 📋 One-click copy to clipboard
- 🎨 Responsive design for all screen sizes
- 🔒 Local network only - no data leaves your network

## Setup

### Prerequisites

- Go 1.21 or later installed
- Devices connected to the same local network

### Quick Start

**Using Makefile (recommended):**

```bash
# Show all available commands
make help

# Run the server (default port 8080)
make run

# Run with custom port
make run PORT=3000

# Build for multiple platforms (macOS, Linux, Windows)
make build
```

The project includes vendored dependencies, so it works completely offline after cloning.

## Usage

1. **Start the server:**

```bash
make run
```

The terminal will show the server URLs for both localhost and your local network IP.

2. **On your laptop:**

- Open `http://localhost:8080`

3. **On your phone/tablet:**

- Open `http://<your-laptop-ip>:8080` (e.g., `http://192.168.1.100:8080`)
- The exact URL is shown in the terminal when you start the server

4. **Start sharing:**

- Type a message and press Enter to send (Shift+Enter for new line)
- Click the 📎 button to attach a file
- Messages and files appear instantly on all connected devices
- Click "Copy" to copy text to clipboard
- Click "Download" to save files

## Find Your Local IP Address

The server displays your IP address when it starts. If you need to find it manually:

**macOS/Linux:**

```bash
ifconfig | grep "inet " | grep -v 127.0.0.1
```

**Windows:**

```bash
ipconfig
```

Look for "IPv4 Address" under your active network adapter.

## Offline Mode

This project includes vendored dependencies for completely offline operation. If you need to update dependencies:

```bash
make vendor
```

This downloads all dependencies into the `vendor/` directory. After that, the project runs completely offline.

## Building Binaries

Build for multiple platforms (macOS, Linux, Windows):

```bash
make build
```

Binaries will be created in the `build/` directory for:

- macOS (Intel & Apple Silicon)
- Linux (amd64)
- Windows (amd64)

Run the binary:

```bash
./build/local-clipboard-1.0-darwin-arm64        # macOS Apple Silicon
./build/local-clipboard-1.0-darwin-amd64        # macOS Intel
./build/local-clipboard-1.0-linux-amd64         # Linux
./build/local-clipboard-1.0-windows-amd64.exe   # Windows
```

## Notes

- All data is stored in memory and cleared when the server stops
- Works only on local network - no internet required
- Multiple devices can connect simultaneously
- Real-time synchronization across all connected devices
- Files are temporarily stored in memory (lost on server restart)
- No data leaves your local network

## Advanced Commands

```bash
make help      # Show all available commands
make install   # Install/update Go dependencies
make vendor    # Update vendored dependencies
```

Enjoy your local clipboard! 🎉
