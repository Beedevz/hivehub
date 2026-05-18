# Deployment · Docker

## Docker Compose (Production)

Uses the pre-built image from Docker Hub.

**docker-compose.yml**
```yaml
services:
  hive:
    image: beedevztech/hive:latest
    container_name: hive
    ports:
      - "3000:8080"
    volumes:
      - ./config:/config
    env_file:
      - .env
    restart: unless-stopped
```

**Pin to a specific version (recommended):**
```yaml
image: beedevztech/hive:v1.6.1
```

### Start

```bash
docker compose up -d
```

### Update

```bash
docker compose pull
docker compose up -d
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `HIVE_TOKEN` | `changeme` | Auth token for protected endpoints |
| `PORT` | `3001` | Internal Go API port (don't change unless needed) |
| Adapter vars | — | `${VAR}` references in config.yaml |

**Minimal .env:**
```env
HIVE_TOKEN=your-strong-secret-here
```

---

## Volume Layout

```
./config/
├── config.yaml          # Main config (required)
├── secrets.yaml         # Secrets store (auto-created)
├── logo-dark.png        # Custom dark logo (optional)
├── logo-light.png       # Custom light logo (optional)
└── config.backup.*.yaml # Automatic backups
```

The `/config` volume is the only persistent data. Back it up to keep your configuration safe.

---

## Docker Compose (Build from Source)

For development or customization:

```yaml
services:
  hive:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3000:8080"
    volumes:
      - ./config:/config
    env_file:
      - .env
    restart: unless-stopped
```

```bash
docker compose up -d --build
```

---

## Reverse Proxy (Traefik)

```yaml
services:
  hive:
    image: beedevztech/hive:latest
    volumes:
      - ./config:/config
    env_file:
      - .env
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.hive.rule=Host(`hive.example.com`)"
      - "traefik.http.routers.hive.entrypoints=websecure"
      - "traefik.http.routers.hive.tls.certresolver=letsencrypt"
      - "traefik.http.services.hive.loadbalancer.server.port=8080"
    restart: unless-stopped
```

## Reverse Proxy (nginx)

```nginx
server {
    listen 80;
    server_name hive.example.com;

    location / {
        proxy_pass http://hive:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## Healthcheck

The container exposes a healthcheck endpoint:

```
GET http://localhost/api/health
```

Docker Compose healthcheck:
```yaml
healthcheck:
  test: ["CMD", "wget", "-qO-", "http://localhost/api/health"]
  interval: 30s
  timeout: 5s
  retries: 3
```
