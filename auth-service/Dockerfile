# Dockerfile multi-stage para auth-service.
# Optimizado: build estatico, alpine minimal al final.
#
# El entrypoint NO es solo el binario: corre scripts/apply.sh primero
# para crear el schema 'auth' en la BD compartida si no existe (idempotente).
# Luego arranca el binario. Si la migracion falla, el contenedor se
# reinicia y vuelve a intentar.
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/auth-service .

# Runtime: alpine + psql cliente + tini + bash.
FROM alpine:3.20
RUN apk --no-cache add ca-certificates postgresql16-client tini bash && \
    addgroup -S auth && adduser -S auth -G auth
USER auth
COPY --from=builder --chown=auth:auth /out/auth-service /usr/local/bin/auth-service
COPY --chown=auth:auth migrations/ /home/auth/migrations/
COPY --chown=auth:auth scripts/   /home/auth/scripts/
COPY --chown=auth:auth scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
WORKDIR /home/auth
ENV AUTH_SERVICE_PORT=3001
EXPOSE 3001
# 127.0.0.1 y no localhost: dentro del contenedor "localhost" resuelve a ::1
# y el server escucha solo en IPv4, asi que wget no conecta y el healthcheck
# falla para siempre. El comentario va FUERA de la continuacion de linea: un
# comentario entre el HEALTHCHECK y su CMD rompe el parseo del Dockerfile.
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:3001/auth/health || exit 1
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/entrypoint.sh"]
CMD ["/usr/local/bin/auth-service"]
