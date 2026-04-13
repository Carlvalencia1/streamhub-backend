# Scripts y comandos para gestionar NGINX RTMP

## 📋 INSTALACIÓN Y SETUP

### 1. Crear directorios necesarios
```bash
# SSH a tu EC2 (54.144.66.251)
ssh -i your-key.pem ec2-user@54.144.66.251

# Crear directorios
sudo mkdir -p /var/www/hls
sudo mkdir -p /var/www/html
sudo mkdir -p /var/www/recordings
sudo chown -R nginx:nginx /var/www/hls
sudo chown -R nginx:nginx /var/www/recordings
sudo chmod -R 755 /var/www/hls
sudo chmod -R 755 /var/www/recordings

# Crear logs directory
sudo mkdir -p /var/log/nginx
sudo chown -R nginx:nginx /var/log/nginx
```

### 2. Copiar configuración NGINX
```bash
# En tu servidor EC2:
sudo cp /ruta/a/nginx.conf /etc/nginx/nginx.conf
# O si usas Homebrew en Mac:
sudo cp /ruta/a/nginx.conf /usr/local/etc/nginx/nginx.conf
```

### 3. Verificar sintaxis
```bash
sudo nginx -t
# Salida esperada: "nginx: configuration file test is successful"
```

---

## 🚀 COMANDOS DE CONTROL

### Reiniciar NGINX
```bash
# Opción 1: Reinicio limpio
sudo systemctl restart nginx
sudo service nginx restart

# Opción 2: Reload (preserva conexiones activas)
sudo systemctl reload nginx
sudo nginx -s reload

# Opción 3: Stop
sudo systemctl stop nginx

# Opción 4: Start
sudo systemctl start nginx
```

### Verificar estado
```bash
# Estado del servicio
sudo systemctl status nginx
sudo service nginx status

# Ver procesos activos
ps aux | grep nginx

# Ver puertos escuchando
sudo netstat -tlnp | grep nginx
sudo ss -tlnp | grep nginx
```

---

## 📊 MONITOREO Y LOGS

### Ver logs en tiempo real
```bash
# Error logs
sudo tail -f /var/log/nginx/error.log

# Access logs
sudo tail -f /var/log/nginx/access.log

# RTMP logs (si están habilitados)
sudo tail -f /var/log/nginx/rtmp.log
```

### Ver estadísticas de RTMP en vivo
```bash
# Accede a: http://54.144.66.251/stat
# Podrás ver todos los streams activos
```

### Limpiar archivos HLS antiguos
```bash
# Limpiar segmentos más antiguos de 1 hora
find /var/www/hls -name "*.ts" -mmin +60 -delete

# Limpiar playlist vacíos
find /var/www/hls -name "*.m3u8" -empty -delete
```

---

## 🧪 PRUEBAS

### 1. Verificar que NGINX está escuchando
```bash
# Puerto RTMP (1935)
nc -zv 54.144.66.251 1935

# Puerto HTTP (80)
nc -zv 54.144.66.251 80

# Con curl
curl -v http://54.144.66.251/health
# Esperado: 200 OK
```

### 2. Probar con FFmpeg (verificación, no producción)
```bash
# Generar stream de prueba
ffmpeg -f lavfi -i testsrc=s=1280x720:d=300 \
        -f lavfi -i sine=f=440:d=300 \
        -pix_fmt yuv420p -c:v libx264 -b:v 2500k -preset veryfast \
        -c:a aac -b:a 128k -flvflags no_duration_filesize \
        -rtmp_live live \
        rtmp://54.144.66.251/live/test-stream-key

# En otra terminal, verificar HLS
curl http://54.144.66.251/live/test-stream-key/index.m3u8
```

### 3. Ver archivos HLS creados
```bash
# SSH a EC2
ssh -i your-key.pem ec2-user@54.144.66.251

# Ver archivos generados
ls -la /var/www/hls/

# Ver estructura de directories anidados
find /var/www/hls -type f | head -20

# Contar streams activos
ls -1d /var/www/hls/*/ | wc -l
```

### 4. Verificar URL de playback
```bash
# En tu navegador o VLC:
http://54.144.66.251/live/{stream_key}/index.m3u8

# Ver contenido de playlist
curl http://54.144.66.251/live/{stream_key}/index.m3u8
```

### 5. Probar validación con backend
```bash
# Ver si NGINX intenta validar (logs)
sudo tail -f /var/log/nginx/error.log | grep "upstream timed out"

# Ver requests al backend
sudo tail -f /var/log/nginx/access.log | grep "/api/streams"
```

---

## 🔍 TROUBLESHOOTING

### NGINX no inicia
```bash
# Verificar sintaxis
sudo nginx -t

# Ver logs detallados
sudo tail -100 /var/log/nginx/error.log

# Verificar permisos de directorios
ls -la /var/www/hls
ls -la /var/log/nginx
```

### HLS no se genera
```bash
# Verificar que RTMP module está compilado
sudo nginx -V 2>&1 | grep rtmp

# Ver si directorios tienen permisos
sudo touch /var/www/hls/test.txt
sudo rm /var/www/hls/test.txt
```

### Validación con backend falla
```bash
# Verificar conectividad
curl -v http://3.232.197.126:8081/api/streams/validate-key \
     -X POST \
     -H "Content-Type: application/json" \
     -d '{"app":"live","name":"test-key"}'

# Ver firewall rules
sudo ufw status
sudo iptables -L -n | grep 8081
```

### Streams ocupan mucho disco
```bash
# Ver tamaño de directorio HLS
du -sh /var/www/hls

# Limpiar automáticamente (agregar a crontab)
# Crear script /usr/local/bin/cleanup-hls.sh
#!/bin/bash
find /var/www/hls -name "*.ts" -mmin +30 -delete
find /var/www/hls -name "*.m3u8" -size 0 -delete
```

---

## ⚙️ CONFIGURACIÓN EN CRONTAB (Automático)

```bash
# Editar crontab
sudo crontab -e

# Agregar línea para limpiar cada 30 minutos
*/30 * * * * find /var/www/hls -name "*.ts" -mmin +30 -delete

# Agregar línea para verificar NGINX cada 5 minutos
*/5 * * * * systemctl is-active --quiet nginx || systemctl restart nginx
```

---

## 📱 CONFIGURACIÓN EN OBS

**RTMP Configuration:**
- Server: `rtmp://54.144.66.251/live`
- Stream Key: `{stream_key}` (el que genera tu backend)

**Verificar en cliente:**
- URL de reproducción: `http://54.144.66.251/live/{stream_key}/index.m3u8`
- Usar VLC o cualquier reproductor HLS

---

## 🎯 VERIFICACIÓN FINAL

✅ NGINX escucha en puerto 1935 (RTMP)
✅ NGINX escucha en puerto 80 (HTTP)
✅ Directorios /var/www/hls tienen permisos correctos
✅ Archivos .m3u8 y .ts se generan automáticamente
✅ Backend recibe validaciones en /api/streams/validate-key
✅ CORS habilitado para playback
✅ Cache headers configurados correctamente

