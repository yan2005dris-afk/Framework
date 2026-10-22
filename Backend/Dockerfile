# Dockerfile multi-stage para el Backend Go/Fiber.
# Optimizado: build estatico sin CGO, alpine minimal al final.
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/multicatalogo-backend .

# Runtime: alpine + psql (las migraciones se aplican desde el entrypoint).
FROM alpine:3.20
RUN apk --no-cache add ca-certificates postgresql-client && \
    addgroup -S backend && adduser -S backend -G backend
USER backend
COPY --from=builder /out/multicatalogo-backend /usr/local/bin/multicatalogo-backend
# Las migraciones SQL viajan en la imagen: sin esto el entrypoint no
# encontraria el directorio y el Backend arrancaria contra un esquema viejo.
COPY --chown=backend:backend database/migrations/ /home/backend/database/migrations/
COPY --chown=backend:backend scripts/apply_migrations.sh /usr/local/bin/apply_migrations.sh
RUN chmod +x /usr/local/bin/apply_migrations.sh
WORKDIR /home/backend
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:3000/api/health || exit 1
ENTRYPOINT ["/usr/local/bin/apply_migrations.sh"]
