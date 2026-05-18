# Quick Start

## Option 1 — Docker Compose (recommended)

### 1. Create working directory

```bash
mkdir hive && cd hive
```

### 2. Create `.env`

```bash
cp .env.example .env
# Edit .env and set HIVE_TOKEN to a strong secret
```

At minimum set:
```env
HIVE_TOKEN=your-strong-secret-here
```

### 3. Create `config/config.yaml`

```bash
mkdir config
```

Minimal config:
```yaml
settings:
  title: "Hive"
  theme: dark

services:
  - category: My Services
    items:
      - name: Portainer
        url: http://portainer.local:9000
        icon: https://cdn.jsdelivr.net/gh/selfhst/icons/svg/portainer.svg
```

See [[Configuration]] for full reference.

### 4. Create `docker-compose.yml`

```yaml
services:
  hive:
    image: beedevztech/hive:latest
    ports:
      - "3000:8080"
    volumes:
      - ./config:/config
    env_file:
      - .env
    restart: unless-stopped
```

### 5. Start

```bash
docker compose up -d
```

Open `http://localhost:3000`

---

## Option 2 — Kubernetes (Helm)

See [[Deployment · Kubernetes]] for the full Helm guide.

---

## Option 3 — Build from Source

```bash
git clone https://github.com/Beedevz/hive.git
cd hive
cp .env.example .env   # edit as needed
docker compose up --build
```

---

## Unlock Editing

Click the **lock icon** in the top right and enter your `HIVE_TOKEN`. While unlocked you can:

- Add / remove / reorder services and bookmarks
- Edit service details (name, URL, icon, adapter config)
- Toggle widgets
- Import / export config

Changes are saved to `/config/config.yaml` immediately. A timestamped backup is created automatically on every save.

---

## Next Steps

- Configure adapters → [[Adapters]]
- Store API keys securely → [[Secrets Management]]
- Set up Kubernetes ingress → [[Deployment · Kubernetes]]
