# Dockerfile multi-stage para auth-service.
# Optimizado para tamano de imagen: compila estatico, copia solo el binario.
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/auth-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates && \
    addgroup -S auth && adduser -S auth -G auth
USER auth
COPY --from=builder /out/auth-service /usr/local/bin/auth-service
WORKDIR /home/auth
ENV AUTH_SERVICE_PORT=3001
EXPOSE 3001
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:3001/auth/health || exit 1
ENTRYPOINT ["/usr/local/bin/auth-service"]
