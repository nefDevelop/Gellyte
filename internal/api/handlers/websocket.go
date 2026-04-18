package handlers

import (
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

	// Responder con el cambio de protocolo
	log.Printf("[WS] Nueva conexión desde %s", c.Request.RemoteAddr)
	
	// Devolvemos la respuesta HTTP mínima para completar el handshake
	conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"\r\n"))

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
