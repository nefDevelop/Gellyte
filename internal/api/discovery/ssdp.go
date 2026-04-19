package discovery

import (
	"fmt"
	"net"
	"strings"
)

// SSDPServer maneja el descubrimiento del servidor Gellyte en la red local.
type SSDPServer struct {
	Port     int
	ServerID string
}

const (
	multicastAddr = "239.255.255.250:1900"
)

// Start inicia el escucha de peticiones SSDP M-SEARCH.
func (s *SSDPServer) Start() {
	addr, err := net.ResolveUDPAddr("udp4", multicastAddr)
	if err != nil {
		//log.Printf("[SSDP] Error resolviendo dirección multicast: %v", err)
		return
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		//log.Printf("[SSDP] Error escuchando multicast: %v", err)
		return
	}

	//log.Printf("[SSDP] Servidor de descubrimiento iniciado en %s", multicastAddr)

	buf := make([]byte, 1024)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			//log.Printf("[SSDP] Error leyendo de UDP: %v", err)
			continue
		}

		message := string(buf[:n])
		if strings.Contains(message, "M-SEARCH") {
			// Respondería a la petición de búsqueda
			// Por simplicidad en este entorno, logueamos el intento.
			// Implementación mínima de respuesta:
			go s.respond(src)
		}
	}
}

func (s *SSDPServer) respond(dest *net.UDPAddr) {
	conn, err := net.DialUDP("udp4", nil, dest)
	if err != nil {
		return
	}
	defer conn.Close()

	// Location debe apuntar a la IP real del servidor.
	// Aquí usamos una simplificación; en producción se detectaría la IP de la interfaz.
	location := fmt.Sprintf("http://localhost:%d/System/Info/Public", s.Port)

	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"CACHE-CONTROL: max-age=1800\r\n"+
			"ST: urn:schemas-upnp-org:device:MediaServer:1\r\n"+
			"USN: uuid:%s::urn:schemas-upnp-org:device:MediaServer:1\r\n"+
			"EXT:\r\n"+
			"SERVER: Gellyte/%s\r\n"+
			"LOCATION: %s\r\n"+
			"\r\n",
		s.ServerID, "1.0.0", location,
	)

	conn.Write([]byte(response))
}
