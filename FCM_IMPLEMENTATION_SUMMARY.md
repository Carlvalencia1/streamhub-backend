# 📦 Firebase Cloud Messaging - Resumen de Implementación

## ✅ Estado: IMPLEMENTACIÓN COMPLETA

Sistema de notificaciones push con Firebase Cloud Messaging (FCM) integrado en StreamHub Backend.

---

## 📋 Archivos Nuevos

### Domain Layer (2 archivos)
```
✓ internal/notifications/domain/device_token.go
  - Entity: DeviceToken
  - Constructor: NewDeviceToken()

✓ internal/notifications/domain/repository.go
  - Interface: NotificationRepository (7 métodos)
  - Interface: PushProvider (2 métodos)
  - Type: PushPayload
```

### Application Layer (5 archivos)
```
✓ internal/notifications/application/register_fcm_token.go
  - UseCase: RegisterFcmToken
  - Input: RegisterFcmTokenInput
  - Funcionalidad: Registrar/actualizar token FCM (upsert)

✓ internal/notifications/application/remove_fcm_token.go
  - UseCase: RemoveFcmToken
  - Input: RemoveFcmTokenInput
  - Funcionalidad: Eliminar token específico

✓ internal/notifications/application/notify_stream_live.go
  - UseCase: NotifyStreamLive
  - Input: NotifyStreamLiveInput (genérico)
  - Funcionalidad: Enviar notificación multicast al iniciar stream

✓ internal/notifications/application/errors.go
  - Errores comunes del módulo

✓ internal/notifications/application/notify_stream_live_test.go
  - Tests unitarios (6 test cases)
  - Mocks: MockNotificationRepository, MockPushProvider
```

### Infrastructure Layer (2 archivos)
```
✓ internal/notifications/infrastructure/device_token_repository_mysql.go
  - Implementación: NotificationRepository
  - CRUD completo con upsert
  - Consultas optimizadas

✓ internal/notifications/infrastructure/firebase_provider.go
  - Implementación: PushProvider
  - Inicialización Firebase Admin SDK
  - SendMulticast con manejo de errores
  - SendMulticastBatch para límite de tokens
```

### Interfaces/HTTP Layer (2 archivos)
```
✓ internal/notifications/interfaces/http/handler.go
  - Handler: RegisterFCMToken
  - Handler: RemoveFCMToken
  - DTOs: RegisterFCMTokenRequest, RemoveFCMTokenRequest

✓ internal/notifications/interfaces/http/routes.go
  - Rutas: POST /api/notifications/fcm-token
  - Rutas: DELETE /api/notifications/fcm-token
  - Middleware: AuthMiddleware en todas las rutas
```

### Database (1 archivo)
```
✓ internal/platform/database/migrations/005_create_device_tokens_table.sql
  - Tabla: device_tokens
  - Campos: id, user_id, token, platform, device_id, app_version, is_valid, last_used_at, timestamps
  - Índices: user_id, is_valid, created_at
  - Constraint: UNIQUE(user_id, token)
```

### Streams Application (1 archivo NEW)
```
✓ internal/streams/application/notification_injector.go
  - Variable global: streamLiveNotifier
  - Funciones: SetStreamLiveNotifier(), GetStreamLiveNotifier()
  - Propósito: Evitar importaciones cíclicas
```

### Documentation (2 archivos)
```
✓ FCM_IMPLEMENTATION_GUIDE.md
  - Guía completa de 300+ líneas
  - API Reference
  - Ejemplos cURL
  - Integración Android
  - Troubleshooting

✓ FCM_QUICK_START.md
  - Quick start en 5 minutos
  - Flujos visuales ASCII
  - Ejemplos prácticos
```

---

## 📝 Archivos Modificados

### Configuration (1 archivo)
```
✓ internal/platform/config/config.go
  + FirebaseCredentialsPath string
```

### Database Migrations (1 archivo)
```
✓ internal/platform/database/migrations/migrate.go
  + Migración 005_create_device_tokens_table incluida en migrations[]
```

### Server Setup (2 archivos)
```
✓ internal/server/server.go
  ~ RegisterRoutes(router, s.cfg, s.db)  // Antes: RegisterRoutes(router, s.db)

✓ internal/server/router.go
  + Importaciones: notifications packages
  + Inicialización Firebase Provider
  + Creación de repositories y usecases
  + Inyección en streams module
  ~ RegisterRoutes signature: (r, cfg, db)
```

### Streams Application (1 archivo)
```
✓ internal/streams/application/start_stream.go
  + Lógica de notificación al iniciar stream
  + Obtención de datos del stream
  + Llamada a StreamLiveNotifier en background
  + Manejo graceful de errores
```

### Response Package (1 archivo)
```
✓ pkg/response/response.go
  + SuccessResponse struct
  + ErrorResponse struct
```

---

## 🔧 Configuración Requerida

### 1. .env
```env
FIREBASE_CREDENTIALS_PATH=/secure/path/serviceAccountKey.json
```

### 2. Firebase Console
- Descargar service account JSON
- Guardar en ruta segura del servidor
- Verificar permisos de FCM

### 3. Dependencias Go
```bash
go get firebase.google.com/go/v4
go get google.golang.org/api
```

---

## 🚀 Cómo Usar

### 1. Registrar Token FCM
```bash
POST /api/notifications/fcm-token
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "token": "cdJoxvz0QYW_JjAZ1kLtN4:APA91bH3...",
  "platform": "android",
  "device_id": "device_abc123",
  "app_version": "1.2.3"
}

Response: 200 OK
{
  "success": true,
  "message": "Token registrado exitosamente"
}
```

### 2. Eliminar Token FCM
```bash
DELETE /api/notifications/fcm-token
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "token": "cdJoxvz0QYW_JjAZ1kLtN4:APA91bH3..."
}

Response: 200 OK
{
  "success": true,
  "message": "Token eliminado exitosamente"
}
```

### 3. Iniciar Stream (Auto-notifica)
```bash
PUT /api/streams/{stream_id}/start
Authorization: Bearer <JWT>

Response: 200 OK
{
  "success": true,
  "data": { stream_data }
}

# Automáticamente:
# 1. Stream marcado como is_live=true
# 2. Obtiene todos los tokens (excepto owner)
# 3. Envía notificación multicast a Firebase
# 4. Firebase entrega a dispositivos Android
```

---

## 📱 Payload FCM Recibido en Android

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

---

## 🧪 Tests

```bash
# Ejecutar tests
go test ./internal/notifications/application -v

# Resultados incluyen:
✓ NotifyStreamLive_Execute_Success
✓ NotifyStreamLive_Execute_NoTokens
✓ NotifyStreamLive_Execute_InvalidInput
✓ NotifyStreamLive_Execute_RepositoryError
✓ NotifyStreamLive_Execute_GenericMapInput
✓ NotifyStreamLive_Execute_NoProvider
```

---

## 📊 Base de Datos

Tabla `device_tokens` schema:
```sql
CREATE TABLE device_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token TEXT NOT NULL,
    platform VARCHAR(50) DEFAULT 'android',
    device_id VARCHAR(255),
    app_version VARCHAR(50),
    is_valid BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_token (user_id, token(100)),
    INDEX idx_user_id (user_id),
    INDEX idx_is_valid (is_valid),
    INDEX idx_created_at (created_at)
);
```

---

## ✨ Características

- ✅ **Upsert automático**: No hay duplicados de tokens
- ✅ **Multicast**: Envía a múltiples dispositivos en una llamada
- ✅ **Batch processing**: Maneja límite de 500 tokens Firebase
- ✅ **Manejo de errores**: Marca tokens inválidos automáticamente
- ✅ **Background async**: Notificaciones no bloquean stream start
- ✅ **JWT Auth**: Solo usuarios autenticados pueden registrar tokens
- ✅ **Logging seguro**: No imprime tokens/JWT completos
- ✅ **Graceful degradation**: Si Firebase no está configurado, sigue funcionando

---

## 🔐 Seguridad

- 🔒 Rutas `/api/notifications/*` protegidas con JWT
- 🔒 Tokens FCM no se imprimen en logs
- 🔒 Cada usuario solo puede acess/delete sus propios tokens
- 🔒 Credenciales Firebase en ruta segura (no en repo)
- 🔒 Context timeout para llamadas Firebase

---

## 📚 Documentación

Para documentación completa ver:
- **FCM_IMPLEMENTATION_GUIDE.md** - Guía exhaustiva
- **FCM_QUICK_START.md** - Guía rápida de 5 minutos

---

## ✅ Criterios de Aceptación

- ✅ **Compila**: Go build sin errores
- ✅ **Registra token correctamente**: POST /api/notifications/fcm-token 200 OK
- ✅ **Al iniciar stream envía push**: PUT /api/streams/:id/start + notificación
- ✅ **Android recibe stream_id**: Data contiene "stream_id": "uuid"
- ✅ **Tokens inválidos se limpian**: Automáticamente marcados como is_valid=false
- ✅ **Tests pasan**: 6/6 test cases passing

---

## 🎯 Próximos Pasos (Opcionales)

- [ ] Agregar scheduled cleanup de tokens inválidos (cron job)
- [ ] Implementar NotifyStreamOffline (cuando stream termina)
- [ ] Agregar preferencias de notificación por usuario
- [ ] Soporte para iOS (Apple Push Notification Service)
- [ ] Dashboard de estadísticas de entrega
- [ ] Retry policy personalizado

---

**Implementación completada:** ✅  
**Versión:** 1.0  
**Fecha:** Abril 2025  
**Stack:** Go + Gin + Firebase Admin SDK + MySQL
