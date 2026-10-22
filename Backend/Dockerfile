# Dockerfile multi-stage para el Backend Go/Fiber.
# Optimizado: build estatico sin CGO, alpine minimal al final.
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/multicatalogo-backend .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates && \
    addgroup -S backend && adduser -S backend -G backend
USER backend
COPY --from=builder /out/multicatalogo-backend /usr/local/bin/multicatalogo-backend
WORKDIR /home/backend
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:3000/api/health || exit 1
ENTRYPOINT ["/usr/local/bin/multicatalogo-backend"]
