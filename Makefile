.PHONY: dev dev-be dev-fe build docker docker-run

# Run backend and frontend concurrently in development
dev:
	make -j2 dev-be dev-fe

# Go backend with hot reload via Air (like ts-node-dev in NestJS)
dev-be:
	cd backend && air

# Vue frontend with HMR via Vite (like ng serve in Angular)
dev-fe:
	cd frontend && pnpm dev

# Build frontend then compile Go binary with embedded assets
build:
	cd frontend && pnpm install && pnpm build
	cp -r frontend/dist backend/internal/ui/dist
	cd backend && go build -o ../dist/toggleflow ./cmd/server

# Build Docker image
docker:
	docker build -t toggleflow .

# Run Docker image locally
docker-run:
	mkdir -p data
	docker run -p 8080:8080 \
		-v ./data:/data \
		-e ADMIN_TOKEN=dev-secret \
		toggleflow
