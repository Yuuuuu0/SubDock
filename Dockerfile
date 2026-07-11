# Stage 1: Build frontend
FROM node:20-alpine AS frontend
ENV PNPM_HOME=/pnpm
ENV PATH=$PNPM_HOME:$PATH
WORKDIR /app/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# Stage 2: Build backend
FROM golang:1.24-alpine AS backend
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./internal/router/dist
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o subdock .

# Stage 3: Runtime
FROM alpine:3.21
WORKDIR /app
RUN addgroup -S -g 1000 subdock \
    && adduser -S -D -H -u 1000 -G subdock subdock \
    && apk add --no-cache ca-certificates tzdata \
    && mkdir -p /data \
    && chown -R subdock:subdock /app /data \
    && chmod 0700 /data
COPY --from=backend --chown=subdock:subdock /app/subdock .
ENV DATA_DIR=/data PORT=8080
EXPOSE 8080
VOLUME /data
USER subdock:subdock
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -T 2 -O /dev/null "http://127.0.0.1:${PORT}/api/config" || exit 1
CMD ["./subdock"]
