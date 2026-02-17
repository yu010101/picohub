.PHONY: dev dev-backend dev-frontend build seed clean install lint

# Development - run both backend and frontend
# Uses trap to kill both processes on Ctrl+C
dev:
	@echo "Starting backend on :8080 and frontend on :3000"
	@trap 'kill 0' INT TERM; \
		cd backend && CGO_ENABLED=1 go run -tags fts5 . & \
		cd frontend && npm run dev & \
		wait

dev-backend:
	cd backend && CGO_ENABLED=1 go run -tags fts5 .

dev-frontend:
	cd frontend && npm run dev

# Build
build: build-backend build-frontend

build-backend:
	cd backend && CGO_ENABLED=1 go build -tags fts5 -o picohub .

build-frontend:
	cd frontend && npm run build

# Seed database (backend will auto-seed on first run)
seed:
	cd backend && CGO_ENABLED=1 go run -tags fts5 .

# Install dependencies
install:
	cd backend && go mod tidy
	cd frontend && npm install

# Clean build artifacts (keeps sample skill zips)
clean:
	rm -f backend/picohub
	rm -f backend/picohub.db backend/picohub.db-wal backend/picohub.db-shm
	rm -rf frontend/.next

# Lint
lint:
	cd backend && CGO_ENABLED=1 go vet -tags fts5 ./...
	cd frontend && npm run lint
