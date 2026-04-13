# Docker Compose: NGINX RTMP + Backend Go (Opcional)

> Esta es una **alternativa opcional** para desarrollar localmente o desplegar en contenedores.
> Si ya tienes EC2 con NGINX instalado manualmente, puedes ignorar esto.

---

## 📦 Opción: Usar Docker Compose

### docker-compose.yml

```yaml
version: '3.8'

services:
  # Backend Go
  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend
    container_name: streamhub-backend
    ports:
      - "8081:8081"
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=streamhub
      - DB_PASSWORD=streamhub_pass
      - DB_NAME=streamhub
      - PORT=8081
    depends_on:
      - mysql
    volumes:
      - ./logs:/app/logs
    networks:
      - streamhub-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/api/health"]
      interval: 10s
      timeout: 5s
      retries: 5

  # MySQL Database
  mysql:
    image: mysql:8.0
    container_name: streamhub-mysql
    environment:
      - MYSQL_ROOT_PASSWORD=root_pass
      - MYSQL_DATABASE=streamhub
      - MYSQL_USER=streamhub
      - MYSQL_PASSWORD=streamhub_pass
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./internal/platform/database/migrations:/docker-entrypoint-initdb.d
    networks:
      - streamhub-network
    healthcheck:
      test: ["CMD", "mysqladmin" ,"ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  # NGINX RTMP
  nginx:
    build:
      context: .
      dockerfile: Dockerfile.nginx
    container_name: streamhub-nginx
    ports:
      - "1935:1935"
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - hls_data:/var/www/hls
      - recordings:/var/www/recordings
    depends_on:
      - backend
    networks:
      - streamhub-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  mysql_data:
  hls_data:
  recordings:

networks:
  streamhub-network:
    driver: bridge
```

---

## 🔨 Dockerfile para Backend

#### Dockerfile.backend

```dockerfile
# Builder stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Compilar
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates curl

# Copiar binario desde builder
COPY --from=builder /app/api .

# Create logs directory
RUN mkdir -p /app/logs

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/api/health || exit 1

# Ejecutar
CMD ["./api"]
```

---

## 🔨 Dockerfile para NGINX con RTMP

#### Dockerfile.nginx

```dockerfile
FROM alpine:latest

RUN apk add --no-cache \
    build-base \
    pcre-dev \
    zlib-dev \
    openssl-dev \
    git \
    curl \
    tzdata

WORKDIR /tmp

# Descargar NGINX
RUN wget http://nginx.org/download/nginx-1.24.0.tar.gz && \
    tar xzf nginx-1.24.0.tar.gz

# Descargar nginx-rtmp-module
RUN git clone https://github.com/arut/nginx-rtmp-module.git

# Compilar NGINX con RTMP
RUN cd nginx-1.24.0 && \
    ./configure \
      --prefix=/etc/nginx \
      --sbin-path=/usr/sbin/nginx \
      --conf-path=/etc/nginx/nginx.conf \
      --error-log-path=/var/log/nginx/error.log \
      --http-log-path=/var/log/nginx/access.log \
      --pid-path=/var/run/nginx.pid \
      --add-module=../nginx-rtmp-module && \
    make && \
    make install

# Crear directorios
RUN mkdir -p /var/www/hls /var/www/html /var/log/nginx

# Limpiar
RUN rm -rf /tmp/*

EXPOSE 1935 80

STOPSIGNAL SIGTERM

CMD ["/usr/sbin/nginx", "-g", "daemon off;"]
```

---

## 🚀 Cómo usar Docker Compose

### 1. Preparar archivos
```bash
# Copiar nginx.conf a raíz del proyecto
cp nginx.conf .

# Crear Dockerfiles
# (copiar contenido arriba a los archivos)
```

### 2. Start Services
```bash
docker-compose up -d

# Ver logs
docker-compose logs -f

# Ver estado
docker-compose ps
```

### 3. Verificar
```bash
# Backend health
curl http://localhost:8081/api/health

# NGINX health
curl http://localhost/health

# Ver directorios HLS
docker exec streamhub-nginx ls -la /var/www/hls/
```

### 4. Stop Services
```bash
docker-compose down

# Con limpieza de datos (⚠️ borra BD)
docker-compose down -v
```

---

## 🧪 Testing con Docker

### Crear stream
```bash
curl -X POST http://localhost:8081/api/streams \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test"}'
```

### Enviar RTMP
```bash
ffmpeg -f lavfi -i testsrc=s=1280x720:d=30 \
        -f lavfi -i sine=f=440:d=30 \
        -pix_fmt yuv420p -c:v libx264 -b:v 2500k -preset veryfast \
        -c:a aac -b:a 128k \
        rtmp://localhost/live/{stream_key}
```

### Reproducir
```bash
# En VLC:
# Media → Abrir URL de red
# http://localhost/live/{stream_key}/index.m3u8
```

---

## 📊 Monitoreo Docker

### Ver logs de cada servicio
```bash
# Backend
docker-compose logs backend -f

# NGINX
docker-compose logs nginx -f

# MySQL
docker-compose logs mysql -f
```

### Entrar a contenedor
```bash
# Backend
docker exec -it streamhub-backend sh

# NGINX
docker exec -it streamhub-nginx sh

# MySQL
docker exec -it streamhub-mysql mysql -u streamhub -p
```

### Ver stats
```bash
docker stats
```

---

## 🔧 Variables de entorno

### .env (crear en raíz)

```env
# Backend
BACKEND_PORT=8081
BACKEND_CONTAINER=streamhub-backend

# Database
DB_HOST=mysql
DB_PORT=3306
DB_USER=streamhub
DB_PASSWORD=streamhub_pass
DB_NAME=streamhub
MYSQL_ROOT_PASSWORD=root_pass

# NGINX
NGINX_CONTAINER=streamhub-nginx
NGINX_RTMP_PORT=1935
NGINX_HTTP_PORT=80

# Streaming
BACKEND_IP=http://backend:8081
NGINX_IP=nginx
```

### Actualizar docker-compose.yml para usar .env
```yaml
environment:
  - DB_HOST=${DB_HOST}
  - DB_USER=${DB_USER}
  - DB_PASSWORD=${DB_PASSWORD}
  - PORT=${BACKEND_PORT}
```

---

## 📈 Production Deployment (EC2)

### Opción 1: Docker en EC2
```bash
# En EC2:
sudo yum install docker docker-compose

# Copiar proyecto
scp -r StreamHub-Back -i key.pem ec2-user@IP:/home/ec2-user/

# Build y run
cd StreamHub-Back
docker-compose up -d
```

### Opción 2: NGINX manual + Backend Go
```bash
# Método actual (tradicional)
# NGINX manual en EC2
# Backend Go en EC2 o RDS
```

---

## 🔒 Security (Docker)

### docker-compose.override.yml (desarrollo con secrets)

```yaml
version: '3.8'

services:
  backend:
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - API_KEY=${API_KEY}

  mysql:
    environment:
      - MYSQL_PASSWORD=${DB_PASSWORD}
```

### Usar con secrets
```bash
# Crear .env.local (no versionar)
echo "DB_PASSWORD=super_secret_pass" > .env.local

# Usar
docker-compose --env-file .env.local up
```

---

## 🛑 Troubleshooting Docker

### Backend no conecta a MySQL
```bash
docker-compose logs backend
# Ver error de conexión a mysql

# Verificar:
docker exec streamhub-backend ping mysql
```

### NGINX no genera HLS
```bash
docker exec streamhub-nginx ls -la /var/www/hls/
docker-compose logs nginx
```

### Permisos de directorios
```bash
# En NGINX
docker exec streamhub-nginx chmod -R 755 /var/www/hls
docker exec streamhub-nginx chown -R nobody:nobody /var/www/hls
```

---

## 📦 Multi-Stage Build Optimization

El Dockerfile.backend ya usa multi-stage para reducir tamaño:
- Builder stage: ~500MB (compilación)
- Final stage: ~15MB (solo binario)

Resultado final: **~400MB para backend** + dependencias

---

## 🎯 Checklist Docker

- [ ] docker-compose.yml creado
- [ ] Dockerfile.backend creado
- [ ] Dockerfile.nginx creado
- [ ] nginx.conf en raíz
- [ ] go.mod en raíz
- [ ] docker-compose up sin errores
- [ ] Backend health ok
- [ ] NGINX health ok
- [ ] MySQL corriendo
- [ ] Puertos disponibles (no en uso)
- [ ] HLS directory montado como volumen
- [ ] Logs accesibles

