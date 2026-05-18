<p align="center">
  <img src="https://raw.githubusercontent.com/beedevz/hive/main/frontend/public/logo.png" alt="Hive" width="100" />
</p>

<h1 align="center">Hive</h1>

<p align="center">
  A self-hosted, open-source homelab start page.<br/>
  Config-driven · 45+ service integrations · No cloud dependencies.
</p>

<p align="center">
  <a href="https://hive.beedevz.com">Live Demo</a> ·
  <a href="https://github.com/beedevz/hive/releases">Releases</a> ·
  <a href="DEVELOPMENT.md">Developer Guide</a>
</p>

---

## Quick Start

### Docker (single container)

```bash
docker run -d \
  -p 3000:8080 \
  -v $(pwd)/config:/config \
  -e HIVE_TOKEN=your-secret-token \
  --name hive \
  --restart unless-stopped \
  beedevztech/hive:latest
```

Open `http://localhost:3000`. On first boot, a sample `config.yaml` is created automatically.

> **HIVE_TOKEN** is the password used to unlock editing in the UI. Default is `changeme` — always set a strong value before exposing Hive to a network.

### Docker Compose

```bash
# 1. Copy the environment file and fill in your tokens
cp .env.example .env

# 2. Start
docker compose -f docker-compose.prod.yml up -d
```

`docker-compose.prod.yml` pulls the pre-built image from Docker Hub. To pin a specific version:

```bash
VERSION=v1.3.1 docker compose -f docker-compose.prod.yml up -d
```

### Helm (Kubernetes)

```bash
helm repo add hive https://beedevz.github.io/hive
helm repo update

helm install hive hive/hive \
  --set hiveToken=your-secret-token \
  --set persistence.enabled=true
```

With Ingress:

```bash
helm install hive hive/hive \
  --set hiveToken=your-secret-token \
  --set persistence.enabled=true \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set "ingress.hosts[0].host=hive.example.com" \
  --set "ingress.hosts[0].paths[0].path=/" \
  --set "ingress.hosts[0].paths[0].pathType=Prefix"
```

To install or upgrade:

```bash
helm repo update
helm upgrade --install hive hive/hive --reuse-values
```

> `--install` installs if no release exists, upgrades if it does — safe to use every time.

---

## Configuration

Everything lives in `config/config.yaml`. Edit it directly on disk or use the UI (unlock with your token — click 🔒 in the top-right corner).

### Settings

```yaml
settings:
  title: "Hive"
  theme: dark          # dark | light
  columns: 2           # 1–4
  show_greeting: true
  greeting: "Good morning"
```

### Services

```yaml
services:
  - category: Infrastructure
    items:
      - name: Portainer
        url: http://portainer.local:9000
        icon: https://cdn.jsdelivr.net/gh/selfhst/icons/svg/portainer.svg
        description: Container management
        adapter: portainer
        adapter_config:
          token: "${PORTAINER_TOKEN}"
```

### Bookmarks

```yaml
bookmarks:
  - category: Dev Tools
    items:
      - name: GitHub
        url: https://github.com
        icon: https://cdn.jsdelivr.net/gh/selfhst/icons/svg/github.svg
```

### Widgets

```yaml
widgets:
  - type: clock
    enabled: true
  - type: search
    enabled: true
    config:
      engine: google    # google | duckduckgo | bing | startpage | custom
  - type: resources
    enabled: true
  - type: weather
    enabled: true
    config:
      location_name: Istanbul   # city name; leave empty to use browser geolocation
```

### Environment variables

Sensitive values (API keys, passwords) should be stored in `.env` and referenced in `config.yaml` as `${VAR_NAME}` — they are expanded server-side and never sent to the browser.

```bash
cp .env.example .env
nano .env
```

---

## Service Adapters

Adapters pull live metrics from your services and display them as stat badges on each card. Results are cached for 60 seconds.

### Monitoring & DNS

| Adapter | Auth | Env vars |
|---------|------|----------|
| `adguard` | Username + password | `ADGUARD_USER` `ADGUARD_PASS` |
| `pihole` | Token (v5) or password (v6) | `PIHOLE_TOKEN` |
| `grafana` | Service account token | `GRAFANA_TOKEN` |
| `netdata` | — | — |
| `uptime-kuma` | Username + password | `UPTIMEKUMA_USER` `UPTIMEKUMA_PASS` |

### Infrastructure

| Adapter | Auth | Env vars |
|---------|------|----------|
| `proxmox` | API token (`USER@REALM!ID=SECRET`) | `PROXMOX_TOKEN` |
| `portainer` | API key | `PORTAINER_TOKEN` |
| `traefik` | Username + password (optional) | `TRAEFIK_USER` `TRAEFIK_PASS` |
| `npm` | Username + password | `NPM_USER` `NPM_PASS` |
| `glances` | — | — |
| `truenas` | API key | `TRUENAS_APIKEY` |
| `scrutiny` | — | — |
| `synology` | Username + password | `SYNO_USER` `SYNO_PASS` |
| `unifi` | Username + password | `UNIFI_USER` `UNIFI_PASS` |
| `opnsense` | API key + secret | `OPNS_KEY` `OPNS_SECRET` |
| `frigate` | — | — |
| `watchtower` | HTTP API token | `WATCHTOWER_TOKEN` |

### Media & Downloads

| Adapter | Auth | Env vars |
|---------|------|----------|
| `jellyfin` | API key | `JELLYFIN_TOKEN` |
| `plex` | X-Plex-Token | `PLEX_TOKEN` |
| `emby` | API key | `EMBY_TOKEN` |
| `sonarr` | API key | `SONARR_APIKEY` |
| `radarr` | API key | `RADARR_APIKEY` |
| `lidarr` | API key | `LIDARR_APIKEY` |
| `readarr` | API key | `READARR_APIKEY` |
| `prowlarr` | API key | `PROWLARR_APIKEY` |
| `bazarr` | API key | `BAZARR_APIKEY` |
| `overseerr` | API key | `OVERSEERR_APIKEY` |
| `jellyseerr` | API key | `JELLYSEERR_APIKEY` |
| `qbittorrent` | Username + password | `QB_USER` `QB_PASS` |
| `transmission` | Username + password | `TR_USER` `TR_PASS` |
| `deluge` | Password | `DELUGE_PASS` |
| `sabnzbd` | API key | `SABNZBD_APIKEY` |
| `nzbget` | Username + password | `NZBGET_USER` `NZBGET_PASS` |

### Services & Tools

| Adapter | Auth | Env vars |
|---------|------|----------|
| `nextcloud` | Username + password | `NC_USER` `NC_PASS` |
| `immich` | API key | `IMMICH_APIKEY` |
| `vaultwarden` | — | — |
| `homeassistant` | Long-lived token | `HASS_TOKEN` |
| `gitea` / `forgejo` | API token | `GITEA_TOKEN` / `FORGEJO_TOKEN` |
| `gitlab` | Personal access token | `GITLAB_TOKEN` |
| `paperless` | API token | `PAPERLESS_TOKEN` |
| `firefly` | Personal token | `FIREFLY_TOKEN` |
| `speedtest` | — | — |
| `wdmycloud` | Username + password | `WD_USER` `WD_PASS` |

### Network

| Adapter | Auth | Env vars |
|---------|------|----------|
| `cloudflare` | API token + account ID | `CF_TOKEN` `CF_ACCOUNT_ID` |
| `tailscale` | API key + tailnet | `TS_APIKEY` `TS_TAILNET` |

---

## Icons

Use any emoji or image URL in the `icon` field. The UI includes a built-in icon picker (click 🔍 next to the icon field) that lets you search and select from the [selfh.st icon library](https://selfh.st/icons/). You can also paste any URL directly:

```
https://cdn.jsdelivr.net/gh/selfhst/icons/svg/{service-name}.svg
```

---

## License

MIT — see [LICENSE](LICENSE) for details.
