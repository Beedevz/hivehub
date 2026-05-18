# Deployment · Kubernetes

Hive ships with a Helm chart for Kubernetes deployment.

## Prerequisites

- Kubernetes 1.22+
- Helm 3+
- A `StorageClass` that supports `ReadWriteOnce` PVCs
- (Optional) An ingress controller (nginx-ingress, Traefik, etc.)

---

## Add Helm Repository

```bash
helm repo add hive https://beedevz.github.io/hive
helm repo update
```

---

## Install

### Minimal (no ingress)

```bash
helm install hive hive/hive \
  --set hiveToken=your-strong-secret
```

Access via port-forward:
```bash
kubectl port-forward svc/hive 3000:8080
```

### With Ingress

```bash
helm install hive hive/hive \
  --set hiveToken=your-strong-secret \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set "ingress.hosts[0].host=hive.example.com" \
  --set "ingress.hosts[0].paths[0].path=/" \
  --set "ingress.hosts[0].paths[0].pathType=Prefix"
```

### With TLS (cert-manager)

```bash
helm install hive hive/hive \
  --set hiveToken=your-strong-secret \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set "ingress.hosts[0].host=hive.example.com" \
  --set "ingress.hosts[0].paths[0].path=/" \
  --set "ingress.hosts[0].paths[0].pathType=Prefix" \
  --set "ingress.tls[0].secretName=hive-tls" \
  --set "ingress.tls[0].hosts[0]=hive.example.com" \
  --set "ingress.annotations.cert-manager\\.io/cluster-issuer=letsencrypt-prod"
```

---

## values.yaml Reference

```yaml
replicaCount: 1

image:
  repository: beedevztech/hive
  pullPolicy: IfNotPresent
  tag: ""   # Defaults to chart appVersion

hiveToken: "changeme"   # !! Change this before exposing publicly !!

persistence:
  enabled: true
  size: 100Mi
  accessMode: ReadWriteOnce
  # storageClass: ""   # Use cluster default

ingress:
  enabled: false
  className: ""
  annotations: {}
  hosts:
    - host: hive.local
      paths:
        - path: /
          pathType: Prefix
  tls: []

resources:
  limits:
    cpu: 200m
    memory: 64Mi
  requests:
    cpu: 50m
    memory: 32Mi

nodeSelector: {}
tolerations: []
affinity: {}
```

---

## Upgrade

```bash
helm upgrade hive hive/hive \
  --set hiveToken=your-strong-secret \
  --set image.tag=v1.7.0
```

> **Important:** Always pass `--set image.tag=<version>` when upgrading to ensure the correct image is used. If you omit it, Helm uses the chart's default `appVersion`.

---

## Persistent Config

The chart uses a **PersistentVolumeClaim** for `/config`. On first boot, an init container copies the default config from the ConfigMap into the PVC. On subsequent restarts, the PVC data is used as-is.

This means:
- Config survives pod restarts and upgrades
- Edit config via the UI or API — it is saved to the PVC
- To reset to defaults, delete and recreate the PVC

---

## Environment Variables for Adapters

Pass adapter credentials as extra environment variables:

```bash
helm upgrade hive hive/hive \
  --set hiveToken=your-secret \
  --set "extraEnv[0].name=PORTAINER_TOKEN" \
  --set "extraEnv[0].value=ptr_xyz123" \
  --set "extraEnv[1].name=PIHOLE_TOKEN" \
  --set "extraEnv[1].value=abc456"
```

Or via a values file (`values-prod.yaml`):
```yaml
hiveToken: your-secret

extraEnv:
  - name: PORTAINER_TOKEN
    value: ptr_xyz123
  - name: PIHOLE_TOKEN
    value: abc456
  - name: JELLYFIN_TOKEN
    valueFrom:
      secretKeyRef:
        name: hive-adapter-secrets
        key: jellyfin-token
```

---

## Uninstall

```bash
helm uninstall hive
# PVC is NOT deleted automatically — delete manually if needed:
kubectl delete pvc hive
```
