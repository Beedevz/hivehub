# Hive — Homelab Start Page

**Hive** is a self-hosted, config-driven homelab start page with live service stats, a built-in API, and a clean UI.

## Quick Links

| Topic | Page |
|---|---|
| First install | [[Quick Start]] |
| Config format | [[Configuration]] |
| All adapters | [[Adapters]] |
| API endpoints | [[API Reference]] |
| Docker setup | [[Deployment·Docker]] |
| Kubernetes setup | [[Deployment·Kubernetes]] |
| Secrets & tokens | [[Secrets Management]] |
| Contributing | [[Contributing]] |
| Release process | [[Release Process]] |

## What Hive Does

- **Service dashboard** — cards for every service in your homelab, with live stats pulled from 45+ adapters
- **Bookmarks** — quick-link grid for frequently visited pages
- **Widgets** — clock, search bar, system resources (CPU/RAM/disk), weather
- **In-UI editing** — unlock with `HIVE_TOKEN` and edit everything live; changes are saved to disk instantly
- **Secrets management** — store API keys and credentials server-side; they are never sent to the browser
- **Swagger UI** — full interactive API docs at `/api/swagger/`
- **Drag-and-drop** — reorder services and bookmarks

## Architecture

```
Browser → Nginx (port 8080)
               ├── /          → React SPA (Vite build)
               └── /api/      → Go API (port 3001)
                                    └── /config, /secrets,
                                        /probe, /adapters, ...
```

Both the nginx process and the Go binary run inside a **single Docker container** managed by supervisord.

## Supported Adapters (45+)

Monitoring, infrastructure, media, download clients, storage, networking — see [[Adapters]] for the full list.

## Links

- Docker Hub: [beedevztech/hive](https://hub.docker.com/r/beedevztech/hive)
- GitHub: [Beedevz/hive](https://github.com/Beedevz/hive)
- Swagger UI: `http://your-host/api/swagger/`
