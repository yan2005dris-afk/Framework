# Dockerfile del gateway (Caddy 2.x).
# Compila estatico, copia solo el binario + Caddyfile, expone :8080.
#
# Para personalizar modulos (e.g. caddy-jwt) usariamos caddy:2-builder
# con 'xcaddy build' + 'COPY --from=builder /build/caddy /usr/bin/caddy'.
# En este PR no agregamos modulos: el forward_auth cubre la validacion
# central via el auth-service.

FROM caddy:2

# Copiamos el Caddyfile en la ruta default que Caddy lee si existe.
# Tambien soportamos /etc/caddy/Caddyfile para volumenes custom.
COPY Caddyfile /etc/caddy/Caddyfile

# Por defecto Caddy escucha en :2019 (HTTP) y :443 (HTTPS). Cambiamos a 8080
# explicito via env. Si QUEREMOS que el Dockerfile arranque Caddy solo,
# podriamos sobreescribir CMD, pero por defecto heredado es OK con
# 'caddy run --config /etc/caddy/Caddyfile' desde el entrypoint base.
ENV CADDY_PORT=8080

EXPOSE 8080

# Healthcheck: hit / y verificar 200 (Caddy siempre devuelve metadata
# incluso cuando file_server 404; conviene un endpoint /health dedicado
# en el gateway que llegue al upstream health, pero eso es GA7).
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/ || exit 1

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile"]
