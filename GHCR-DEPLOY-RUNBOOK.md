# Reporte de Arquitectura y Despliegue: Framework vía GHCR en Proxmox

Este documento resume la infraestructura preparada y el procedimiento para desplegar el proyecto **Framework** en la máquina virtual `104` (`framework`) en Proxmox VE sin exponer un Self-Hosted Runner en un repositorio público.

---

## 1. Estado de la Infraestructura en Proxmox

- **Host Proxmox:** `twintailserver` (`10.10.10.1`, Tailscale: `100.83.236.35`).
- **Template Golden 9000:** Actualizado a **Debian 12 Bookworm** (`debian-12-genericcloud-amd64.qcow2`) con soporte Cloud-Init, QEMU Guest Agent, Docker CE, Docker Compose y Podman.
- **VM 104 (`framework`):**
  - IP: `10.10.10.90/24` (Gateway: `10.10.10.1`).
  - Usuario: `uadmin` (con permisos de `sudo` sin contraseña y grupo `docker`).

---

## 2. Estrategia de Despliegue: CI en GitHub vs. Ejecución en Servidor

### ¿Por qué NO usar un Self-Hosted Runner en este caso?
Dado que el repositorio `yan2005dris-afk/Framework` es **público**, registrar un runner self-hosted conectado a tu red interna permite que un tercero (mediante un Pull Request) ejecute código arbitrario en tu infraestructura.

### Arquitectura Adoptada: Build en GitHub + Pull en Proxmox
1. **GitHub Actions (Cloud `ubuntu-latest`):** Compila las imágenes Docker de cada servicio de forma gratuita y segura mediante `.github/workflows/deploy.yml`.
2. **GHCR (GitHub Container Registry):** Almacena las imágenes compiladas bajo `ghcr.io/yan2005dris-afk/*`.
3. **VM 104 (`framework`):** Descarga las imágenes (`docker compose pull`) y las levanta (`docker compose up -d`).

---

## 3. Guía de Puesta en Marcha en la VM 104

Conectarse a la máquina virtual:
```bash
ssh uadmin@10.10.10.90
```

### Paso 1: Clonar el repositorio
```bash
sudo mkdir -p /srv/framework
sudo chown uadmin:uadmin /srv/framework
git clone https://github.com/yan2005dris-afk/Framework.git /srv/framework
cd /srv/framework
```

### Paso 2: Autenticación en GHCR (si las imágenes son privadas)
Si los paquetes en GHCR son privados, generar un **Personal Access Token (classic)** en GitHub con permiso `read:packages` y autenticarse en la VM:
```bash
echo "<TU_GITHUB_PAT>" | docker login ghcr.io -u yan2005dris-afk --password-stdin
```
*(Si los paquetes son públicos en tu perfil de GitHub, no es necesario hacer login).*

### Paso 3: Configurar variables de entorno y levantar
```bash
cp .env.example .env
# Ajustar valores de producción en .env
docker compose pull
docker compose up -d
```

---

## 4. Opciones para Automatizar Nuevos Deploys

Para que la VM 104 se actualice automáticamente al hacer `push` a `main`:

### Opción A: Watchtower (Recomendada, Cero Fricción)
Agregar el contenedor de Watchtower al `docker-compose.yml` en la VM. Monitorea periódicamente GHCR y actualiza los servicios cuando hay una imagen nueva:

```yaml
services:
  watchtower:
    image: containrrr/watchtower
    container_name: watchtower
    restart: always
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /home/uadmin/.docker/config.json:/config.json:ro
    command: --interval 300 --cleanup
```

### Opción B: Disparador SSH desde GitHub Actions
Agregar un step final en `.github/workflows/deploy.yml` que se conecte por SSH a `10.10.10.90` (vía Tailscale) tras publicar las imágenes:
```yaml
- name: Deploy to Proxmox VM
  uses: appleboy/ssh-action@v1.0.3
  with:
    host: 10.10.10.90
    username: uadmin
    key: ${{ secrets.SSH_PRIVATE_KEY }}
    script: |
      cd /srv/framework
      docker compose pull
      docker compose up -d
```
