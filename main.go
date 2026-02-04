package main

import (
	"embed"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed web/*
var webFS embed.FS

var Version = "dev"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local network
	},
}

type FileData struct {
	ID      string `json:"id,omitempty"` // File ID for download
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"` // base64 encoded (only stored server-side)
}

type Message struct {
	ID       string    `json:"id"`
	Text     string    `json:"text"`
	SenderIP string    `json:"senderIp,omitempty"`
	File     *FileData `json:"file,omitempty"`
}

type broadcastMsg struct {
	msg    Message
	sender *websocket.Conn
}

type FileStore struct {
	files map[string]*FileData
	mu    sync.RWMutex
}

func newFileStore() *FileStore {
	return &FileStore{
		files: make(map[string]*FileData),
	}
}

func (fs *FileStore) set(id string, file *FileData) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	// Only overwrite if the new file has content, or if it doesn't exist
	if existing, exists := fs.files[id]; !exists || file.Content != "" {
		fs.files[id] = file
	} else {
		// Keep existing file but update metadata if needed
		if file.Name != "" {
			existing.Name = file.Name
		}
		if file.Size > 0 {
			existing.Size = file.Size
		}
		if file.Type != "" {
			existing.Type = file.Type
		}
	}
}

func (fs *FileStore) get(id string) (*FileData, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	file, ok := fs.files[id]
	return file, ok
}

type connInfo struct {
	conn *websocket.Conn
	ip   string
}

type Hub struct {
	clients    map[*websocket.Conn]string // conn -> remote IP
	broadcast  chan broadcastMsg
	register   chan connInfo
	unregister chan *websocket.Conn
	fileStore  *FileStore
	mu         sync.Mutex
}

func newHub(fileStore *FileStore) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]string),
		broadcast:  make(chan broadcastMsg),
		register:   make(chan connInfo),
		unregister: make(chan *websocket.Conn),
		fileStore:  fileStore,
	}
}

func (h *Hub) run() {
	for {
		select {
		case ci := <-h.register:
			h.mu.Lock()
			h.clients[ci.conn] = ci.ip
			h.mu.Unlock()
			log.Printf("Client connected: %s. Total clients: %d", ci.ip, len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			ip := h.clients[conn]
			if ip != "" {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: %s. Total clients: %d", ip, len(h.clients))

		case bm := <-h.broadcast:
			message := bm.msg
			// Set sender IP
			h.mu.Lock()
			message.SenderIP = h.clients[bm.sender]
			h.mu.Unlock()

			// Store file if present and prepare file data for broadcast
			if message.File != nil && message.ID != "" {
				// Use file ID if provided, otherwise use message ID
				fileID := message.File.ID
				if fileID == "" {
					fileID = message.ID
				}

				// Only store file if it has content (from upload endpoint)
				// Files sent via WebSocket don't have content, so we don't overwrite existing files
				if message.File.Content != "" {
					fileToStore := &FileData{
						Name:    message.File.Name,
						Size:    message.File.Size,
						Type:    message.File.Type,
						Content: message.File.Content,
					}
					h.fileStore.set(fileID, fileToStore)
					log.Printf("Stored file during broadcast: ID=%s, Name=%s, ContentLength=%d", fileID, message.File.Name, len(message.File.Content))
				} else {
					// Check if file exists in store
					if existingFile, exists := h.fileStore.get(fileID); exists {
						log.Printf("File already exists in store: ID=%s, Name=%s, ContentLength=%d", fileID, existingFile.Name, len(existingFile.Content))
					} else {
						log.Printf("Warning: File ID %s not found in store and no content provided", fileID)
					}
				}

				// Create file data for broadcast (without content, with ID)
				// This ensures all clients get the file metadata and can download it
				fileData := &FileData{
					ID:   fileID,
					Name: message.File.Name,
					Size: message.File.Size,
					Type: message.File.Type,
				}
				message.File = fileData
			}

			h.mu.Lock()
			for conn := range h.clients {
				err := conn.WriteJSON(message)
				if err != nil {
					log.Printf("Error writing message: %v", err)
					delete(h.clients, conn)
					conn.Close()
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	h.register <- connInfo{conn: conn, ip: ip}

	defer func() {
		h.unregister <- conn
	}()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if msg.Text != "" || msg.File != nil {
			if msg.ID == "" {
				msg.ID = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			h.broadcast <- broadcastMsg{msg: msg, sender: conn}
		}
	}
}

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	// Iterate through network interfaces
	for _, iface := range ifaces {
		// Skip interfaces that are down or loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip := ipnet.IP.To4()
				if ip != nil {
					if ip.IsLoopback() {
						continue
					}
					// Skip APIPA addresses (169.254.0.0/16) - these are auto-assigned when DHCP fails
					if ip[0] == 169 && ip[1] == 254 {
						continue
					}
					// Return the first valid private IP address
					return ip.String()
				}
			}
		}
	}
	return ""
}

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	fileStore := newFileStore()
	hub := newHub(fileStore)
	go hub.run()

	// Serve static files from embedded filesystem
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		switch r.URL.Path {
		case "/":
			data, err := webFS.ReadFile("web/index.html")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(data)
		case "/styles.css":
			w.Header().Set("Content-Type", "text/css")
			data, err := webFS.ReadFile("web/styles.css")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Write(data)
		case "/script.js":
			w.Header().Set("Content-Type", "application/javascript")
			data, err := webFS.ReadFile("web/script.js")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Write(data)
		default:
			http.NotFound(w, r)
		}
	})

	// Version endpoint
	http.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(Version))
	})

	http.HandleFunc("/ws", hub.handleWebSocket)

	// File upload endpoint
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error reading file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read file content
		fileContent, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file content", http.StatusInternalServerError)
			return
		}

		// Encode to base64
		encoded := base64.StdEncoding.EncodeToString(fileContent)

		fileData := &FileData{
			Name:    header.Filename,
			Size:    int64(len(fileContent)),
			Type:    header.Header.Get("Content-Type"),
			Content: encoded,
		}

		fileID := fmt.Sprintf("%d", time.Now().UnixNano())
		fileStore.set(fileID, fileData)
		log.Printf("File uploaded: ID=%s, Name=%s, Size=%d, ContentLength=%d", fileID, fileData.Name, fileData.Size, len(fileData.Content))

		// Return file ID
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"%s","name":"%s","size":%d,"type":"%s"}`, fileID, fileData.Name, fileData.Size, fileData.Type)
	})

	// File download endpoint
	http.HandleFunc("/file/", func(w http.ResponseWriter, r *http.Request) {
		fileID := r.URL.Path[len("/file/"):]
		if fileID == "" {
			http.NotFound(w, r)
			return
		}

		file, ok := fileStore.get(fileID)
		if !ok {
			log.Printf("File not found: %s", fileID)
			http.NotFound(w, r)
			return
		}

		if file.Content == "" {
			log.Printf("File %s has no content", fileID)
			http.Error(w, "File content is empty", http.StatusInternalServerError)
			return
		}

		// Decode base64 content
		content, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			log.Printf("Error decoding file %s: %v", fileID, err)
			http.Error(w, "Error decoding file", http.StatusInternalServerError)
			return
		}

		if len(content) == 0 {
			log.Printf("File %s decoded to empty content", fileID)
			http.Error(w, "File content is empty after decoding", http.StatusInternalServerError)
			return
		}

		// Set content type, default to application/octet-stream if empty
		contentType := file.Type
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))

		written, err := w.Write(content)
		if err != nil {
			log.Printf("Error writing file %s: %v", fileID, err)
			return
		}
		if written != len(content) {
			log.Printf("Warning: incomplete write for file %s: wrote %d of %d bytes", fileID, written, len(content))
		}
		log.Printf("Successfully served file %s (%s, %d bytes)", fileID, file.Name, len(content))
	})

	addr := "0.0.0.0:" + *port

	// Get local IP address
	localIP := getLocalIP()
	if localIP != "" {
		log.Printf("Server starting on %s", addr)
		log.Printf("Open http://localhost:%s on your laptop", *port)
		log.Printf("Open http://%s:%s on your phone", localIP, *port)
	} else {
		log.Printf("Server starting on %s", addr)
		log.Printf("Open http://localhost:%s on your laptop", *port)
		log.Printf("Open http://<your-laptop-ip>:%s on your phone", *port)
	}
	log.Println("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
