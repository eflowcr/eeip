# Executive Email Intelligence Platform (EEIP)

EEIP es una plataforma diseñada para actuar como un asistente ejecutivo inteligente para la gestión del correo electrónico corporativo.

## Arquitectura

- **Backend**: Golang, Clean Architecture, DDD, PostgreSQL.
- **Frontend**: Angular 18, TailwindCSS.
- **Infraestructura**: Docker Compose, GitHub Actions.

## Requisitos

- Docker y Docker Compose instalados.

## Inicialización y Despliegue

Para desplegar localmente:

1. Renombrar o crear `.env` basado en las variables del `docker-compose.yml` (Asegúrate de configurar `OPENAI_API_KEY`).
2. Ejecutar:
   ```bash
   make deploy
   ```
   *Alternativamente*: `docker compose up -d`

## Servicios y Puertos

- **Backend API**: `localhost:10000`
- **Frontend App**: `localhost:11000`
- **PostgreSQL**: `localhost:5432`

## Endpoints Principales (API)

- `GET /api/v1/health` - Healthcheck
- `GET /api/v1/emails/important` - Retorna los correos clasificados como Critical/High
- `GET /api/v1/accounts/:id/emails` - Retorna el Global Executive Inbox por cuenta

## Migraciones

Las migraciones de base de datos se ejecutan de manera automática usando `golang-migrate` al iniciar el backend (Ver `cmd/api/main.go`).

## CI/CD

El proyecto cuenta con un workflow de GitHub Actions que:
- Ejecuta test automáticos.
- Construye las imágenes Docker.
- Pushea las imágenes al Docker Registry.
- Despliega por SSH a producción de forma autónoma.
