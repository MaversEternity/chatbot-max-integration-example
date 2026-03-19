package mockserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MockServer struct {
	mu          sync.Mutex
	updates     []json.RawMessage
	marker      int64
	nextMid     int64
	nextCbID    int64
	wsClients   []*websocket.Conn
	wsMu        sync.Mutex
	updateReady chan struct{}
}

func New() *MockServer {
	return &MockServer{
		updates:     make([]json.RawMessage, 0),
		marker:      0,
		nextMid:     1000,
		nextCbID:    1,
		updateReady: make(chan struct{}, 100),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *MockServer) Start(addr string) error {
	mux := http.NewServeMux()

	// Bot API endpoints
	mux.HandleFunc("/me", s.handleGetMe)
	mux.HandleFunc("/updates", s.handleGetUpdates)
	mux.HandleFunc("/messages", s.handleMessages)
	mux.HandleFunc("/answers", s.handleAnswers)
	mux.HandleFunc("/uploads", s.handleUploads)

	// WebSocket for client simulator
	mux.HandleFunc("/ws", s.handleWebSocket)

	// REST API for mini app
	mux.HandleFunc("/api/tasks", s.handleAPITasks)

	// Serve mini app (React build)
	mux.Handle("/miniapp/", http.StripPrefix("/miniapp/", http.FileServer(http.Dir("web/miniapp"))))

	// Serve static web files (client simulator)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Printf("Mock MAX API server on %s", addr)
	log.Printf("Client simulator: http://localhost%s/client.html", addr)
	log.Printf("Mini app: http://localhost%s/miniapp/", addr)
	return http.ListenAndServe(addr, mux)
}

// GET /me
func (s *MockServer) handleGetMe(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"user_id":   int64(999),
		"name":      "БухКомпания Финлид",
		"username":  "finlid_bot",
		"commands":  []map[string]string{{"name": "/start", "description": "Начать работу"}},
		"description": "Бот бухгалтерской компании Финлид",
	}
	writeJSON(w, resp)
}

// GET /updates — long polling
func (s *MockServer) handleGetUpdates(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if len(s.updates) > 0 {
		updates := s.updates
		s.updates = make([]json.RawMessage, 0)
		s.marker++
		marker := s.marker
		s.mu.Unlock()
		writeJSON(w, map[string]interface{}{
			"updates": updates,
			"marker":  marker,
		})
		return
	}
	s.mu.Unlock()

	// Wait for updates or timeout
	select {
	case <-s.updateReady:
		s.mu.Lock()
		updates := s.updates
		s.updates = make([]json.RawMessage, 0)
		s.marker++
		marker := s.marker
		s.mu.Unlock()
		writeJSON(w, map[string]interface{}{
			"updates": updates,
			"marker":  marker,
		})
	case <-time.After(25 * time.Second):
		writeJSON(w, map[string]interface{}{
			"updates": []interface{}{},
			"marker":  nil,
		})
	case <-r.Context().Done():
		return
	}
}

// POST /messages — bot sends a message
func (s *MockServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// GET /messages — return empty list (SDK may call this)
		writeJSON(w, map[string]interface{}{"messages": []interface{}{}})
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	s.mu.Lock()
	mid := fmt.Sprintf("mid_%d", s.nextMid)
	s.nextMid++
	s.mu.Unlock()

	text, _ := body["text"].(string)
	attachments, _ := body["attachments"].([]interface{})

	// Parse chat_id as int64 — SDK expects numeric, not string
	var chatID int64 = 1
	if cid := r.URL.Query().Get("chat_id"); cid != "" {
		fmt.Sscanf(cid, "%d", &chatID)
	}

	// Build response message
	respMsg := map[string]interface{}{
		"sender": map[string]interface{}{
			"user_id":  int64(999),
			"name":     "БухКомпания Финлид",
			"username": "finlid_bot",
			"is_bot":   true,
		},
		"recipient": map[string]interface{}{
			"chat_id":   chatID,
			"chat_type": "dialog",
			"user_id":   int64(100),
		},
		"timestamp": time.Now().Unix(),
		"body": map[string]interface{}{
			"mid":         mid,
			"seq":         1,
			"text":        text,
			"attachments": attachments,
		},
	}

	// Push to WebSocket clients
	log.Printf("[mock] bot -> POST /messages: text=%q mid=%s", text, mid)
	wsMsg := map[string]interface{}{
		"type":    "bot_message",
		"text":    text,
		"mid":     mid,
		"buttons": extractButtons(attachments),
	}
	s.broadcastWS(wsMsg)

	writeJSON(w, map[string]interface{}{"message": respMsg})
}

// POST /answers — callback answer
func (s *MockServer) handleAnswers(w http.ResponseWriter, r *http.Request) {
	log.Printf("[mock] bot -> POST /answers")
	writeJSON(w, map[string]interface{}{"success": true})
}

// POST /uploads
func (s *MockServer) handleUploads(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"url":   "http://localhost:8081/upload-target",
		"token": "fake-upload-token",
	})
}

// WebSocket handler for client simulator
func (s *MockServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	s.wsMu.Lock()
	s.wsClients = append(s.wsClients, conn)
	s.wsMu.Unlock()

	defer func() {
		conn.Close()
		s.wsMu.Lock()
		for i, c := range s.wsClients {
			if c == conn {
				s.wsClients = append(s.wsClients[:i], s.wsClients[i+1:]...)
				break
			}
		}
		s.wsMu.Unlock()
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		msgType, _ := msg["type"].(string)
		switch msgType {
		case "message":
			text, _ := msg["text"].(string)
			s.enqueueMessageUpdate(text)
		case "callback":
			payload, _ := msg["payload"].(string)
			s.enqueueCallbackUpdate(payload)
		case "file":
			filename, _ := msg["filename"].(string)
			s.enqueueFileUpdate(filename)
		}
	}
}

// REST API for mini app
func (s *MockServer) handleAPITasks(w http.ResponseWriter, r *http.Request) {
	// This is a simplified endpoint — in production, the mini app would talk to FinLid backend
	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeJSON(w, map[string]interface{}{
		"message": "Tasks API — connect to your backend here",
	})
}

// Enqueue a message_created update
func (s *MockServer) enqueueMessageUpdate(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mid := fmt.Sprintf("user_mid_%d", s.nextMid)
	s.nextMid++

	update := map[string]interface{}{
		"update_type": "message_created",
		"timestamp":   time.Now().Unix(),
		"message": map[string]interface{}{
			"sender": map[string]interface{}{
				"user_id":  int64(100),
				"name":     "Иван Петрович",
				"username": "ivan_p",
				"is_bot":   false,
			},
			"recipient": map[string]interface{}{
				"chat_id":   int64(1),
				"chat_type": "dialog",
				"user_id":   int64(999),
			},
			"timestamp": time.Now().Unix(),
			"body": map[string]interface{}{
				"mid":         mid,
				"seq":         1,
				"text":        text,
				"attachments": []interface{}{},
			},
		},
	}

	data, _ := json.Marshal(update)
	s.updates = append(s.updates, data)
	log.Printf("[mock] client -> message enqueued: %q", text)
	s.notifyUpdate()
}

// Enqueue a message_callback update
func (s *MockServer) enqueueCallbackUpdate(payload string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cbID := fmt.Sprintf("cb_%d", s.nextCbID)
	s.nextCbID++

	update := map[string]interface{}{
		"update_type": "message_callback",
		"timestamp":   time.Now().Unix(),
		"callback": map[string]interface{}{
			"timestamp":   time.Now().Unix(),
			"callback_id": cbID,
			"payload":     payload,
			"user": map[string]interface{}{
				"user_id":  int64(100),
				"name":     "Иван Петрович",
				"username": "ivan_p",
				"is_bot":   false,
			},
		},
		"message": map[string]interface{}{
			"sender": map[string]interface{}{
				"user_id": int64(999),
				"name":    "БухКомпания Финлид",
				"is_bot":  true,
			},
			"recipient": map[string]interface{}{
				"chat_id":   int64(1),
				"chat_type": "dialog",
				"user_id":   int64(100),
			},
			"timestamp": time.Now().Unix(),
			"body": map[string]interface{}{
				"mid":         "msg_0",
				"seq":         1,
				"text":        "",
				"attachments": []interface{}{},
			},
		},
	}

	data, _ := json.Marshal(update)
	s.updates = append(s.updates, data)
	s.notifyUpdate()
}

// Enqueue a message with file attachment
func (s *MockServer) enqueueFileUpdate(filename string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mid := fmt.Sprintf("user_mid_%d", s.nextMid)
	s.nextMid++

	update := map[string]interface{}{
		"update_type": "message_created",
		"timestamp":   time.Now().Unix(),
		"message": map[string]interface{}{
			"sender": map[string]interface{}{
				"user_id":  int64(100),
				"name":     "Иван Петрович",
				"username": "ivan_p",
				"is_bot":   false,
			},
			"recipient": map[string]interface{}{
				"chat_id":   int64(1),
				"chat_type": "dialog",
				"user_id":   int64(999),
			},
			"timestamp": time.Now().Unix(),
			"body": map[string]interface{}{
				"mid": mid,
				"seq": 1,
				"text": fmt.Sprintf("📎 %s", filename),
				"attachments": []interface{}{
					map[string]interface{}{
						"type":     "file",
						"payload":  map[string]interface{}{"token": "fake-token"},
						"filename": filename,
						"size":     1024,
					},
				},
			},
		},
	}

	data, _ := json.Marshal(update)
	s.updates = append(s.updates, data)
	s.notifyUpdate()
}

func (s *MockServer) notifyUpdate() {
	select {
	case s.updateReady <- struct{}{}:
	default:
	}
}

func (s *MockServer) broadcastWS(msg interface{}) {
	data, _ := json.Marshal(msg)
	s.wsMu.Lock()
	defer s.wsMu.Unlock()
	for _, c := range s.wsClients {
		c.WriteMessage(websocket.TextMessage, data)
	}
}

func extractButtons(attachments []interface{}) [][]map[string]interface{} {
	if attachments == nil {
		return nil
	}
	for _, a := range attachments {
		att, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		if att["type"] == "inline_keyboard" {
			payload, ok := att["payload"].(map[string]interface{})
			if !ok {
				continue
			}
			rows, ok := payload["buttons"].([]interface{})
			if !ok {
				continue
			}
			var result [][]map[string]interface{}
			for _, row := range rows {
				rowArr, ok := row.([]interface{})
				if !ok {
					continue
				}
				var btnRow []map[string]interface{}
				for _, btn := range rowArr {
					b, ok := btn.(map[string]interface{})
					if ok {
						btnRow = append(btnRow, b)
					}
				}
				result = append(result, btnRow)
			}
			return result
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
