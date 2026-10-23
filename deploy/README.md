# Guía de Operación y Despliegue en Producción (VM Proxmox)

Este directorio contiene la configuración de producción para el stack **Framework / MultiCatálogo** basado en imágenes publicadas en **GitHub Container Registry (GHCR)** y actualización automática mediante **Watchtower**.

---

## 1. Estructura de Archivos

```
deploy/
├── docker-compose.prod.yml   # Definición de servicios productivos desde GHCR + Watchtower
├── .env.prod.example        # Plantilla de variables de entorno para producción
└── README.md                # Este documento de instrucciones
```

---

## 2. Puesta en Marcha en la VM Proxmox (VM 104)

### Paso 1: Conexión SSH
```bash
ssh uadmin@10.10.10.90
```

### Paso 2: Clonar o actualizar el repositorio en `/srv/framework`
```bash
sudo mkdir -p /srv/framework
sudo chown -R uadmin:uadmin /srv/framework
git clone https://github.com/yan2005dris-afk/Framework.git /srv/framework
cd /srv/framework
```

### Paso 3: Autenticación en GHCR (para que Docker y Watchtower descarguen las imágenes)
Generar un **GitHub Personal Access Token (classic)** con permiso `read:packages` y ejecutar:
```bash
echo "<TU_GITHUB_PAT>" | docker login ghcr.io -u yan2005dris-afk --password-stdin
```
*Esto creará `/home/uadmin/.docker/config.json`, que Watchtower montará para autenticarse automáticamente con GHCR.*

### Paso 4: Configurar variables de entorno
```bash
cp deploy/.env.prod.example deploy/.env
nano deploy/.env
```
Asegurarse de definir contraseñas seguras para:
- `POSTGRES_PASSWORD`
- `AUTH_JWT_SECRET`
- `SAAS_DB_PASSWORD`

### Paso 5: Descargar imágenes y levantar el stack
```bash
docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env pull
docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env up -d
```

---

## 3. Funcionamiento de Watchtower (CD Automático)

El contenedor `multicatalogo-watchtower` monitorea periódicamente (cada 300 segundos por defecto) el registro `ghcr.io/yan2005dris-afk/*`.

Cuando mergeás código a `main`:
1. El workflow de GitHub Actions [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml) compila y publica las imágenes con tag `latest`.
2. Watchtower detecta el nuevo digest en GHCR.
3. Descarga la nueva imagen (`pull`), recrea el contenedor correspondiente y limpia la imagen antigua sin interrumpir el resto de los servicios.

### Ver logs de Watchtower:
```bash
docker logs -f multicatalogo-watchtower
```

---

## 4. Comandos de Mantenimiento

- **Ver estado de los contenedores:**
  ```bash
  docker compose -f deploy/docker-compose.prod.yml ps
  ```
- **Ver logs de un servicio:**
  ```bash
  docker compose -f deploy/docker-compose.prod.yml logs -f gateway
  docker compose -f deploy/docker-compose.prod.yml logs -f backend
  ```
- **Reiniciar el stack:**
  ```bash
  docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env restart
  ```
- **Detener el stack:**
  ```bash
  docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env down
  ```
