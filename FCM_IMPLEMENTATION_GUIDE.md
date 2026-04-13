# Firebase Cloud Messaging (FCM) - Implementación Completa

## 📋 Descripción General

Sistema de notificaciones push que permite notificar usuarios cuando un stream va en vivo. Utiliza Firebase Cloud Messaging (FCM) para enviar notificaciones a dispositivos Android.

## 🏗️ Arquitectura

```
Domain Layer
├── DeviceToken (Entity)
└── NotificationRepository (Interface)
    └── PushProvider (Interface)

Application Layer
├── RegisterFcmToken (UseCase)
├── RemoveFcmToken (UseCase)
└── NotifyStreamLive (UseCase)

Infrastructure Layer
├── DeviceTokenRepository (MySQL)
└── FirebasePushProvider (FCM)

Interfaces Layer
└── HTTP Handler + Routes

Integration
└── Inyectado en StartStream para notificar
```

## 🚀 Configuración

### 1. Variables de Entorno

Agregar al archivo `.env`:

```env
# Firebase Credentials Path (ruta absoluta a service account JSON)
FIREBASE_CREDENTIALS_PATH=/path/to/serviceAccountKey.json
```

### 2. Obtener Credenciales de Firebase

1. Ir a [Firebase Console](https://console.firebase.google.com)
2. Seleccionar proyecto Android
3. Ir a Settings → Service Accounts
4. Generar nueva clave privada (descarga JSON)
5. Guardar en ruta segura del servidor
6. Apuntar FIREBASE_CREDENTIALS_PATH a esa ruta

### 3. Instalar Dependencias Go

```bash
go get firebase.google.com/go/v4
go get google.golang.org/api
```

## 📁 Archivos Creados

### Domain Layer
- `internal/notifications/domain/device_token.go` - Entidad DeviceToken
- `internal/notifications/domain/repository.go` - Interfaces NotificationRepository y PushProvider

### Application Layer
- `internal/notifications/application/register_fcm_token.go` - Caso de uso registrar token
- `internal/notifications/application/remove_fcm_token.go` - Caso de uso eliminar token
- `internal/notifications/application/notify_stream_live.go` - Caso de uso notificar stream en vivo
- `internal/notifications/application/errors.go` - Errores del módulo

### Infrastructure Layer
- `internal/notifications/infrastructure/device_token_repository_mysql.go` - Persistencia en MySQL
- `internal/notifications/infrastructure/firebase_provider.go` - Proveedor de FCM

### Interfaces Layer
- `internal/notifications/interfaces/http/handler.go` - Handlers HTTP
- `internal/notifications/interfaces/http/routes.go` - Rutas HTTP

### Database
- `internal/platform/database/migrations/005_create_device_tokens_table.sql` - Tabla device_tokens

### Modified Files
- `internal/platform/config/config.go` - Agregar FirebaseCredentialsPath
- `internal/platform/database/migrations/migrate.go` - Incluir migración 005
- `internal/server/server.go` - Pasar config a RegisterRoutes
- `internal/server/router.go` - Inicializar Firebase y registrar rutas
- `internal/streams/application/start_stream.go` - Enviar notificación al iniciar stream
- `internal/streams/application/notification_injector.go` - Inyección de dependencias
- `pkg/response/response.go` - Agregar SuccessResponse y ErrorResponse

### Tests
- `internal/notifications/application/notify_stream_live_test.go` - Tests unitarios

## 🔌 API Endpoints

### 1. Registrar Token FCM

**Endpoint:** `POST /api/notifications/fcm-token`  
**Auth:** Bearer JWT Token  
**Content-Type:** `application/json`

**Request Body:**
```json
{
  "token": "cdJoxvz0QYW_JjAZ1kLtN4:APA91bH3...",
  "platform": "android",
  "device_id": "device_abc123",
  "app_version": "1.2.3"
}
```

**Response 200:**
```json
{
  "success": true,
  "message": "Token registrado exitosamente"
}
```

**Regla:** Upsert (sin duplicar). Si el usuario ya registró este token, se actualiza.

---

### 2. Eliminar Token FCM

**Endpoint:** `DELETE /api/notifications/fcm-token`  
**Auth:** Bearer JWT Token  
**Content-Type:** `application/json`

**Request Body:**
```json
{
  "token": "cdJoxvz0QYW_JjAZ1kLtN4:APA91bH3..."
}
```

**Response 200:**
```json
{
  "success": true,
  "message": "Token eliminado exitosamente"
}
```

---

### 3. Evento: Stream En Vivo (Automático)

Cuando se inicia un stream (`PUT /api/streams/:id/start`):

1. Backend inicia el stream
2. Obtiene título y owner del stream
3. Obtiene todos los tokens FCM (excepto owner)
4. Envía notificación multicast a Firebase
5. Firebase entrega a dispositivos Android

**Payload FCM enviado:**
```json
{
  "notification": {
    "title": "Stream en vivo",
    "body": "My Stream Title está transmitiendo"
  },
  "data": {
    "type": "stream_live",
    "stream_id": "550e8400-e29b-41d4-a716-446655440000",
    "stream_title": "My Stream Title",
    "title": "Stream en vivo",
    "message": "My Stream Title está en vivo"
  }
}
```

## 📝 Ejemplos cURL

### 1. Obtener JWT Token

```bash
# Registrarse
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# Login
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Response incluye token:
# {
#   "token": "eyJhbGciOiJIUzI1NiIs..."
# }
```

### 2. Registrar Token FCM

```bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."
FCM_TOKEN="cdJoxvz0QYW_JjAZ1kLtN4:APA91bH3vKdJGDl1..."

curl -X POST http://localhost:8080/api/notifications/fcm-token \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "'$FCM_TOKEN'",
    "platform": "android",
    "device_id": "device_abc123",
    "app_version": "1.2.3"
  }'

# Response:
# {
#   "success": true,
#   "message": "Token registrado exitosamente"
# }
```

### 3. Registrar Múltiples Tokens (desde diferentes dispositivos)

```bash
# Device 1
curl -X POST http://localhost:8080/api/notifications/fcm-token \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "token_device_1",
    "platform": "android",
    "device_id": "device_1",
    "app_version": "1.2.3"
  }'

# Device 2
curl -X POST http://localhost:8080/api/notifications/fcm-token \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "token_device_2",
    "platform": "android",
    "device_id": "device_2",
    "app_version": "1.2.3"
  }'
```

### 4. Eliminar Token FCM

```bash
curl -X DELETE http://localhost:8080/api/notifications/fcm-token \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "'$FCM_TOKEN'"
  }'

# Response:
# {
#   "success": true,
#   "message": "Token eliminado exitosamente"
# }
```

### 5. Iniciar Stream (Trigger Notificación)

```bash
STREAM_ID="550e8400-e29b-41d4-a716-446655440000"

curl -X PUT http://localhost:8080/api/streams/$STREAM_ID/start \
  -H "Authorization: Bearer $TOKEN"

# Response:
# {
#   "success": true,
#   "data": { stream_data }
# }

# En background:
# - Stream marcado como is_live=true
# - Notificación enviada a todos excepto owner
# - Logs: "stream live notification sent to X devices"
```

## 📱 Integración Android

En tu app Android, recibir el payload:

```kotlin
class NotificationHandler : FirebaseMessagingService() {
    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        val data = remoteMessage.data
        
        val type = data["type"]  // "stream_live"
        val streamId = data["stream_id"]
        val streamTitle = data["stream_title"]
        
        if (type == "stream_live") {
            // Abrir stream automáticamente
            openStreamScreen(streamId, streamTitle)
        }
    }
}
```

## 🗄️ Estructura de Base de Datos

Tabla `device_tokens`:
```sql
id VARCHAR(36) PRIMARY KEY
user_id VARCHAR(36) FOREIGN KEY
token TEXT (token FCM)
platform VARCHAR(50) (android/ios)
device_id VARCHAR(255) (ID único del dispositivo)
app_version VARCHAR(50) (versión de app)
is_valid BOOLEAN (false si Firebase reporta inválido)
last_used_at TIMESTAMP (último envío exitoso)
created_at TIMESTAMP
updated_at TIMESTAMP

UNIQUE: (user_id, token)
INDEXES: user_id, is_valid, created_at
```

## 🔍 Logging

El sistema registra eventos clave:

```log
[INFO] Firebase Messaging client initialized successfully
[DEBUG] RegisterFcmToken usecase started for user: user_123
[INFO] FCM token registered successfully for user: user_123
[DEBUG] NotifyStreamLive usecase started for stream: stream_456
[INFO] stream live notification sent to 42 devices for stream: stream_456
[DEBUG] marked token as invalid
[WARN] FIREBASE_CREDENTIALS_PATH not set, push notifications will be disabled
```

**IMPORTANTE:** Los tokens NO se imprimen completos en logs por seguridad.

## ⚠️ Manejo de Errores

### Tokens Inválidos

Cuando Firebase reporta un token como inválido/unregistered:

1. Se marca en BD como `is_valid = false`
2. No se intenta enviar en futuros multicast
3. Se pueden eliminar directamente con `DELETE /api/notifications/fcm-token`

### Sin Credenciales de Firebase

Si `FIREBASE_CREDENTIALS_PATH` no está configurado:
- Sistema logged pero notificaciones deshabilitadas
- No bloquea inicio del servidor
- Resgistro de tokens sigue funcionando (storing in DB)

### Fallos de Red/Firebase

- Retry automático (handled by Firebase SDK)
- Logs de errores por token
- No bloquea el stream start

## 📊 Estadísticas y Monitoreo

Para obtener estadísticas:

```sql
-- Tokens registrados
SELECT COUNT(*) FROM device_tokens WHERE is_valid = true;

-- Tokens por user
SELECT user_id, COUNT(*) FROM device_tokens
WHERE is_valid = true
GROUP BY user_id;

-- Tokens inválidos (para cleanup)
SELECT COUNT(*) FROM device_tokens WHERE is_valid = false;

-- Últimos tokens registrados
SELECT user_id, platform, created_at 
FROM device_tokens 
ORDER BY created_at DESC 
LIMIT 20;
```

## 🧪 Testing

Ejecutar tests:

```bash
go test ./internal/notifications/application -v
```

Tests incluyen:
- ✅ Envío exitoso a múltiples usuarios
- ✅ Sin tokens registrados
- ✅ Validación de input
- ✅ Errores del repository
- ✅ Input genérico (map)
- ✅ Sin provider de Firebase

## 🔐 Seguridad

✅ Tokens FCM NO se imprimen en logs  
✅ JWT required para registrar/eliminar tokens  
✅ Cada usuario solo puede eliminar sus propios tokens  
✅ Credenciales Firebase protegidas en ruta del servidor  
✅ Conexión HTTPS (recomendado en producción)  

## 📚 Archivos de Referencia

- [Firebase Admin SDK Go Docs](https://firebase.google.com/docs/reference/admin/go)
- [FCM Concepts](https://firebase.google.com/docs/cloud-messaging)
- [Android Integration Guide](https://firebase.google.com/docs/cloud-messaging/android/client)

---

**Estado:** ✅ Implementación Completa  
**Versión:** 1.0  
**Última actualización:** 2025
