FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json .
RUN npm install
COPY . .
RUN npm run build

FROM nginxinc/nginx-unprivileged:alpine
USER root
RUN apk update && apk upgrade --no-cache
COPY --from=builder /app/dist /usr/share/nginx/html
# nginx.conf is mounted via docker-compose volume
USER nginx
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
