# Stage 1 — build Vue dashboard
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN npm i -g pnpm && pnpm install --frozen-lockfile
COPY frontend/ .
RUN pnpm build

# Stage 2 — build Go binary with embedded frontend
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend /app/dist ./internal/ui/dist
RUN go build -o toggleflow ./cmd/server

# Stage 3 — final image (~20MB, no Node, no Go toolchain)
FROM alpine:3.19
WORKDIR /
COPY --from=backend /app/toggleflow /toggleflow
RUN mkdir -p /data && \
    printf '#!/bin/sh\nulimit -n 65536 2>/dev/null || true\nexec /toggleflow "$@"\n' > /entrypoint.sh && \
    chmod +x /entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
