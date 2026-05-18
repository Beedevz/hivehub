FROM golang:1.26-alpine AS builder
ARG VERSION=dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o config-api .

FROM alpine:3.23
RUN apk update && apk upgrade --no-cache && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/config-api .
EXPOSE 3001
CMD ["./config-api"]
