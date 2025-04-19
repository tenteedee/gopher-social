include .env

MIGRATIONS_PATH := cmd/migrate/migrations
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(filter-out $@, $(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose up

.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose down

.PHONY: migrate-status
migrate-status:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose version

.PHONY: migrate-force
migrate-force:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose force 1

.PHONY: seed
seed:
	go run cmd/migrate/seed/main.go	