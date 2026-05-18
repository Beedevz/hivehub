# Secrets Management

Hive has a built-in secrets store for managing API keys and credentials. Secrets are stored server-side in `/config/secrets.yaml` and are **never** sent to the browser.

## How It Works

1. Store a secret via the API or UI (`key=MY_TOKEN`, `value=abc123`)
2. Reference it in `config.yaml` with `${MY_TOKEN}`
3. The backend expands it at request time before forwarding to the adapter

## Option 1 — Environment Variables (simple)

The easiest approach for Docker Compose:

**.env**
```env
HIVE_TOKEN=your-hive-token
PORTAINER_TOKEN=ptr_xyz123
PIHOLE_TOKEN=abc456
```

**config.yaml**
```yaml
adapter_config:
  token: "${PORTAINER_TOKEN}"
```

---

## Option 2 — Secrets API (recommended for production)

Use the Secrets API to store credentials without putting them in `.env`.

### Add a secret

```bash
curl -X PUT http://hive.local/api/secrets \
  -H "X-Hive-Token: $HIVE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key": "PORTAINER_TOKEN", "value": "ptr_xyz123"}'
```

### List secret keys (values are never exposed)

```bash
curl http://hive.local/api/secrets \
  -H "X-Hive-Token: $HIVE_TOKEN"
```

```json
{ "keys": ["PORTAINER_TOKEN", "PIHOLE_TOKEN"] }
```

### Delete a secret

```bash
curl -X DELETE "http://hive.local/api/secrets?key=PORTAINER_TOKEN" \
  -H "X-Hive-Token: $HIVE_TOKEN"
```

### Import from YAML

```bash
curl -X POST http://hive.local/api/secrets/import \
  -H "X-Hive-Token: $HIVE_TOKEN" \
  -H "Content-Type: application/yaml" \
  --data-binary @secrets.yaml
```

`secrets.yaml` format:
```yaml
PORTAINER_TOKEN: ptr_xyz123
PIHOLE_TOKEN: abc456
JELLYFIN_TOKEN: xyz789
```

### Backup secrets

```bash
curl http://hive.local/api/secrets/backup \
  -H "X-Hive-Token: $HIVE_TOKEN" \
  -o secrets-backup.yaml
```

---

## Storage

Secrets are stored in `/config/secrets.yaml` inside the container. Make sure `/config` is a persistent volume:

```yaml
volumes:
  - ./config:/config
```

## Kubernetes

In Kubernetes, use the Helm chart's built-in secret for `HIVE_TOKEN`. For adapter credentials, you can:

1. Pass additional environment variables via `extraEnv` in Helm values
2. Use the Secrets API after deployment to store credentials in the PVC

See [[Deployment · Kubernetes]] for details.
