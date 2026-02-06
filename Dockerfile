# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

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
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /app/subdock .
ENV DATA_DIR=/data PORT=8080
EXPOSE 8080
VOLUME /data
CMD ["./subdock"]
