# Quick Start - Firebase Cloud Messaging

## ⚡ 5 Minutos para Empezar

### Paso 1: Configurar Firebase
```bash
# 1. Descargar credenciales de Firebase Console
# 2. Guardar en: /secure/path/serviceAccountKey.json
# 3. En .env agregar:
FIREBASE_CREDENTIALS_PATH=/secure/path/serviceAccountKey.json
```

### Paso 2: Ejecutar Migraciones
```bash
# Las migraciones corren automáticamente en el start del server
# Las rutas de notificaciones ya están registradas
go run cmd/api/main.go
```

### Paso 3: Registrar Token desde Android

```bash
# Primero, obtener token JWT
JWT_TOKEN="..."

# Obtener FCM token desde Android
FCM_TOKEN="..."

# Registrar
curl -X POST http://localhost:8080/api/notifications/fcm-token \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "'$FCM_TOKEN'",
    "platform": "android"
  }'
```

### Paso 4: Iniciar Stream

```bash
# Cuando inicides un stream, automáticamente se enviarán notificaciones
curl -X PUT http://localhost:8080/api/streams/{stream_id}/start \
  -H "Authorization: Bearer $JWT_TOKEN"

# ✅ Notificación enviada a todos los usuarios
# ✅ Android recibe: type="stream_live", stream_id="{id}"
```

## 📦 Funcionalidades Incluidas

- ✅ Registrar tokens FCM (upsert)
- ✅ Eliminar tokens FCM
- ✅ Notificación automática al iniciar stream
- ✅ Soporte multicast (múltiples dispositivos)
- ✅ Manejo de tokens inválidos
- ✅ Base de datos MySQL
- ✅ Logging seguro
- ✅ JWT authentication
- ✅ Tests unitarios

## 📊 Base de Datos

Tabla `device_tokens` se crea automáticamente con:
- user_id + token (unique)
- Índices optimizados
- is_valid flag
- Timestamps

## 🎯 Flujo Completo

```
User Device
    ↓
    Register FCM Token
    ↓
POST /api/notifications/fcm-token
    ↓
Backend: SaveDeviceToken
    ↓
MySQL: Insert device_tokens
    ↓
Response: 200 OK ✅

---

Stream Owner
    ↓
Inicia stream
    ↓
PUT /api/streams/{id}/start
    ↓
Backend: StartStream
    ↓
Obtiene título + owner
    ↓
GetDeviceTokensByUsersExcept(owner)
    ↓
Crea payload FCM
    ↓
SendMulticast a Firebase
    ↓
Firebase
    ↓
Notificación a Android ✅
    ↓
Android App
    ↓
Abre Stream
```

## 🚨 Troubleshooting

### "FIREBASE_CREDENTIALS_PATH not set"
→ Configura el archivo .env con la ruta correcta

### "Failed to initialize Firebase"
→ Verifica que el JSON de credenciales sea válido
→ Verifica permisos de lectura del archivo

### "Notificación no llega"
→ Verifica que el FCM token sea válido (reciente)
→ Verifica que el usuario **NO** sea el owner del stream
→ Revisa logs para errores de Firebase

### Token marcado como inválido
→ Usuario desinstaló la app o limpió datos
→ Elimina el token manualmente: DELETE /api/notifications/fcm-token

## 📱 Datos en Notificación Android

Tu app recibe:
```json
{
  "data": {
    "type": "stream_live",
    "stream_id": "550e8400-e29b-41d4-a716-446655440000",
    "stream_title": "Mi Stream Épico",
    "title": "Stream en vivo",
    "message": "Mi Stream Épico está en vivo"
  }
}
```

Usar `stream_id` en Android para navegar al stream:
```kotlin
val streamId = intent.data.getQueryParameter("stream_id")
showStreamDetail(streamId)
```

## 🔄 Ciclo de Vida del Token

```
1. Generado en Android
   ↓
2. Registrado con POST /api/notifications/fcm-token
   ↓
3. Almacenado en DB (device_tokens)
   ↓
4. Usado en multicast cuando stream inicia
   ↓
5. Last_used_at actualizado
   ↓
6. (Opcional) Eliminado con DELETE si usuario lo indica
   ↓
7. (Auto) Marcado como inválido si Firebase reporta error
```

## 🎨 Ejemplo de Uso Completo

```bash
#!/bin/bash

# 1. Obtener JWT
RESPONSE=$(curl -s -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }')
TOKEN=$(echo $RESPONSE | jq -r '.token')

# 2. Registrar token FCM
curl -X POST http://localhost:8080/api/notifications/fcm-token \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "android_fcm_token_here",
    "platform": "android",
    "device_id": "my-device-123",
    "app_version": "1.0.0"
  }'

# 3. Crear stream
STREAM=$(curl -s -X POST http://localhost:8080/api/streams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Gaming Session",
    "description": "Live gaming",
    "thumbnail": "https://example.com/thumb.jpg",
    "category": "Gaming"
  }')
STREAM_ID=$(echo $STREAM | jq -r '.id')

# 4. Iniciar stream (envía notificación)
curl -X PUT http://localhost:8080/api/streams/$STREAM_ID/start \
  -H "Authorization: Bearer $TOKEN"

# ✅ Notificación enviada a todos los usuarios excepto el owner
```

---

¡Listo! El sistema de notificaciones está completo y funcionando. 🎉
