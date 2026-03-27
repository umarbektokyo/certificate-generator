# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o certificate-server .

# Stage 3: Production image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /app/server/certificate-server .
COPY --from=frontend /app/web/build ./static/
EXPOSE 8181
HEALTHCHECK --interval=10s --timeout=3s CMD wget -qO- http://localhost:8181/api/health || exit 1
CMD ["./certificate-server"]
