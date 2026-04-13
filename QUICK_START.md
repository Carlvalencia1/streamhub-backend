# 🚀 QUICK START: NGINX RTMP + Backend Integration

> **Resumen ejecutivo** - Todo lo que necesitas para comenzar en 15 minutos

---

## ✅ LO QUE YA SE HIZO

Tu código Go ya tiene:
- ✅ Nuevo handler `StreamValidationHandler` para NGINX webhooks
- ✅ Nuevos métodos `GetByStreamKey()` y `Update()` en repository
- ✅ Nuevas rutas sin autenticación para validación
- ✅ Configuración NGINX completa con HLS automático

**Archivos nuevos/modificados:**
- `internal/streams/interfaces/http/validation_handler.go` ← NUEVO
- `internal/streams/domain/repository.go` ← MODIFICADO
- `internal/streams/infrastructure/stream_repository_mysql.go` ← MODIFICADO
- `internal/streams/interfaces/http/routes.go` ← MODIFICADO
- `internal/server/router.go` ← MODIFICADO
- `nginx.conf` ← NUEVO

---

## 🎯 PRÓXIMOS PASOS (en orden)

### PASO 1: Compilar Backend (5 min)
```bash
cd ~/StreamHub-Back
go mod tidy
go build -o api cmd/api/main.go
```

### PASO 2: Configurar NGINX en EC2 (5 min)
```bash
# SSH a EC2
ssh -i your-key.pem ec2-user@54.144.66.251

# Crear directorios
sudo mkdir -p /var/www/hls
sudo chown -R nginx:nginx /var/www/hls
sudo chmod -R 755 /var/www/hls

# Copiar nginx.conf desde tu máquina
scp -i your-key.pem nginx.conf ec2-user@IP:/tmp/
sudo cp /tmp/nginx.conf /etc/nginx/nginx.conf

# Verificar
sudo nginx -t

# Restart
sudo systemctl restart nginx
```

### PASO 3: Verificar (5 min)
```bash
# En EC2
sudo netstat -tlnp | grep nginx
# Debe ver: 1935 (RTMP) y 80 (HTTP)

# Desde tu máquina
curl http://54.144.66.251/health
# Debe retornar: OK
```

---

## 🧪 PROBAR FLUJO COMPLETO

### 1. Crear stream
```bash
curl -X POST http://3.232.197.126:8081/api/streams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Stream"}'

# Guardar stream_key de la respuesta
STREAM_KEY="abc-123-def"
```

### 2. Enviar RTMP (FFmpeg)
```bash
ffmpeg -f lavfi -i testsrc=s=1280x720:d=30 \
        -f lavfi -i sine=f=440 \
        -pix_fmt yuv420p -c:v libx264 -b:v 2500k \
        -c:a aac -b:a 128k \
        rtmp://54.144.66.251/live/$STREAM_KEY
```

### 3. Ver HLS generado
```bash
curl http://54.144.66.251/live/$STREAM_KEY/index.m3u8
# Debe ver contenido M3U8
```

### 4. Reproducir en VLC
```
VLC > Media > Abrir URL
http://54.144.66.251/live/$STREAM_KEY/index.m3u8
```

---

## 📋 URLS IMPORTANTES

| Componente | URL |
|-----------|-----|
| RTMP Publish | `rtmp://54.144.66.251/live/{stream_key}` |
| HLS Playback | `http://54.144.66.251/live/{stream_key}/index.m3u8` |
| Backend API | `http://3.232.197.126:8081` |
| Backend Health | `http://3.232.197.126:8081/api/health` |
| NGINX Stat | `http://54.144.66.251/stat` |

---

## 🔑 ENDPOINTS CRÍTICOS

### Crear Stream (Con Auth)
```
POST /api/streams
Authorization: Bearer {JWT_TOKEN}

Response:
{
  "stream_key": "uuid-123",
  "rtmp_url": "rtmp://54.144.66.251/live/uuid-123",
  "playback_url": "http://54.144.66.251/live/uuid-123/index.m3u8"
}
```

### Validar Stream (NGINX - Sin Auth)
```
POST /api/streams/validate-key?app=live&name=uuid-123

Response:
{
  "valid": true
}
```

### Parar Stream (NGINX - Sin Auth)
```
POST /api/streams/stop?app=live&name=uuid-123

Response:
{
  "message": "Stream stopped successfully"
}
```

---

## 🎬 CONFIGURACIÓN OBS

1. Abrir OBS
2. Ir a Settings → Stream
3. Service: Custom
4. Server: `rtmp://54.144.66.251/live`
5. Stream Key: `{stream_key_de_tu_backend}`
6. Click Start Streaming

---

## 📊 BD REQUIRED

Ya debes tener estas columnas en tabla `streams`:
- `stream_key` VARCHAR(36) UNIQUE ← IMPORTANTE
- `is_live` BOOLEAN DEFAULT false
- `playback_url` TEXT

Si faltan, ejecutar:
```sql
ALTER TABLE streams 
ADD COLUMN stream_key VARCHAR(36) UNIQUE,
ADD COLUMN is_live BOOLEAN DEFAULT false,
ADD COLUMN playback_url TEXT;
```

---

## ⚠️ COMMON ISSUES

### "Stream key validation failed"
```
→ Verificar que backend está corriendo
→ Verificar que BD tiene el stream_key
→ Ver logs: tail -f /var/log/nginx/error.log
```

### "RTMP connection refused"
```
→ Verificar puerto 1935 abierto en Security Group EC2
→ Ver si NGINX está corriendo: sudo systemctl status nginx
→ Verificar que NGINX tiene módulo RTMP: nginx -V | grep rtmp
```

### "HLS files not generating"
```
→ Verificar permisos: ls -la /var/www/hls/
→ Ver si stream llega a NGINX: ffmpeg -i rtmp://... -t 1 -f null -
→ Ver logs: tail -f /var/log/nginx/error.log
```

---

## 📚 DOCUMENTACIÓN COMPLETA

Lee estos archivos para más detalles:
1. **NGINX_RTMP_COMPLETE_GUIDE.md** - Guía paso a paso completa
2. **NGINX_BACKEND_INTEGRATION.md** - Detalle técnico de integración
3. **CODE_EXAMPLES_NGINX_INTEGRATION.md** - Ejemplos de código
4. **DOCKER_SETUP_OPTIONAL.md** - Si quieres usar Docker (opcional)

---

## ✅ CHECKLIST FINAL

- [ ] Backend Go compilado
- [ ] nginx.conf copiada a EC2
- [ ] NGINX reiniciado
- [ ] Puertos 1935 y 80 abiertos en Security Group
- [ ] Backend corriendo en puerto 8081
- [ ] Stream creado en backend (tiene stream_key)
- [ ] FFmpeg conecta a RTMP sin errores
- [ ] Archivos .m3u8 y .ts en /var/www/hls
- [ ] URL HLS accesible desde navegador
- [ ] VLC reproduce el stream
- [ ] Logs no muestran errores

---

## 🎯 PRÓXIMOS PASOS AVANZADOS

- [ ] Implementar seguridad RTMP con token
- [ ] Agregar recordación de streams (VOD)
- [ ] Multi-bitrate (ABR)
- [ ] WebRTC para baja latencia
- [ ] CDN para distribución global

---

## 🆘 SOPORTE

Si algo no funciona:
1. Ver logs: `tail -f /var/log/nginx/error.log`
2. Probar conectividad: `nc -zv 54.144.66.251 1935`
3. Ver archivos generados: `ls -la /var/www/hls/`
4. Probar con ffmpeg: `ffmpeg -i rtmp://... stats`

