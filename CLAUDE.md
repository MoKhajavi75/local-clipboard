# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Local Clipboard is a self-contained local network clipboard/chat app for sharing text and files between devices on the same network. It compiles to a single Go binary with all web assets embedded — no external dependencies at runtime.

## Commands

```bash
# Run locally (default port 8080)
make run

# Run on a custom port
make run PORT=3000

# Build release binaries for macOS (Intel/Silicon), Linux, Windows
make build

# Build with a specific version
make build VERSION=1.2.3

# Update Go dependencies
make update
```

Binaries are output to `./build/`. Releases are triggered by pushing a `v*` git tag, which also builds and pushes a multi-platform Docker image to GHCR (`ghcr.io/<owner>/local-clipboard`).

There are no tests or linting configurations in this project.

## Architecture

The entire backend is in `main.go`. Web assets (`web/index.html`, `web/script.js`, `web/styles.css`) are served directly by the Go HTTP server — there is no build step for the frontend.

**Backend core components:**

- **Hub** — central WebSocket connection manager. Runs in its own goroutine with a `select` loop over `register`, `unregister`, and `broadcast` channels. Maintains a `clients` map of `*websocket.Conn → sender IP`.
- **FileStore** — in-memory file storage protected by `sync.RWMutex`. Files are stored as base64-encoded content and cleared on server restart.

**HTTP endpoints:** `/ws` (WebSocket), `POST /upload`, `GET /file/:id`, `/qr` (QR code PNG), `/api/version`, `POST /clear`, `POST /set-interval`, `POST /toggle-pause`.

**File sharing flow:** Client uploads file via `POST /upload` → backend stores in FileStore and returns a file ID → frontend sends a WebSocket message with the file ID (no content) → Hub broadcasts to all clients → each client fetches `/file/:id` to download.

**Auto-clear timer:** Hub owns a `ClearConfig` (interval, paused, nextClearTime) and communicates with three buffered channels (`clearNowCh`, `setIntervalCh`, `togglePauseCh`). A nil `<-chan time.Time` (the nil-channel trick) disables the timer without a separate flag. On any clear or config change, Hub broadcasts a `type:"clear"` or `type:"config"` WS message to all clients. New clients receive the current config immediately on connect. The default interval is **10 minutes** — the timer is started immediately in `hub.run()` when `IntervalMin > 0`.

**Connected-device count:** On every client connect or disconnect, Hub broadcasts a `type:"clients"` WS message with a `count` field to all clients. The count is the number of **unique IPs** in the clients map (via `uniqueDeviceCount`), not raw WebSocket connections — multiple connections from the same device (e.g. browser prefetch + real tab) count as one device. The register case also drains any pending unregisters before computing the count, to avoid a brief spike when a reconnect races with a stale connection's cleanup. The frontend updates the status bar to show e.g. `Connected ✅ · 3 devices`.

**Frontend (`web/script.js`):** Manages the WebSocket connection with auto-reconnect (2s interval). Own messages are filtered client-side to prevent echo via an `ownMessageIds` Set keyed by message id — a single bool would misfilter when multiple messages are in flight (e.g. sending several files at once). The update checker fetches from the GitHub API on page load to detect new releases and is platform/arch-aware. Includes an auto-clear control bar with interval selector, pause/resume button, live countdown display, and manual clear button. Both the message text and the textarea input use `dir="auto"` for automatic RTL/LTR detection (e.g. Persian, Arabic render right-to-left). New messages scroll to bottom with `messagesDiv.scrollTop = messagesDiv.scrollHeight`. **File attach** supports multi-select (`<input type="file" multiple>`); picked files append to a `selectedFiles` array and render as compact chips in `#fileAttachments` below the input row (outside the flex row, so the textarea keeps full width). On send, each file is uploaded and broadcast as its own WS message; any typed text is attached to the first file's message. **Important:** `styles.css` has chip animations (`fileChipIn`/`fileChipOut`) — attachment animations must NOT be named `slideIn`, as that would override the message entry animation and make messages start with `max-height: 0`, breaking `scrollHeight` reads.

**Versioning:** The version string is injected at build time via `-ldflags "-X main.Version=$(VERSION)"` and exposed via `/api/version`. The Docker image receives it via the `VERSION` build arg.

**Client IP resolution:** The `realIP` helper in `main.go` resolves the real client IP by checking `X-Forwarded-For` (first entry in the comma-separated list) then `X-Real-IP`, falling back to `r.RemoteAddr`. This is necessary when running behind a reverse proxy or Docker, where NAT would otherwise cause all clients to appear as the bridge gateway IP (e.g. `172.18.0.1`). The WebSocket handler uses `realIP(r)` when registering each connection.

**Docker:** A multi-stage `Dockerfile` uses `ARG GO_VERSION` and `ARG ALPINE_VERSION` to parameterise the base images (`golang:${GO_VERSION}-alpine` / `alpine:${ALPINE_VERSION}`). The canonical values live in the top-level `env:` block of `release.yml` (`GO_VERSION`, `ALPINE_VERSION`) and are passed as build args by the workflow — one place to bump either version. Build locally: `docker build --build-arg VERSION=x.y.z -t local-clipboard .`. Run: `docker run -p 8080:8080 local-clipboard`. When running via Docker Compose behind a reverse proxy, configure the proxy to set `X-Forwarded-For` or `X-Real-IP` so client IPs are preserved correctly. Because the container IP is not reachable from the LAN, set `LOCAL_CLIPBOARD_HOST` to the host's address when you want the QR code and startup logs to advertise a reachable IP.
