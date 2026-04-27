package middleware

import (
	"bytes"
	"fmt"
	"strings"
	"time"

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
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Emby-Authorization, X-Emby-Token, X-MediaBrowser-Token, accept, origin, Cache-Control, X-Requested-With, Range, X-Emby-Device-Id, X-Emby-Device-Name, X-Emby-Client, X-Emby-Client-Version")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Emby-Token, X-Emby-Authorization, X-MediaBrowser-Token, Content-Range, Accept-Ranges, Content-Length, Content-Encoding, X-Emby-Version")
		c.Writer.Header().Set("Server", "Kestrel")
		c.Writer.Header().Set("X-Emby-Version", "10.11.8")
		c.Writer.Header().Set("X-Powered-By", "ASP.NET")

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
		case "Client":
			auth.Client = val
		case "Device":
			auth.Device = val
		case "DeviceId":
			auth.DeviceId = val
		case "Version":
			auth.Version = val
		case "Token":
			auth.Token = val
		}
	}
	return auth
}
func ResponseLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		client := "Unknown"
		device := "Unknown"
		if auth, ok := c.Get("auth"); ok {
			if a, ok := auth.(EmbyAuth); ok {
				client = a.Client
				device = a.Device
			}
		}

		if raw != "" {
			path = path + "?" + raw
		}

		// Color según el status para verlo rápido en consola
		statusColor := "\033[32m" // Verde
		if status >= 400 {
			statusColor = "\033[31m" // Rojo
		} else if status >= 300 {
			statusColor = "\033[33m" // Amarillo
		}
		reset := "\033[0m"

		fmt.Printf("[API] %s %v %s %3d %s | %13v | %15s | %s (%s)\n",
			method,
			reset,
			statusColor, status, reset,
			latency,
			c.ClientIP(),
			client,
			device,
		)
		fmt.Printf("      Path: %s\n", path)
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
