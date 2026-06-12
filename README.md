# Local Clipboard 📋

A simple, elegant local network clipboard/chat application. Share text and files between devices on your local network without internet.

<p align="center">
  <img src="./assets/screenshot.jpeg" alt="Local Clipboard Screenshot" width="700">
</p>

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

## Getting Started

### Download Latest Release

Download the latest prebuilt binary for your operating system from the [Releases](../../releases) page:

- **macOS**: Download `local-clipboard-vX.X.X-mac-silicon` or `local-clipboard-vX.X.X-mac-intel`
- **Linux**: Download `local-clipboard-vX.X.X-linux-amd64`
- **Windows**: Download `local-clipboard-vX.X.X-windows-amd64.exe`

All binaries are self-contained with embedded web assets - no additional dependencies required.

### Run the Application

**macOS/Linux:**

```bash
# Make it executable
chmod +x ./local-clipboard-*

# Run the server (default port 8080)
./local-clipboard-*

# Or with custom port
./local-clipboard-* -port 3000
```

**Windows:**

```bash
# Run the server
local-clipboard-*.exe

# Or with custom port
local-clipboard-*.exe -port 3000
```

The terminal will display the server URLs for both localhost and your local network IP.

## Usage

1. **On your laptop/desktop:**
   - Open `http://localhost:8080`

2. **On your phone/tablet:**
   - Open `http://<your-ip>:8080` (e.g., `http://192.168.1.100:8080`)
   - The exact URL is shown in the terminal when you start the server

3. **Start sharing:**
   - Type a message and press Enter to send (Shift+Enter for new line)
   - Click the 📎 button to attach a file
   - Messages and files appear instantly on all connected devices
   - Click "Copy" to copy text to clipboard
   - Click "Download" to save files

## Building from Source

If you prefer to build from source:

### Prerequisites

- Go 1.24.0 or later installed
- Devices connected to the same local network

### Build and Run

```bash
# Show all available commands
make help

# Install dependencies
go mod download

# Run the server (default port 8080)
make run

# Run with custom port
make run PORT=3000

# Build for multiple platforms
make build
```

### Offline Development

If you need to work offline, you can vendor the dependencies:

```bash
go mod vendor
```

This will download all dependencies into a `vendor/` directory for offline use.

## Docker

### Run with Docker

```bash
docker run -p 8080:8080 ghcr.io/mokhajavi75/local-clipboard:latest
```

### Run with Docker Compose

```yaml
services:
  local-clipboard:
    image: ghcr.io/mokhajavi75/local-clipboard:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
```

### QR Code Host in Docker

Inside a container `getLocalIP()` returns the container's internal IP (e.g. `172.17.0.x`), which is not reachable from the host or other devices on the LAN. Set `LOCAL_CLIPBOARD_HOST` to the host's reachable address so the QR code and startup logs point clients to the right place:

```bash
docker run -p 8080:8080 -e LOCAL_CLIPBOARD_HOST=192.168.1.100 ghcr.io/mokhajavi75/local-clipboard:latest
```

```yaml
services:
  local-clipboard:
    image: ghcr.io/mokhajavi75/local-clipboard:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - LOCAL_CLIPBOARD_HOST=192.168.1.100
```

### Behind a Reverse Proxy

When running behind a reverse proxy, configure it to forward the real client IP so sender labels show correctly instead of the Docker gateway IP.

**Caddy:**

```caddy
example.com {
  reverse_proxy 127.0.0.1:8080 {
    header_up X-Real-IP {remote_host}
  }
}
```

> Caddy sets `X-Forwarded-For` automatically; the `header_up` line adds `X-Real-IP` as well.

**nginx:**

```nginx
location / {
  proxy_pass http://127.0.0.1:8080;
  proxy_set_header X-Forwarded-For $remote_addr;
  proxy_set_header X-Real-IP       $remote_addr;
}
```

Enjoy your local clipboard! 🎉
