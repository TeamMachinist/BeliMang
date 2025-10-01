.PHONY: up-dev up-dev-build up-db up-redis down-dev down-reset down-clean

# Start all services in compose.dev.yml
up-dev: 
	docker compose -f compose.dev.yaml up -d

# Force rebuild dev + start all
# Use when Dockerfile or go.mod dependencies change
up-dev-build: 
	docker compose -f compose.dev.yaml up --build -d

# Individual services
up-db:
	docker compose -f compose.dev.yaml up -d postgres

up-redis:
	docker compose -f compose.dev.yaml up -d redis

# Stop all services  in compose.dev.yml
down-dev: 
	docker compose -f compose.dev.yaml down

# Stop + wipe database (fresh start)
down-reset: 
	docker compose -f compose.dev.yaml down -v

# Stop + cleanup Docker resources
down-clean: 
	docker compose -f compose.dev.yaml down -v
	docker container prune -f
	docker volume prune -f
	docker system prune -f

POSTGRES_USER ?= postgres
POSTGRES_DB ?= belimang
seed-dev:
	docker compose -f compose.dev.yaml exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) < ./seeds/01_seeds_data.sql