# ── Stage 1: Build React ──────────────────────────────────────────
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json .
RUN npm install
COPY frontend/ .
RUN npm run build

# ── Stage 2: Build Go API ─────────────────────────────────────────
FROM golang:1.26-alpine AS api-builder
ARG VERSION=dev
WORKDIR /app
COPY config-api/go.mod config-api/go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY config-api/ .
RUN swag init --generalInfo main.go --output docs
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o config-api .

# ── Stage 3: Final image (nginx + supervisord) ────────────────────
FROM nginx:alpine
RUN apk update && apk upgrade --no-cache && apk add --no-cache supervisor

# React build
COPY --from=frontend-builder /app/frontend/dist /usr/share/nginx/html

# Go binary
COPY --from=api-builder /app/config-api /usr/local/bin/config-api

# Nginx config
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf

# Supervisord config
COPY supervisord.conf /etc/supervisord.conf

# Default config template — kept outside the /config volume so the mount
# does not hide it; bootstrapConfig() copies it to /config/config.yaml on first boot.
COPY config/config.example.yaml /etc/hive/config.example.yaml

EXPOSE 80
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
