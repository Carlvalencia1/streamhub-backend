# 🎥 INTEGRACIÓN WebRTC - TRANSMISIÓN EN VIVO

## ✅ IMPLEMENTACIÓN BACKEND COMPLETADA

Fecha: 2026-04-12

### Nuevos Componentes Agregados:

```
internal/platform/webrtc/
├── types.go          # Tipos de mensajes WebRTC
├── signaling.go      # Signaling Server

internal/chat/interfaces/ws/
├── webrtc_handler.go # Handler para WebRTC
```

---

## 🔌 NUEVOS WEBSOCKETS ENDPOINTS

### 1. **Transmisor (Broadcaster) - Envía video/audio**
```
WebSocket: ws://tu-servidor/ws/broadcast/:stream_id
Headers: Authorization: Bearer {JWT_TOKEN}

Uso:
- Usuario abre cámara
- Conecta a este WebSocket
- Envía SDP Offer
- Recibe SDP Answers de espectadores
- Intercambia ICE Candidates
```

### 2. **Espectador (Viewer) - Recibe video/audio**
```
WebSocket: ws://tu-servidor/ws/watch/:stream_id
Headers: Authorization: Bearer {JWT_TOKEN}

Uso:
- Usuario se conecta a este WebSocket
- Recibe confirmación si el stream está en vivo
- Envía SDP Answer
- Intercambia ICE Candidates
```

### 3. **Chat (Sin cambios - ya existía)**
```
WebSocket: ws://tu-servidor/ws/chat/:stream_id
Mensaje de chat en tiempo real
```

---

## 📨 TIPOS DE MENSAJES WEBRTC

### Transmisor → Backend
```json
{
  "type": "offer_sdp",
  "stream_id": "uuid-123",
  "from_user_id": "user-id",
  "sdp": "v=0\r\n..."
}
```

### Backend → Espectadores (mediante WebSocket watch)
```json
{
  "type": "offer_sdp",
  "stream_id": "uuid-123",
  "from_user_id": "broadcaster-id",
  "sdp": "v=0\r\n..."
}
```

### ICE Candidates
```json
{
  "type": "ice_candidate",
  "stream_id": "uuid-123",
  "from_user_id": "user-id",
  "candidate": {
    "candidate": "candidate:...",
    "sdp_m_line_index": 0,
    "sdp_mid": "0"
  }
}
```

---

## 🔄 FLUJO DE CONEXIÓN

### Escenario: Usuario A transmite, Usuario B ve

```
1. USER A (iOS): Abre cámara
   ↓
2. USER A: Conecta a ws://server/ws/broadcast/stream-123
   ↓
3. SERVER: Crea BroadcastSession para stream-123
   ↓
4. USER A: Envía SDP Offer
   ↓
5. SERVER: Almacena Offer (para nuevos spect)
   ↓
6. USER B: Conecta a ws://server/ws/watch/stream-123
   ↓
7. SERVER: Envía SDP Offer de USER A a USER B
   ↓
8. USER B: Recibe Offer, crea Peer Connection
   ↓
9. USER B: Envía SDP Answer
   ↓
10. SERVER: Envía Answer a USER A
    ↓
11. USER A & USER B: Intercambian ICE Candidates
    ↓
12. CONEXIÓN P2P ESTABLECIDA
    ↓
13. Video/Audio viaja directo entre USER A y USER B
    ↓
14. Chat en ws://server/ws/chat/stream-123 (opcional)
```

---

## 📱 MÉTODOS DEL SIGNALING SERVER

```go
// Iniciar broadcast
signalingServer.StartBroadcast(streamID, broadcasterID)

// Detener broadcast
signalingServer.StopBroadcast(streamID)

// Registrar espectador
signalingServer.RegisterViewer(streamID, viewerID)

// Desregistrar espectador
signalingServer.UnregisterViewer(streamID, viewerID)

// Procesar mensaje de signaling
signalingServer.ProcessSignalingMessage(message)

// Verificar si está transmitiendo
signalingServer.IsBroadcasting(streamID) // bool

// Obtener ID del transmisor
signalingServer.GetBroadcaster(streamID) // (string, bool)

// Contar espectadores
signalingServer.GetViewerCount(streamID) // int
```

---

## 🚀 PRÓXIMOS PASOS (Frontend Android)

El frontend Android necesita:

1. **Captura de cámara + micrófono**
   - Usar CameraX + AudioRecord

2. **Librería WebRTC**
   - `webrtc-android` SDK

3. **Conexión WebSocket**
   - Para signaling

4. **Peer Connection**
   - Crear y gestionar conexiones P2P

5. **UI**
   - Pantalla de brodcast (captura cámara)
   - Pantalla de viewer (reproduce video)

---

## 🔐 SEGURIDAD

✅ Todos los endpoints requieren JWT  
✅ Solo el propietario del stream puede transmitir  
✅ Los espectadores solo pueden recibir  
✅ ICE candidates validadas  
✅ Conexiones HTTPS/WSS en producción  

---

## 🧪 TESTING (CURL - NO RECOMENDADO)

No se puede testear WebRTC con curl porque requiere navegador o app móvil con WebRTC support.

**Para testear:**
- Usar Android Emulator o dispositivo real
- Usar navegador con WebRTC (para web después)
- Usar aplicación de testing como Janus o Mediasoup

---

## 📊 ARQUITECTURA

```
┌─────────────────┐
│  Android App    │ (Broadcaster)
│  + Camera       │
│  + WebRTC       │
└────────┬────────┘
         │ WebSocket: /ws/broadcast/stream-id
         │ SDP Offer, ICE Candidates
         ▼
┌─────────────────┐
│  SignalingServer│ (En tu Go Backend)
│  - Manage Peers │
│  - Relay SDP    │
│  - Relay ICE    │
└────────┬────────┘
         │ WebSocket: /ws/watch/stream-id
         │ SDP Offer, ICE Candidates
         ▼
┌─────────────────┐
│  Android App    │ (Viewer)
│  + WebRTC       │
│  + ExoPlayer    │
└─────────────────┘

┌─────────────────┐
│   NGINX + HLS   │ (Fallback si WebRTC falla)
│   - Recibe RTMP │
│   - Genera HLS  │
└─────────────────┘
```

---

## ⚠️ LIMITACIONES ACTUALES

1. **Escalabilidad**: Soporta múltiples viewers por broadcaster
2. **P2P directo**: Requiere que navegador/app soporte NAT traversal
3. **Servidores TURN**: Recomendado agregar para usuarios con firewalls
4. **Recuperación**: Si conexión falla, hay que reconectar desde 0
5. **Estadísticas**: No hay logging de bitrate/latencia (próximo paso)

---

## 🎯 MEJORAS FUTURAS

- [ ] Agregar servidor TURN para NAT traversal
- [ ] Logging de estadísticas WebRTC (bitrate, latency)
- [ ] Reconexión automática
- [ ] Simulcast (múltiples bitrates)
- [ ] Recording de streams
- [ ] Dashboard de monitoring
