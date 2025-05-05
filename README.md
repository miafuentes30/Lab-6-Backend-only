# Lab-6-Backend-only

# Backend

Backend para gestionar series de TV y Anime, desarrollado en Go y PostgreSQL.

## Requisitos

- Docker
- Docker Compose
- Go
- PostgreSQL

## Instalación

1. Clonar el repositorio:
   git clone https://github.com/miafuentes30/Lab-6-Backend-only.git
   cd backend

2. Descargar dependencias:
    go mod tidy

3. Documentación swagger
    go install github.com/swaggo/swag/cmd/swag@latest
    swag init -g main.go -o docs

4. Correr Docker
    docker-compose up --build