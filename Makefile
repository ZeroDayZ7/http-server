include .env
DB_PATH=mysql://${MYSQL_DSN}
MIGRATIONS_DIR=database/migrations
MAIN_FILE =cmd/main.go


.PHONY: run migrate-up migrate-down migrate-create

run:
	go run ${MAIN_FILE}

migrate-up:
	migrate -path database/migrations -database "$(DB_PATH)" -verbose up

migrate-down:
	migrate -path database/migrations -database "$(DB_PATH)" -verbose down 1

migrate-goto:
	@echo "Podaj numer wersji do której chcesz cofnąć lub przejść (np. 1):"
	@read version; \
	migrate -path database/migrations -database "$(DB_PATH)" -verbose goto $$version


migrate-create:
	@echo "Podaj nazwę migracji (np. add_column):"
	@read name; \
	migrate create -ext sql -dir database/migrations -seq $$name
