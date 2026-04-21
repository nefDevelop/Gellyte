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

- [x] **Hito 5: Motor de Transcodificación**
    - [x] Integración con FFmpeg (Detección y Wrapper en Go).
    - [x] Lógica de decisión avanzada en `PlaybackInfo` basada en capacidades del cliente.
    - [x] Generación de perfiles de transcodificación (Video: H264, Audio: AAC/MP3).
- [x] **Hito 6: Streaming Adaptativo**
    - [x] Soporte para HLS (HTTP Live Streaming).
    - [x] Generación de segmentos dinámicos con FFmpeg.
- [x] **Hito 7: Automatización y Escaneo Profundo**
    - [x] Escáner de sistema de archivos (fsnotify).
    - [x] Extracción de metadatos técnicos avanzados (ffprobe: duración, resolución, streams).
    - [x] Soporte básico para archivos `.nfo` locales.
    - [x] Generación de capturas de pantalla (Thumbnails).

- [ ] **Hito 8: Experiencia de Usuario y Home Screen**
    - [x] Endpoint `/UserViews` para compatibilidad con Streamyfin.
    - [x] Endpoints `/Items/Latest` y `/Items/Suggestions`.
    - [x] Persistencia real de progreso en DB (Resume Items).
    - [x] Lógica de "Next Up" para series.
    - [x] Búsqueda básica por texto (`searchTerm`).

- [x] **Hito 9: Jerarquía de Series (TV Shows)**
    - [x] Detección inteligente de Estructura: Serie -> Temporada -> Episodio.
    - [x] Endpoints `/Shows/:id/Seasons` y `/Seasons/:id/Episodes`.
    - [x] Agrupamiento en la interfaz de usuario.

- [x] **Hito 10: WebSocket y Notificaciones en Tiempo Real**
    - [x] Handshake estándar compatible con RFC 6455.
    - [x] Mensajes de "LibraryChanged" automáticos.
    - [x] Mensajes de "UserDataChanged" para sincronización instantánea.

---

---

### Fase 3: Arquitectura Enterprise y Funciones Avanzadas (Gellyte v2)
*Objetivo: Convertir el MVP en un producto robusto, seguro, mantenible y listo para competir.*

- [ ] **Hito 11: Arquitectura Limpia (Clean Architecture)**
    - [ ] Refactorización a capas: Controladores (Handlers) -> Casos de Uso (Services) -> Repositorios (DB).
    - [ ] Módulo `Auth`: Extraer validación de credenciales y generación de token al servicio.
    - [ ] Módulo `Library`: Separar la lógica compleja de filtrado SQL del handler.
    - [ ] Módulo `Playback`: Aislar la lógica de ffmpeg/transcoding en su propio servicio.

- [ ] **Hito 12: Configuración y CLI Profesional**
    - [ ] Implementar lectura de configuración (`.env` o `config.yaml`) usando Viper/Godotenv.
    - [ ] Eliminar constantes "hardcodeadas" (puertos, rutas, UUIDs semilla).
    - [ ] Integrar `Cobra` para comandos de CLI (`serve`, `scan`, `user add`).

- [ ] **Hito 13: Seguridad Mejorada**
    - [ ] Sustituir hashing de passwords en texto plano/simple por `bcrypt` o `argon2`.
    - [ ] Migrar el token de sesión MD5 a JWT (JSON Web Tokens) con fecha de expiración.
    - [ ] Asegurar los endpoints con validación JWT estricta.

- [ ] **Hito 14: Scrapers y Metadatos Dinámicos**
    - [ ] Integración de APIs de terceros (TMDB / TVmaze).
    - [ ] Descarga y almacenamiento automático de pósters y fondos (Fanart).
    - [ ] Parseo y extracción profunda de metadatos (elenco, sinopsis, géneros).

- [ ] **Hito 15: Transcodificación por Hardware y Estabilidad**
    - [ ] Soporte HWA: Detección y uso de NVENC (Nvidia), QSV (Intel) o VAAPI.
    - [ ] Gestor de Procesos FFmpeg: Matar procesos "zombies" por desconexión de clientes (Graceful Shutdown).
    - [ ] Migraciones formales de base de datos (`golang-migrate`).
    - [ ] Logging Estructurado JSON (`slog` o `zap`).

### Consideraciones Técnicas Críticas
- **Seguridad:** Los tokens de sesión deben ser almacenados y validados rigurosamente.
- **Rendimiento:** Uso de **Go + SQLite con índices optimizados** para búsquedas instantáneas en bases de datos masivas.
- **Tiempo Real:** Implementación de **fsnotify** para actualizaciones automáticas de la biblioteca sin escaneos manuales.
- **Bajo Consumo:** Procesamiento asíncrono y streaming eficiente para minimizar el uso de RAM.
- **Compatibilidad:** Seguir estrictamente el esquema JSON de Jellyfin para evitar crashes en las apps clientes.
