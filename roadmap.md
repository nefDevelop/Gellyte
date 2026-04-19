# Roadmap: Gellyte (Jellyfin Compatible Server in Go)

## Proyecto: Servidor compatible con el protocolo Jellyfin usando Go, Gin y SQLite.

---

### Fase 1: El Esqueleto y el "Engaño" (MVP)
*Objetivo: Lograr que una app oficial de Jellyfin pueda conectarse, loguearse y reproducir video mediante Direct Play.*

- [x] **Hito 1: Cimiento del Proyecto**
    - [x] Inicializar módulo Go y estructura de directorios.
    - [x] Configurar Framework Gin.
    - [x] Configurar Swagger (swaggo) para documentación de API.
    - [x] Setup de SQLite con GORM para persistencia.
    - [x] Implementar `/System/Info/Public` y `/System/Info`.
- [x] **Hito 2: Autenticación y Sesiones**
    - [x] Middleware para parsing de Header `X-Emby-Authorization`.
    - [x] Implementar `/Users/AuthenticateByName` (Login).
    - [x] Implementar `/Users/Me` y `/Sessions/Capabilities`.
- [x] **Hito 3: Gestión de Biblioteca (Metadata y Tiempo Real)**
    - [x] Esquema de DB para Items (Películas/Series) optimizado con índices.
    - [x] Implementar **fsnotify** para monitoreo de carpetas en tiempo real.
    - [x] Endpoint `/Library/VirtualFolders`.
    - [x] Endpoint `/Items` (Navegación principal con paginación).
- [x] **Hito 4: Reproducción Básica (Direct Play)**
    - [x] Endpoint `/Videos/{id}/stream` (Soporte de Byte-Range / HTTP 206).
    - [x] Endpoint `/Items/{id}/PlaybackInfo` (Engañar a la app para forzar Direct Play).
    - [x] Implementar `/Sessions/Playing` y `/Playing/Progress`.
    - [x] Soporte para imágenes locales (`poster.jpg`, `folder.jpg`).

---

### Fase 2: Potencia y Transcodificación
*Objetivo: Compatibilidad universal mediante procesamiento de video en tiempo real.*

- [ ] **Hito 5: Motor de Transcodificación**
    - [ ] Integración con FFmpeg (Detección y Wrapper en Go).
    - [ ] Lógica de decisión avanzada en `PlaybackInfo` basada en capacidades del cliente.
    - [ ] Generación de perfiles de transcodificación (Video: H264, Audio: AAC/MP3).
- [ ] **Hito 6: Streaming Adaptativo**
    - [ ] Soporte para HLS (HTTP Live Streaming).
    - [ ] Generación de segmentos dinámicos con FFmpeg.
- [ ] **Hito 7: Automatización y Escaneo Profundo**
    - [x] Escáner de sistema de archivos (fsnotify).
    - [ ] Extracción de metadatos técnicos avanzados (ffprobe: duración, resolución, streams).
    - [ ] Soporte básico para archivos `.nfo` locales.
    - [ ] Generación de capturas de pantalla (Thumbnails).

- [ ] **Hito 8: Experiencia de Usuario y Home Screen**
    - [x] Endpoint `/UserViews` para compatibilidad con Streamyfin.
    - [x] Endpoints `/Items/Latest` y `/Items/Suggestions`.
    - [ ] Persistencia real de progreso en DB (Resume Items).
    - [ ] Lógica de "Next Up" para series.
    - [ ] Búsqueda global (`/Search/Hints`).

- [ ] **Hito 9: Jerarquía de Series (TV Shows)**
    - [ ] Detección inteligente de Estructura: Serie -> Temporada -> Episodio.
    - [ ] Endpoints `/Shows/:id/Seasons` y `/Seasons/:id/Episodes`.
    - [ ] Agrupamiento en la interfaz de usuario.

- [ ] **Hito 10: WebSocket y Notificaciones en Tiempo Real**
    - [x] Handshake estándar compatible con RFC 6455.
    - [ ] Mensajes de "LibraryChanged" automáticos.
    - [ ] Mensajes de "UserDataChanged" para sincronización instantánea.

---

### Consideraciones Técnicas Críticas
- **Seguridad:** Los tokens de sesión deben ser almacenados y validados rigurosamente.
- **Rendimiento:** Uso de **Go + SQLite con índices optimizados** para búsquedas instantáneas en bases de datos masivas.
- **Tiempo Real:** Implementación de **fsnotify** para actualizaciones automáticas de la biblioteca sin escaneos manuales.
- **Bajo Consumo:** Procesamiento asíncrono y streaming eficiente para minimizar el uso de RAM.
- **Compatibilidad:** Seguir estrictamente el esquema JSON de Jellyfin para evitar crashes en las apps clientes.
