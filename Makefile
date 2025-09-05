include .env
DB_URL=$(DB_PATH)
MIGRATIONS_DIR=database/migrations
MAIN_FILE =cmd/main.go


.PHONY: run migrate-up migrate-down migrate-create

DB_URL=$(DB_PATH)

run:
	go run ${MAIN_FILE}

migrate-up:
	migrate -path database/migrations -database "$(DB_URL)" -verbose up

migrate-down:
	migrate -path database/migrations -database "$(DB_URL)" -verbose down 1

migrate-create:
	@echo "Podaj nazwę migracji (np. add_column):"
	@read name; \
	migrate create -ext sql -dir database/migrations -seq $$name
