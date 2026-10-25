# Основные команды разработки.
.PHONY: help backend-deps backend-run backend-test frontend-install frontend-run frontend-test up down

help:
	@echo "make backend-deps     - загрузить зависимости Go (go mod tidy)"
	@echo "make backend-run      - запустить backend с хранилищем в памяти на :8080"
	@echo "make backend-test     - тесты backend (с TEST_DATABASE_URL — и интеграционные)"
	@echo "make frontend-install - npm install"
	@echo "make frontend-run     - dev-сервер frontend на :5173 (прокси /api -> :8080)"
	@echo "make frontend-test    - тесты и проверка типов frontend"
	@echo "make up / down        - запуск/остановка всей системы в Docker"

backend-deps:
	cd backend && go mod tidy

backend-run: backend-deps
	cd backend && STORAGE=memory ADMIN_PASSWORD=admin go run ./cmd/server

backend-test: backend-deps
	cd backend && go vet ./... && go test -race -cover ./...

frontend-install:
	cd frontend && npm install

frontend-run:
	cd frontend && npm run dev

frontend-test:
	cd frontend && npm run typecheck && npm test

up:
	@test -f infra/.env || cp infra/.env.example infra/.env
	docker compose -f infra/docker-compose.yml --env-file infra/.env up --build -d
	@echo "Приложение: http://localhost:8081"

down:
	docker compose -f infra/docker-compose.yml --env-file infra/.env down
