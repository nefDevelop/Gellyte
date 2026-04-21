package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Permitir todos los orígenes para entornos locales
	},
}

// Client representa una conexión activa de WebSocket
type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

// Hub gestiona el conjunto de clientes activos
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

var GlobalHub = Hub{
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Clients:    make(map[*Client]bool),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			//log.Printf("[WS] Cliente registrado. Total: %d", len(h.Clients))
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			//log.Printf("[WS] Cliente desconectado. Total: %d", len(h.Clients))
		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// GetDummySocket ahora es funcional usando Gorilla WebSocket
func GetDummySocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		//log.Printf("[WS] Error al actualizar a WebSocket: %v", err)
		return
	}

	client := &Client{Conn: conn, Send: make(chan []byte, 256)}
	GlobalHub.Register <- client

	// Lector: Para manejar el cierre de la conexión
	go func() {
		defer func() {
			GlobalHub.Unregister <- client
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	// Escritor: Envía mensajes del canal Send al WebSocket real
	for message := range client.Send {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

// NotifyLibraryChanged envía un mensaje a todos los clientes conectados
func NotifyLibraryChanged() {
	// Formato de Jellyfin para LibraryChanged
	msg := []byte(`{"MessageType":"LibraryChanged"}`)
	GlobalHub.Broadcast <- msg
}

// NotifyUserDataChanged avisa que el progreso de un item ha cambiado
func NotifyUserDataChanged(userId string, itemId string) {
	// Formato de Jellyfin para UserDataChanged
	msg := fmt.Sprintf(`{"MessageType":"UserDataChanged","Data":{"UserId":"%s","UserDataList":[{"ItemId":"%s"}]}}`, userId, itemId)
	GlobalHub.Broadcast <- []byte(msg)
}
