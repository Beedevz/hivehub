# Security

## Reporting

If you find a vulnerability in Hive, please open a private security advisory on
GitHub (`Security` tab → `Report a vulnerability`) rather than filing a public
issue. We respond within a few days.

## Scope

Hive is a homelab dashboard intended to run behind a trusted network or a
TLS-terminating reverse proxy (nginx, Traefik, Caddy, an ingress controller…).
Threat model assumes:

- The container runs on an internal network or behind an authenticating proxy.
- `HIVE_TOKEN` is set to a strong, operator-chosen value before exposure.
- The operator controls `/config` contents and the `secrets.yaml` file.

Releases ship as a non-root container (`nginx` user, uid 101) listening on
port 8080.

## Accepted findings

The CI security scans flag a handful of patterns that we have reviewed and
accepted as false positives. They are listed here so future reviewers can
audit the rationale instead of re-discovering it.

### `unsafe-deserialization-interface` — yaml/json Unmarshal into `interface{}`

Sites: `config-api/main.go` (`readConfig`, `writeConfig`, `configRawSaveHandler`).

Go's `encoding/json` and `gopkg.in/yaml.v3` decoders do not execute code and do
not instantiate arbitrary types during `Unmarshal`. The deserialization-gadget
class of vulnerabilities that this rule was written for (Python `pickle`,
Java `Serializable`, Ruby `Marshal`, .NET `BinaryFormatter`) has no analogue
here. The only practical risk is resource exhaustion via very large payloads,
which is bounded by `http.MaxBytesReader` (10 MiB) on the two handlers that
accept user input.

### `possible-nginx-h2c-smuggling` — `proxy_set_header Upgrade $http_upgrade`

Sites: `nginx/nginx-multi.conf` (Portainer and Uptime Kuma WebSocket proxies).

The `Upgrade` header forwarding is required for WebSocket support on those
upstreams. HTTP/2 cleartext smuggling is mitigated by the
`map $http_upgrade $connection_upgrade` block at the top of the file, which
maps `h2c` and `h2` to an empty Connection value so they cannot reach the
upstream.

### `no-direct-write-to-responsewriter` — direct `w.Write` calls

All response writes go through the `writeBytes` helper in `config-api/main.go`.
The bytes written are either JSON marshaled from typed structs, the contents
of an operator-controlled file (`config.yaml`, `secrets.yaml`, a custom logo),
or a small literal status payload. Every response sets an explicit
`Content-Type` and the corsMiddleware adds `X-Content-Type-Options: nosniff`
to prevent MIME sniffing.

### `net.use-tls` — `http.ListenAndServe` in `config-api`

`config-api` binds plain HTTP on the internal docker network only. TLS is
terminated upstream by the reverse proxy. The listener uses
`http.Server` with `ReadHeaderTimeout` set to bound slow-header attacks.

## Outbound TLS

All outbound HTTP clients in `config-api` set `TLSClientConfig.MinVersion =
tls.VersionTLS12`. The probe client allows self-signed certificates by
default (`PROBE_INSECURE_TLS=true`) because home-lab services frequently use
them; set `PROBE_INSECURE_TLS=false` to enforce verification.
