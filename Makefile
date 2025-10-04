.PHONY: up-dev up-dev-build up-prod up-prod-build up-db up-redis down-dev down-prod down-reset down-clean build-images push-images k8s-deploy k8s-cleanup k3s-deploy k3s-cleanup seed-dev

# Docker Registry Configuration
REGISTRY ?= docker.io
REGISTRY_USER ?= your-username
IMAGE_TAG ?= latest
POSTGRES_IMAGE = $(REGISTRY)/$(REGISTRY_USER)/belimang-postgres:$(IMAGE_TAG)
APP_IMAGE = $(REGISTRY)/$(REGISTRY_USER)/belimang-app:$(IMAGE_TAG)

# Development Commands
up-dev: 
	docker compose -f compose.dev.yaml up -d

up-dev-build: 
	docker compose -f compose.dev.yaml up --build -d

# Production Commands
up-prod:
	docker compose -f compose.yaml up -d

up-prod-build:
	docker compose -f compose.yaml up --build -d

# Individual services
up-db:
	docker compose -f compose.dev.yaml up -d postgres

up-redis:
	docker compose -f compose.dev.yaml up -d redis

# Stop services
down-dev: 
	docker compose -f compose.dev.yaml down

down-prod:
	docker compose -f compose.yaml down

# Stop + wipe database (fresh start)
down-reset: 
	docker compose -f compose.dev.yaml down -v

# Stop + cleanup Docker resources
down-clean: 
	docker compose -f compose.dev.yaml down -v
	docker container prune -f
	docker volume prune -f
	docker system prune -f

# Build Docker images for registry
build-images:
	@echo "Building PostgreSQL image..."
	docker build -t $(POSTGRES_IMAGE) --target postgres-h3 .
	@echo "Building application image..."
	docker build -t $(APP_IMAGE) --target production .
	@echo "Images built successfully!"

# Push images to registry
push-images: build-images
	@echo "Pushing PostgreSQL image..."
	docker push $(POSTGRES_IMAGE)
	@echo "Pushing application image..."
	docker push $(APP_IMAGE)
	@echo "Images pushed successfully!"

# Login to Docker registry
docker-login:
	@echo "Logging in to $(REGISTRY)..."
	docker login $(REGISTRY)

# Build and push with login
deploy-images: docker-login push-images

# Kubernetes deployment
k8s-deploy:
	@echo "Updating Kubernetes manifests with registry images..."
	@sed -i.bak 's|belimang-postgres:latest|$(POSTGRES_IMAGE)|g' k8s/postgres-deployment.yaml
	@sed -i.bak 's|belimang-app:latest|$(APP_IMAGE)|g' k8s/app-deployment.yaml
	@sed -i.bak 's|imagePullPolicy: Never|imagePullPolicy: Always|g' k8s/postgres-deployment.yaml k8s/app-deployment.yaml
	@echo "Deploying to Kubernetes..."
	cd k8s && ./deploy.sh deploy
	@echo "Restoring original manifests..."
	@mv k8s/postgres-deployment.yaml.bak k8s/postgres-deployment.yaml
	@mv k8s/app-deployment.yaml.bak k8s/app-deployment.yaml

# Kubernetes cleanup
k8s-cleanup:
	cd k8s && ./deploy.sh cleanup

# K3s deployment (lightweight Kubernetes)
k3s-deploy:
	@echo "Deploying to K3s..."
	cd k8s && ./deploy-k3s.sh deploy

# K3s cleanup
k3s-cleanup:
	cd k8s && ./deploy-k3s.sh cleanup

# Database seeding
POSTGRES_USER ?= postgres
POSTGRES_DB ?= belimang
seed-dev:
	docker compose -f compose.dev.yaml exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) < ./seeds/01_seeds_data.sql

# Help command
help:
	@echo "Available commands:"
	@echo "  Development:"
	@echo "    up-dev          - Start development environment"
	@echo "    up-dev-build    - Build and start development environment"
	@echo "    down-dev        - Stop development environment"
	@echo "    down-reset      - Stop and reset development environment"
	@echo ""
	@echo "  Production:"
	@echo "    up-prod         - Start production environment"
	@echo "    up-prod-build   - Build and start production environment"
	@echo "    down-prod       - Stop production environment"
	@echo ""
	@echo "  Docker Registry:"
	@echo "    build-images    - Build Docker images for registry"
	@echo "    push-images     - Build and push images to registry"
	@echo "    docker-login    - Login to Docker registry"
	@echo "    deploy-images   - Login, build and push images"
	@echo ""
	@echo "  Kubernetes:"
	@echo "    k8s-deploy      - Deploy to Kubernetes using registry images"
	@echo "    k8s-cleanup     - Clean up Kubernetes deployment"
	@echo ""
	@echo "  K3s (Lightweight Kubernetes):"
	@echo "    k3s-deploy      - Deploy to K3s using local images"
	@echo "    k3s-cleanup     - Clean up K3s deployment"
	@echo ""
	@echo "  Database:"
	@echo "    seed-dev        - Seed development database"
	@echo ""
	@echo "  Configuration:"
	@echo "    REGISTRY=docker.io (default)"
	@echo "    REGISTRY_USER=your-username (required for registry operations)"
	@echo "    IMAGE_TAG=latest (default)"