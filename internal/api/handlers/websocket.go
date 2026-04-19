package handlers

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDummySocket maneja la conexión WebSocket inicial.
// Muchos clientes de Jellyfin usan WebSockets para recibir actualizaciones en tiempo real.
func GetDummySocket(c *gin.Context) {
	// Comprobar si es un upgrade a WebSocket
	if c.GetHeader("Upgrade") != "websocket" {
		c.String(http.StatusBadRequest, "Expected WebSocket upgrade")
		return
	}

	// Realizar el "handshake" básico manualmente para no depender de librerías externas en este MVP.
	// Nota: En una app de producción usaríamos gorilla/websocket.
	// Este placeholder permite que las apps no fallen al intentar conectar.
	
	hj, ok := c.Writer.(http.Hijacker)
	if !ok {
		c.String(http.StatusInternalServerError, "Webserver doesn't support hijacking")
		return
	}
	
	conn, _, err := hj.Hijack()
	if err != nil {
		log.Printf("[WS] Error al secuestrar conexión: %v", err)
		return
	}
	defer conn.Close()

	key := c.GetHeader("Sec-WebSocket-Key")
	if key == "" {
		c.String(http.StatusBadRequest, "Sec-WebSocket-Key missing")
		return
	}

	// Calcular el header Sec-WebSocket-Accept según el estándar RFC 6455
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Responder con el cambio de protocolo siguiendo el estándar
	log.Printf("[WS] Nueva conexión desde %s", c.Request.RemoteAddr)
	
	resp := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: %s\r\n"+
		"\r\n", acceptKey)
	
	conn.Write([]byte(resp))

	// Mantener la conexión abierta (estilo dummy)
	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("[WS] Conexión cerrada: %v", err)
			break
		}
	}
}
