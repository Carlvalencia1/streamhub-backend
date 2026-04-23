## CHECKLIST: Notificaciones no llegan a followers

### 1. DATABASE
Ejecuta el script `DEBUG_NOTIFICATIONS.sql` reemplazando los UUIDs:
```sql
SET @STREAMER_ID = 'uuid-del-streamer';
SET @FOLLOWER_ID = 'uuid-del-follower';
```

Verifica:
- ¿Existe la relación en tabla `followers`? (Query #1)
- ¿El follower tiene tokens guardados? (Query #3)
- ¿Están marcados como válidos (is_valid = true)? (Query #4)
- ¿La query de notificación trae tokens? (Query #5)

### 2. ANDROID CLIENT
**Problema probable:** El token del follower no se está sincronizando

En Android, verifica los logs:
```
adb logcat | grep "StreamMessagingService"
```

Deberías ver:
```
D/StreamMessagingService: Token FCM refresheado: <token>
D/StreamMessagingService: Token encolado para sincronizacion: <token>
```

Si NO ves estos logs → El token nuevo no se está registrando

### 3. SERVER LOGS
En los logs del backend cuando se inicia stream:
```
INFO: stream live notification sent to X devices for stream
```

Si dice "sent to 0 devices" → Los tokens del follower no se están recuperando de BD

### 4. FIREBASE LOGS  
Revisa si Firebase rechazó el token. Con la nueva actualización,
deberías ver en logs:
```
WARN: token failed to send: <token>, error: <error>
```

### 5. PRUEBA RÁPIDA EN BD
```sql
-- Ver el último token del follower
SELECT * FROM device_tokens 
WHERE user_id = 'uuid-follower' 
ORDER BY created_at DESC LIMIT 1;

-- Contar cuántos tokens tiene el follower
SELECT COUNT(*) FROM device_tokens 
WHERE user_id = 'uuid-follower';

-- Contar cuántos followers tiene el streamer
SELECT COUNT(*) FROM followers 
WHERE streamer_id = 'uuid-streamer';
```

---

### 💡 Causas Posibles (en orden de probabilidad):

1. **Follower NO registró token** 
   - Android Worker no corre
   - Endpoint 401/403 (no autenticado)
   - Red lenta al sincronizar

2. **Token está marcado como inválido**
   - Firebase rechazó el token
   - Token duplicado/malformado

3. **Relación followers NO existe**
   - Follow endpoint falla silenciosamente
   - Datos desincronizados

4. **Query SQL falla**
   - Índices no funcionan correctamente
