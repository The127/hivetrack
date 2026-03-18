# ─── Stage 1: Build frontend ───────────────────────────────────────────────
FROM node:22-alpine AS frontend
WORKDIR /workspace/hivetrack-ui
COPY hivetrack-ui/package.json hivetrack-ui/package-lock.json ./
RUN npm ci
COPY hivetrack-ui/ ./
# outDir in vite.config.js points to ../hivetrack/web/dist
RUN npm run build

# ─── Stage 2: Build backend (embeds frontend) ──────────────────────────────
FROM golang:1.25-alpine AS backend
WORKDIR /workspace/hivetrack
COPY hivetrack/go.mod hivetrack/go.sum ./
RUN go mod download
COPY hivetrack/ ./
COPY --from=frontend /workspace/hivetrack/web/dist/ ./web/dist/
RUN CGO_ENABLED=0 GOOS=linux go build -o /hivetrack ./cmd/hivetrack

# ─── Stage 3: Runtime ──────────────────────────────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /hivetrack /hivetrack
EXPOSE 8086
ENTRYPOINT ["/hivetrack"]
