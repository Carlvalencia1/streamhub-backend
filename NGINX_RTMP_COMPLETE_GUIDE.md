# Guía Completa: NGINX RTMP con HLS Automático

## 📋 Resumen de lo que se implementó

Tu sistema de streaming ahora tiene:

1. ✅ **NGINX RTMP configurado** para generar HLS automáticamente
2. ✅ **Soporte para múltiples streams** con `stream_key` dinámico
3. ✅ **Webhooks NGINX** que validan streams con tu backend
4. ✅ **Endpoints en Go** para validar y marcar streams como activos/inactivos
5. ✅ **Eliminación total de FFmpeg manual**

---

## 🚀 PASO 1: Compilar tu Backend Go

Primero, actualiza Go al nuevo código con los endpoints de validación:

```bash
cd /path/to/StreamHub-Back

# Actualizar dependencias
go mod tidy

# Compilar
go build -o api cmd/api/main.go

# Verificar que no hay errores
./api --version  # O el comando que uses
```

**Archivos actualizados en Go:**
- `internal/streams/domain/repository.go` - Nueva interfaz `StreamRepository`
- `internal/streams/infrastructure/stream_repository_mysql.go` - Métodos `GetByStreamKey` y `Update`
- `internal/streams/interfaces/http/validation_handler.go` - NUEVO: Handler para NGINX
- `internal/streams/interfaces/http/routes.go` - Nuevas rutas sin auth
- `internal/server/router.go` - Registrar validationHandler

---

## 🚀 PASO 2: Configurar NGINX RTMP en EC2

### 2.1 SSH a tu EC2
```bash
ssh -i your-key.pem ec2-user@54.144.66.251
# O si es Ubuntu:
ssh -i your-key.pem ubuntu@54.144.66.251
```

### 2.2 Verificar si NGINX con RTMP está instalado
```bash
nginx -V 2>&1 | grep rtmp
```

Si **NO aparece `--add-module=../nginx-rtmp-module`**, necesitas compilar NGINX con el módulo RTMP.

### 2.2.1 (SI NO ESTÁ INSTALADO) Compilar NGINX con RTMP

```bash
# Descargar NGINX y RTMP module
cd /tmp
wget http://nginx.org/download/nginx-1.24.0.tar.gz
wget https://github.com/arut/nginx-rtmp-module/archive/master.zip

# Extraer
tar xzf nginx-1.24.0.tar.gz
unzip master.zip

# Compilar
cd nginx-1.24.0
./configure \
  --prefix=/etc/nginx \
  --sbin-path=/usr/sbin/nginx \
  --modules-path=/usr/lib64/nginx/modules \
  --conf-path=/etc/nginx/nginx.conf \
  --error-log-path=/var/log/nginx/error.log \
  --http-log-path=/var/log/nginx/access.log \
  --pid-path=/var/run/nginx.pid \
  --add-module=../nginx-rtmp-module-master

make
sudo make install

# Crear systemd service
sudo systemctl daemon-reload
```

### 2.3 Copiar configuración NGINX
```bash
# En tu máquina local:
scp -i your-key.pem nginx.conf ec2-user@54.144.66.251:/tmp/

# En EC2:
sudo cp /tmp/nginx.conf /etc/nginx/nginx.conf
```

### 2.4 Crear directorios
```bash
sudo mkdir -p /var/www/hls
sudo mkdir -p /var/www/html
sudo chown -R nginx:nginx /var/www/hls
sudo chown -R nginx:nginx /var/www/html
sudo chmod -R 755 /var/www/hls
```

### 2.5 Verificar configuración
```bash
sudo nginx -t
# Esperado: "nginx: the configuration file test is successful"
```

### 2.6 Iniciar NGINX
```bash
sudo systemctl start nginx
sudo systemctl enable nginx  # Auto-start en reboot
sudo systemctl status nginx
```

---

## 🔍 VERIFICACIÓN

### Verificar puertos
```bash
# En EC2:
sudo netstat -tlnp | grep nginx
# Deberías ver puerto 1935 (RTMP) y 80 (HTTP)

# Desde tu máquina local:
telnet 54.144.66.251 1935
telnet 54.144.66.251 80
```

### Ver directorios HLS
```bash
# En EC2:
ls -la /var/www/hls/
du -sh /var/www/hls/
```

### Ver logs en tiempo real
```bash
# En EC2:
sudo tail -f /var/log/nginx/error.log
sudo tail -f /var/log/nginx/access.log
```

---

## 🧪 PRUEBAS PASO A PASO

### Test 1: Verificar que tu backend Go está corriendo
```bash
curl http://3.232.197.126:8081/api/health
# Esperado: {"status":"ok"}
```

### Test 2: Crear un stream desde tu backend
```bash
curl -X POST http://3.232.197.126:8081/api/streams \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Stream",
    "description": "Testing NGINX RTMP"
  }'

# Guarda el stream_key de la respuesta
# Ejemplo de respuesta:
#{
#  "id": "12345",
#  "title": "Test Stream",
#  "stream_key": "abc-def-123",  # ← GUARDA ESTO
#  "rtmp_url": "rtmp://54.144.66.251/live/abc-def-123",
#  "playback_url": "http://54.144.66.251/live/abc-def-123/index.m3u8"
#}
```

### Test 3: Enviar stream con FFmpeg (prueba)
```bash
# Generar stream de prueba
ffmpeg -f lavfi -i testsrc=s=1280x720:d=30 \
        -f lavfi -i sine=f=440:d=30 \
        -pix_fmt yuv420p -c:v libx264 -b:v 2500k -preset veryfast \
        -c:a aac -b:a 128k -flvflags no_duration_filesize \
        -rtmp_live live \
        rtmp://54.144.66.251/live/{STREAM_KEY_DEL_PASO_2}

# El stream debería conectarse automáticamente
# NGINX validará con tu backend
```

### Test 4: Ver logs de validación en backend
```bash
# En tu backend (en la terminal donde corre):
# Deberías ver logs como:
# [StreamValidation] Received validation request
# [StreamValidation] Validating stream key: abc-def-123
# [StreamValidation] ✓ Stream key validated successfully
```

### Test 5: Verificar archivos HLS generados
```bash
# En EC2:
ls -la /var/www/hls/abc-def-123/
# Deberías ver:
# - index.m3u8
# - index-0.ts
# - index-1.ts
# - etc.
```

### Test 6: Ver playlist HLS
```bash
curl http://54.144.66.251/live/abc-def-123/index.m3u8
# Esperado:
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:2
#EXTINF:2.0,
#index-0.ts
```

### Test 7: Reproducir en VLC
1. Abre VLC
2. Media → Abrir URL de red
3. Ingresa: `http://54.144.66.251/live/{STREAM_KEY}/index.m3u8`
4. Deberías ver el stream en vivo

---

## ⚡ FLUJO COMPLETO

### OBS Streamer
```
1. En OBS Settings:
   - Server: rtmp://54.144.66.251/live
   - Stream Key: (la que genera tu backend)

2. Presionar "Start Streaming"
```

### Backend Go
```
1. Usuario hace POST /api/streams → Genera stream_key
2. NGINX envía POST /api/streams/validate-key → Backend valida
3. Backend marca stream como IsLive = true
```

### NGINX RTMP
```
1. Recibe stream RTMP
2. Valida con backend
3. Genera automáticamente HLS en /var/www/hls/{stream_key}/
4. Crea index.m3u8 y segmentos .ts
```

### Cliente (Web/Android)
```
1. Obtiene stream_key del backend
2. Reproduce con: http://54.144.66.251/live/{stream_key}/index.m3u8
3. VLC, HLS.js, ExoPlayer, AVPlayer, etc. reproducen el stream
```

### Cuando termina
```
1. OBS presiona "Stop Streaming"
2. NGINX envía POST /api/streams/stop
3. Backend marca stream como IsLive = false
4. NGINX limpia archivos HLS después de hls_cleanup
```

---

## 🔧 COMANDOS ÚTILES

### Reiniciar NGINX
```bash
sudo systemctl restart nginx
sudo nginx -s reload  # Sin interrumpir conexiones
```

### Limpiar archivos HLS viejos
```bash
sudo find /var/www/hls -name "*.ts" -mmin +30 -delete
sudo find /var/www/hls -type d -empty -delete
```

### Ver streams activos
```bash
curl http://54.144.66.251/stat
# O en navegador: http://54.144.66.251/stat
```

### Ver logs en tiempo real
```bash
sudo tail -f /var/log/nginx/error.log
sudo tail -f /var/log/nginx/access.log
```

---

## ⚠️ PROBLEMAS COMUNES

### "Unable to open RTMP URL"
```
- Verificar puerto 1935 abierto
- Verificar que NGINX está corriendo: sudo systemctl status nginx
- Verificar logs: sudo tail -f /var/log/nginx/error.log
```

### "Stream key validation failed"
```
- Verificar que backend Go está corriendo
- Verificar que la URL de validación es correcta en nginx.conf
- Ver logs de validación en backend
- Probar conectividad: curl http://3.232.197.126:8081/api/health
```

### "HLS files not generating"
```
- Verificar permisos de /var/www/hls
- Verificar que NGINX tiene módulo RTMP: nginx -V | grep rtmp
- Ver logs: sudo tail -f /var/log/nginx/error.log
- Verificar que stream está llegando: ffmpeg -i rtmp://... stats
```

### "Disco lleno"
```
- Limpiar segmentos viejos: find /var/www/hls -mmin +30 -delete
- Ajustar hls_playlist_length y hls_fragment en nginx.conf
- Agregar limpieza automática a crontab
```

---

## 📝 CHECKLIST FINAL

- [ ] Go backend compilado sin errores
- [ ] NGINX con módulo RTMP instalado
- [ ] nginx.conf copiada a EC2
- [ ] Directorios /var/www/hls creados con permisos correctos
- [ ] NGINX iniciado y escuchando puerto 1935 y 80
- [ ] Backend Go corriendo en puerto 8081
- [ ] Prueba de creación de stream en backend
- [ ] Prueba de validación (logs en backend)
- [ ] Stream generado desde FFmpeg/OBS
- [ ] Archivos HLS (.m3u8, .ts) creados en /var/www/hls
- [ ] Reproducción en VLC funciona
- [ ] Múltiples streams/usuarios simultáneos probados

---

## 🎯 SIGUIENTES PASOS

1. **Agregar autenticación RTMP** (opcional pero recomendado):
   - Usar on_publish para validar mejor el stream_key
   - Implementar token JWT en NGINX

2. **ABR (Adaptive Bitrate)** (para multi-calidad):
   - Configurar múltiples calidades de entrada
   - Usar FFmpeg transcoding en NGINX

3. **Grabación y VOD**:
   - Descomentar sección `record` en nginx.conf
   - Guardar streams para reproducción posterior

4. **Escalabilidad**:
   - Añadir WebRTC para baja latencia
   - Usar CDN para distribución global
   - Load balancing para múltiples servidores NGINX

