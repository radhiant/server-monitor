package websocket

import (
	"sync"

	"github.com/rs/zerolog/log"

	"server-monitor/backend/internal/model"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan model.WebSocketMessage
	register   chan *Client
	unregister chan *Client
	quit       chan struct{}
	initial    func() any // Returns full snapshot for newly connected clients
}

// NewHub creates a new Hub instance.
// The initial callback, if non-nil, is invoked for every new client to provide
// a full data snapshot before the client starts receiving incremental broadcasts.
func NewHub(initial func() any) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan model.WebSocketMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		quit:       make(chan struct{}),
		initial:    initial,
	}
}

// Run executes the hub event loop.
func (h *Hub) Run() {
	for {
		select {
		case <-h.quit:
			h.mu.Lock()
			for client := range h.clients {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			count := len(h.clients)
			h.mu.Unlock()
			log.Info().Int("total_clients", count).Msg("websocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Info().Int("total_clients", len(h.clients)).Msg("websocket client disconnected")
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Slow client buffer full: queue for removal
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all active clients.
func (h *Hub) Broadcast(msg model.WebSocketMessage) {
	select {
	case h.broadcast <- msg:
	default:
		// Queue full, drop rather than blocking collector
	}
}

// Stop closes the hub and terminates all client connections.
func (h *Hub) Stop() {
	close(h.quit)
}
