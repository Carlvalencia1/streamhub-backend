# Ejemplos de Código: NGINX RTMP Integration

## 🎯 Ejemplo 1: CreateStream Application

Este es el flujo cuando se crea un stream desde la API.

### create_stream.go (Application Layer)

```go
package application

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"streamhub/internal/streams/domain"
)

type CreateStream struct {
	repository domain.StreamRepository
}

func NewCreateStream(repo domain.StreamRepository) *CreateStream {
	return &CreateStream{repository: repo}
}

// Execute crea un nuevo stream
// IMPORTANTE: Genera stream_key único para NGINX validation
func (uc *CreateStream) Execute(ctx context.Context, 
	title, description string) (*domain.Stream, error) {

	// Generar IDs únicos
	streamID := uuid.New().String()
	streamKey := uuid.New().String() // ← Esto envía a NGINX

	stream := &domain.Stream{
		ID:           streamID,
		Title:        title,
		Description:  description,
		StreamKey:    streamKey,
		PlaybackURL:  fmt.Sprintf("http://54.144.66.251/live/%s/index.m3u8", streamKey),
		IsLive:       false, // No está en vivo hasta que NGINX valide
		ViewersCount: 0,
		CreatedAt:    time.Now(),
	}

	// Guardar en BD
	if err := uc.repository.Create(ctx, stream); err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return stream, nil
}
```

---

## 🎯 Ejemplo 2: Validation Handler (NGINX Webhook)

Este es el codigo que ya pusimos en tu proyecto.

### validation_handler.go (Recibe llamadas de NGINX)

```go
package http

import (
	"encoding/json"
	"log"
	"net/http"
	"streamhub/internal/streams/domain"
)

type StreamValidationHandler struct {
	streamRepository domain.StreamRepository
}

// ValidateKey es llamado por NGINX cuando OBS conecta
// URL: POST /api/streams/validate-key?app=live&name={stream_key}
// Si retorna 200 OK → OBS puede transmitir
// Si retorna 4xx/5xx → OBS desconecta
func (h *StreamValidationHandler) ValidateKey(
	w http.ResponseWriter, r *http.Request) {

	// NGINX envía: ?app=live&name=123-abc-def
	streamKey := r.URL.Query().Get("name")
	app := r.URL.Query().Get("app")

	log.Printf("[ValidateKey] Received: app=%s, name=%s", app, streamKey)

	// 1. Buscar stream en BD por stream_key
	stream, err := h.streamRepository.GetByStreamKey(r.Context(), streamKey)
	if err != nil {
		log.Printf("[ValidateKey] BD error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 2. Si no existe, rechazar
	if stream == nil {
		log.Printf("[ValidateKey] Stream not found: %s", streamKey)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Stream key not found",
		})
		return
	}

	// 3. Marcar como LIVE
	stream.IsLive = true
	if err := h.streamRepository.Update(r.Context(), stream); err != nil {
		log.Printf("[ValidateKey] Update error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 4. Responder 200 OK (permite NSGI/OBS)
	log.Printf("[ValidateKey] ✓ Validated: %s", streamKey)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"valid": true})
}

// StopStream es llamado por NGINX cuando OBS desconecta
// URL: POST /api/streams/stop?app=live&name={stream_key}
func (h *StreamValidationHandler) StopStream(
	w http.ResponseWriter, r *http.Request) {

	streamKey := r.URL.Query().Get("name")

	log.Printf("[StopStream] Received: name=%s", streamKey)

	// 1. Buscar stream
	stream, err := h.streamRepository.GetByStreamKey(r.Context(), streamKey)
	if err != nil || stream == nil {
		log.Printf("[StopStream] Stream not found: %s", streamKey)
		// Retornar 200 OK igual (idempotente)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. Marcar como NOT LIVE
	stream.IsLive = false
	if err := h.streamRepository.Update(r.Context(), stream); err != nil {
		log.Printf("[StopStream] Error updating: %v", err)
	}

	// 3. Responder 200 OK
	log.Printf("[StopStream] ✓ Stopped: %s", streamKey)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Stream stopped",
	})
}
```

---

## 🎯 Ejemplo 3: Repository Implementation

### stream_repository_mysql.go (GetByStreamKey y Update)

```go
package infrastructure

import (
	"context"
	"database/sql"
	"streamhub/internal/streams/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

// GetByStreamKey busca un stream por su stream_key
// Usado por NGINX validation (ValidateKey)
func (r *MySQLRepository) GetByStreamKey(
	ctx context.Context, streamKey string) (*domain.Stream, error) {

	query := `
	SELECT id, title, description, stream_key, playback_url,
	       is_live, viewers_count, owner_id, created_at
	FROM streams
	WHERE stream_key = ?
	LIMIT 1
	`

	var stream domain.Stream
	var createdAt string

	err := r.db.QueryRowContext(ctx, query, streamKey).Scan(
		&stream.ID,
		&stream.Title,
		&stream.Description,
		&stream.StreamKey,
		&stream.PlaybackURL,
		&stream.IsLive,
		&stream.ViewersCount,
		&stream.OwnerID,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No encontrado, pero no es error
	}
	if err != nil {
		return nil, err // Error real
	}

	// Parsear fecha
	parsedTime, _ := time.Parse("2006-01-02 15:04:05", createdAt)
	stream.CreatedAt = parsedTime

	return &stream, nil
}

// Update actualiza un stream existente
// Usado por NGINX validation (ValidateKey y StopStream)
func (r *MySQLRepository) Update(
	ctx context.Context, stream *domain.Stream) error {

	query := `
	UPDATE streams
	SET is_live = ?,
	    viewers_count = ?,
	    title = ?,
	    description = ?,
	    updated_at = NOW()
	WHERE id = ?
	LIMIT 1
	`

	result, err := r.db.ExecContext(ctx, query,
		stream.IsLive,
		stream.ViewersCount,
		stream.Title,
		stream.Description,
		stream.ID,
	)

	if err != nil {
		return err
	}

	// Verificar que se actualizó
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows // Stream no encontrado
	}

	return nil
}
```

---

## 🎯 Ejemplo 4: HTTP Handler (Response al cliente)

### handler.go (GET stream con playback_url)

```go
package http

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createStreamUC    *application.CreateStream
	getStreamsUC      *application.GetStreams
	getStreamByIDUC   *application.GetStreamByID
	// ... más use cases
}

// Create crea un nuevo stream
// POST /api/streams
// Response: incluye stream_key, rtmp_url, playback_url
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Crear stream (genera stream_key)
	stream, err := h.createStreamUC.Execute(
		c.Request.Context(),
		req.Title,
		req.Description,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Responder con datos para OBS
	c.JSON(http.StatusCreated, gin.H{
		"id":           stream.ID,
		"title":        stream.Title,
		"stream_key":   stream.StreamKey, // ← OBS usa esto
		"rtmp_url":     "rtmp://54.144.66.251/live/" + stream.StreamKey, // ← OBS usa esto
		"playback_url": stream.PlaybackURL, // ← Cliente web/app usa esto
	})
}

// GetByID obtiene un stream específico
// GET /api/streams/:id
// Response: incluye playback_url para reproducción
func (h *Handler) GetByID(c *gin.Context) {
	streamID := c.Param("id")

	stream, err := h.getStreamByIDUC.Execute(c.Request.Context(), streamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}

	c.JSON(http.StatusOK, stream)
}
```

---

## 🎯 Ejemplo 5: Routes Registration

### routes.go (Registrar endpoints)

```go
package http

import (
	"github.com/gin-gonic/gin"
	"streamhub/internal/platform/middleware"
)

// RegisterRoutes conecta todos los endpoints
func RegisterRoutes(
	r *gin.RouterGroup,
	handler *Handler,
	validationHandler *StreamValidationHandler) {

	streams := r.Group("/streams")

	// Públicos
	streams.GET("/", handler.GetAll)
	streams.GET("/:id", handler.GetByID)

	// Webhooks NGINX (sin auth)
	streams.POST("/validate-key", validationHandler.ValidateKey)
	streams.POST("/stop", validationHandler.StopStream)

	// Protegidos (requieren JWT)
	protected := streams.Group("/")
	protected.Use(middleware.AuthMiddleware())

	protected.POST("/", handler.Create)              // Crear stream
	protected.PUT("/:id/start", handler.Start)       // Iniciar (deprecated)
	protected.PUT("/:id/stop", handler.Stop)         // Parar (deprecated)
	protected.POST("/:id/join", handler.Join)        // Unirse
}
```

---

## 🎯 Ejemplo 6: Router Initialization

### router.go (Server setup)

```go
package server

import (
	"database/sql"
	"github.com/gin-gonic/gin"

	streamsInfra "streamhub/internal/streams/infrastructure"
	streamsApp "streamhub/internal/streams/application"
	streamsHTTP "streamhub/internal/streams/interfaces/http"
)

func RegisterRoutes(r *gin.Engine, db *sql.DB) {
	api := r.Group("/api")

	// ===== STREAMS MODULE =====

	// Crear repository
	streamsRepo := streamsInfra.NewMySQLRepository(db)

	// Crear use cases
	createStreamUC := streamsApp.NewCreateStream(streamsRepo)
	getStreamsUC := streamsApp.NewGetStreams(streamsRepo)
	getStreamByIDUC := streamsApp.NewGetStreamByID(streamsRepo)
	startStreamUC := streamsApp.NewStartStream(streamsRepo)
	stopStreamUC := streamsApp.NewStopStream(streamsRepo)
	joinStreamUC := streamsApp.NewJoinStream(streamsRepo)

	// Crear handlers
	handler := streamsHTTP.NewHandler(
		createStreamUC,
		getStreamsUC,
		getStreamByIDUC,
		startStreamUC,
		stopStreamUC,
		joinStreamUC,
	)

	// ← NUEVO: Crear validation handler (para NGINX)
	validationHandler := streamsHTTP.NewStreamValidationHandler(streamsRepo)

	// Registrar rutas
	streamsHTTP.RegisterRoutes(api, handler, validationHandler)
}
```

---

## 🔄 Flujo Completo: Paso a Paso

### 1. Cliente Android crea stream
```bash
POST /api/streams
{
  "title": "Mi streaming",
  "description": "En vivo ahora"
}

Response:
{
  "id": "uuid-123",
  "title": "Mi streaming",
  "stream_key": "key-456",
  "rtmp_url": "rtmp://54.144.66.251/live/key-456",
  "playback_url": "http://54.144.66.251/live/key-456/index.m3u8"
}
```

### 2. OBS/Streamer usa stream_key
```
Settings:
- Server: rtmp://54.144.66.251/live
- Stream Key: key-456

Click: Start Streaming
```

### 3. NGINX recibe conexión RTMP
```
127.0.0.1 → rtmp://NGINX/live/key-456
```

### 4. NGINX valida con tu backend
```
NGINX ejecuta:
POST http://backend:8081/api/streams/validate-key?app=live&name=key-456

Backend:
1. GetByStreamKey(ctx, "key-456") → Encuentra stream
2. stream.IsLive = true
3. repository.Update(ctx, stream)
4. Responde 200 OK
```

### 5. NGINX genera HLS
```
RTMP stream → Fragmentación HLS
                ├── /var/www/hls/key-456/index.m3u8
                ├── /var/www/hls/key-456/index-0.ts
                └── /var/www/hls/key-456/index-1.ts
```

### 6. Cliente obtiene playback_url
```bash
GET /api/streams/uuid-123
Response:
{
  ...
  "playback_url": "http://54.144.66.251/live/key-456/index.m3u8"
}
```

### 7. Cliente reproduce
```
VLC/HLS.js abre:
http://54.144.66.251/live/key-456/index.m3u8

NGINX sirve:
- index.m3u8 (playlist)
- index-0.ts (video/audio)
- index-1.ts (video/audio)
```

### 8. OBS detiene
```
Click: Stop Streaming
```

### 9. NGINX marca como parado
```
NGINX ejecuta:
POST http://backend:8081/api/streams/stop?app=live&name=key-456

Backend:
1. GetByStreamKey(ctx, "key-456")
2. stream.IsLive = false
3. repository.Update(ctx, stream)
```

### 10. NGINX limpia
```
hls_cleanup elimina archivos .ts viejos
Directory queda limpio para próximos streams
```

---

## 📊 Estado de BD

### Antes (stream creado)
```sql
SELECT id, stream_key, is_live FROM streams WHERE id='uuid-123';
-- uuid-123 | key-456 | false (en espera)
```

### Durante (OBS transmitiendo)
```sql
-- uuid-123 | key-456 | true (en vivo)
```

### Después (OBS desconectó)
```sql
-- uuid-123 | key-456 | false (parado)
```

---

## 🔒 Seguridad

### ValidateKey solo debe estar en /api/streams
```go
// ✅ CORRECTO: Sin autenticación
streams.POST("/validate-key", validationHandler.ValidateKey)

// ❌ INCORRECTO: Bajo protección
protected.POST("/validate-key", validationHandler.ValidateKey)
// NGINX no puede autenticarse
```

### Validación en BD
```go
// ✅ CORRECTO: Validar stream_key existe
stream, _ := repo.GetByStreamKey(ctx, streamKey)
if stream == nil {
    return 401 Unauthorized
}

// ❌ INCORRECTO: Permitir cualquier stream_key
// (cualquiera podría streamear)
```

---

## ✅ Testing Manual

### 1. Crear stream
```bash
STREAM=$(curl -X POST http://localhost:8081/api/streams \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test"}')

KEY=$(echo $STREAM | jq -r '.stream_key')
echo "Stream key: $KEY"
```

### 2. Validar desde NGINX
```bash
curl -X POST "http://localhost:8081/api/streams/validate-key?app=live&name=$KEY"
# Debe retornar: {"valid":true}
```

### 3. Enviar stream
```bash
ffmpeg -f lavfi -i testsrc=s=1280x720:d=10 \
        -f lavfi -i sine \
        -c:v libx264 -b:v 2500k -preset veryfast \
        -c:a aac -b:a 128k \
        -rtmp_live live \
        rtmp://localhost/live/$KEY
```

### 4. Ver archivos
```bash
ls -la /var/www/hls/$KEY/
# Debe ver: index.m3u8, index-0.ts, etc.
```

### 5. Reproducir
```bash
curl http://localhost/live/$KEY/index.m3u8
# Verá líneas M3U8
```

