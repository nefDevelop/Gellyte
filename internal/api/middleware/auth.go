package middleware

import (
	"log"
	"net/http/httputil"
	"strings"

	"github.com/gin-gonic/gin"
)

type EmbyAuth struct {
	Client   string
	Device   string
	DeviceId string
	Version  string
	Token    string
}

func EmbyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- DUMP TOTAL DE LA PETICIÓN ---
		dump, _ := httputil.DumpRequest(c.Request, true)
		log.Printf("\n--- NEW REQUEST ---\n%s\n-------------------\n", string(dump))

		authHeader := c.GetHeader("X-Emby-Authorization")
		if authHeader == "" {
			authHeader = c.GetHeader("Authorization")
		}

		if authHeader != "" && strings.Contains(authHeader, "MediaBrowser") {
			auth := parseEmbyHeader(authHeader)
			c.Set("auth", auth)
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Emby-Authorization, X-Emby-Token, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Emby-Token, X-Emby-Authorization")
		c.Writer.Header().Set("Server", "Kestrel")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func parseEmbyHeader(header string) EmbyAuth {
	auth := EmbyAuth{}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) < 2 {
		return auth
	}
	params := strings.Split(parts[1], ",")
	for _, p := range params {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		val := strings.Trim(kv[1], "\"")
		switch key {
		case "Client": auth.Client = val
		case "Device": auth.Device = val
		case "DeviceId": auth.DeviceId = val
		case "Version": auth.Version = val
		case "Token": auth.Token = val
		}
	}
	return auth
}
