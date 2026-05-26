.PHONY: dev dev-be dev-fe build docker docker-run lint lint-be lint-fe

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

# Lint both frontend and backend
lint:
	make -j2 lint-be lint-fe

# Run golangci-lint on the backend (install: brew install golangci-lint)
lint-be:
	cd backend && golangci-lint run ./...

# Run ESLint on the frontend
lint-fe:
	cd frontend && pnpm lint:check

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
