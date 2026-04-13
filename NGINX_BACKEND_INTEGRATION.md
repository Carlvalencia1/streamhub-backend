# Referencia: Integración NGINX ↔ Backend Go

## 📡 FLUJO DE EVENTOS

```
┌─────────┐       ┌─────────┐       ┌────────┐
│   OBS   │──────→│  NGINX  │──────→│ Backend│
│Streamer │ RTMP  │  RTMP   │ HTTP  │   Go   │
└─────────┘       └─────────┘       └────────┘
                       │
                       ↓
                  ┌──────────┐
                  │ /var/www │
                  │   /hls   │
                  │ (output) │
                  └──────────┘
```

---

## 🔄 SECUENCIA DETALLADA

### 1️⃣ OBS Inicia Stream
```
OBS envía RTMP a: rtmp://54.144.66.251/live
Con stream_key: {abc-123-def}
```

### 2️⃣ NGINX recibe conexión RTMP
```
application live {
    on_publish http://3.232.197.126:8081/api/streams/validate-key?app=live&name={abc-123-def};
}
```

### 3️⃣ NGINX HTTP POST → Backend
```
REQUEST:
POST /api/streams/validate-key?app=live&name=abc-123-def
Host: 3.232.197.126:8081

HEADERS:
Content-Type: application/x-www-form-urlencoded
Content-Length: 0
```

### 4️⃣ Backend valida en BD
```go
// En validation_handler.go

func (h *StreamValidationHandler) ValidateKey(w http.ResponseWriter, r *http.Request) {
    streamKey := r.URL.Query().Get("name")
    
    // 1. Buscar stream en BD por stream_key
    stream, err := h.streamRepository.GetByStreamKey(ctx, streamKey)
    
    // 2. Si existe:
    stream.IsLive = true  // Marcar como en vivo
    h.streamRepository.Update(ctx, stream)
    
    // 3. Responder 200 OK (permite RTMP)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]bool{"valid": true})
}
```

### 5️⃣ Backend responde 200 OK
```
RESPONSE:
200 OK
Content-Type: application/json

{
  "valid": true
}
```

### 6️⃣ NGINX verifica respuesta
```
Si 200 OK: Permite encoding del stream RTMP → HLS
Si NO 200: Bloquea la conexión RTMP
```

### 7️⃣ NGINX genera HLS automáticamente
```
Stream RTMP → Fragmentación → HLS
                               ├── /var/www/hls/abc-123-def/index.m3u8
                               ├── /var/www/hls/abc-123-def/index-0.ts
                               ├── /var/www/hls/abc-123-def/index-1.ts
                               └── ...
```

### 8️⃣ Cliente reproduce
```
URL: http://54.144.66.251/live/abc-123-def/index.m3u8

VLC / HLS.js / ExoPlayer / AVPlayer
     ↓
GET /live/abc-123-def/index.m3u8
     ↓
Lee segmentos:
GET /live/abc-123-def/index-0.ts
GET /live/abc-123-def/index-1.ts
     ↓
Reproducción en vivo
```

### 9️⃣ OBS detiene stream
```
OBS cierra conexión RTMP
```

### 🔟 NGINX on_publish_done
```
NGINX envía:
POST /api/streams/stop?app=live&name=abc-123-def
```

### 1️⃣1️⃣ Backend marca como detenido
```go
func (h *StreamValidationHandler) StopStream(w http.ResponseWriter, r *http.Request) {
    stream.IsLive = false  // Marcar como detenido
    h.streamRepository.Update(ctx, stream)
}
```

### 1️⃣2️⃣ NGINX limpia archivos
```
Con hls_cleanup on - Elimina archivos .ts viejos
Con hls_nested on - Organiza en directorios por stream_key
```

---

## 📊 BASE DE DATOS

### Tabla: streams

```sql
CREATE TABLE streams (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url VARCHAR(255),
    category VARCHAR(100),
    owner_id VARCHAR(36),
    viewers_count INT DEFAULT 0,
    is_live BOOLEAN DEFAULT false,           ← ¡CLAVE!
    stream_key VARCHAR(36) UNIQUE NOT NULL,  ← ¡CLAVE!
    playback_url TEXT NOT NULL,
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Índices importantes:
CREATE INDEX idx_stream_key ON streams(stream_key);
CREATE INDEX idx_is_live ON streams(is_live);
```

### Para validación rápida:
```sql
-- Buscar stream por stream_key
SELECT id, is_live, user_id 
FROM streams 
WHERE stream_key = 'abc-123-def';

-- Ver streams activos
SELECT id, stream_key, user_id, created_at 
FROM streams 
WHERE is_live = true 
ORDER BY created_at DESC;
```

---

## 🔗 MÉTODOS Go NUEVOS

### En repository.go
```go
type StreamRepository interface {
    GetByStreamKey(ctx context.Context, streamKey string) (*Stream, error)
    Update(ctx context.Context, stream *Stream) error
}
```

### En stream_repository_mysql.go
```go
// Buscar por stream_key para validación
func (r *MySQLRepository) GetByStreamKey(ctx context.Context, streamKey string) (*Stream, error) {
    // Retorna el stream si existe
    // Retorna nil si no existe
    // Retorna error si hay problema de BD
}

// Actualizar stream (para marcar como live/offline)
func (r *MySQLRepository) Update(ctx context.Context, stream *Stream) error {
    sql := "UPDATE streams SET is_live = ?, updated_at = NOW() WHERE id = ?"
    // Ejecutar UPDATE
}
```

---

## 📡 ENDPOINTS SIN AUTENTICACIÓN

```
POST /api/streams/validate-key?app=live&name={stream_key}
└─ Llamada desde NGINX on_publish
└─ No requiere autenticación (es servidor → servidor)
└─ Retorna 200 OK para permitir stream

POST /api/streams/stop?app=live&name={stream_key}
└─ Llamada desde NGINX on_publish_done
└─ No requiere autenticación
└─ Idempotente (seguro llamar múltiples veces)

GET /api/streams/health
└─ Health check para NGINX upstream
└─ Retorna 200 OK si backend está vivo
```

---

## 🚀 CONFIGURACIÓN nginx.conf (Recordatorio)

```nginx
rtmp {
    server {
        listen 1935;
        application live {
            live on;
            hls on;
            hls_path /var/www/hls;
            hls_fragment 2s;
            hls_playlist_length 12s;
            hls_nested on;
            hls_cleanup on;
            
            # Webhooks sin auth
            on_publish http://3.232.197.126:8081/api/streams/validate-key;
            on_publish_done http://3.232.197.126:8081/api/streams/stop;
        }
    }
}

http {
    server {
        listen 80;
        
        # HLS playback
        location /live {
            alias /var/www/hls;
            add_header Access-Control-Allow-Origin *;
        }
    }
}
```

---

## ✅ CHECKLIST INTEGRACIÓN

- [ ] Backend Go tiene endpoint POST /api/streams/validate-key
- [ ] Backend Go tiene endpoint POST /api/streams/stop
- [ ] StreamRepository tiene método GetByStreamKey
- [ ] StreamRepository tiene método Update
- [ ] Stream domain tiene campo IsLive
- [ ] BD tabla streams tiene columna stream_key UNIQUE
- [ ] BD tabla streams tiene columna is_live
- [ ] nginx.conf tiene on_publish y on_publish_done
- [ ] NGINX escucha puerto 1935
- [ ] /var/www/hls existe con permisos correctos
- [ ] Backend y NGINX en mismo/red accesible
- [ ] Logs configurados para debugging

---

## 🧪 TESTING MANUAL

### Verificar que backend responde
```bash
curl -X POST "http://3.232.197.126:8081/api/streams/validate-key?app=live&name=test-123"
# Esperado: {"valid":true}

curl -X POST "http://3.232.197.126:8081/api/streams/stop?app=live&name=test-123"
# Esperado: {"message":"Stream stopped successfully"}
```

### Ver request que envía NGINX
```bash
# En backend, agrega logging:
log.Printf("[StreamValidation] Method: %s", r.Method)
log.Printf("[StreamValidation] URL: %s", r.RequestURI)
log.Printf("[StreamValidation] Headers: %v", r.Header)
```

### Verificar que stream_key se genera
```bash
# Cuando creas stream:
curl -X POST http://3.232.197.126:8081/api/streams \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test"}'

# Response debe tener stream_key:
# {
#   "stream_key": "xxxxxxx-yyyy-zzzz",
#   "rtmp_url": "rtmp://54.144.66.251/live/...",
#   ...
# }
```

---

## 📈 ESCALABILIDAD

**Para múltiples streamers simultáneos:**
1. ✅ Ya soporta con hls_nested on
2. ✅ Cada stream obtiene stream_key único
3. ✅ NGINX genera directorios separados
4. ✅ Backend valida cada uno independientemente

**Para múltiples servidores NGINX:**
1. Backend en BD centralizada
2. Múltiples NGINX apuntando a misma BD
3. Load balancer frente a NGINX servers
4. CDN para distribución de HLS

