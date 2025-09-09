# ===============================
# HTTP-Server Makefile
# ===============================

include .env
include .env.dev

.PHONY: run migrate-up migrate-down migrate-create migrate-goto del-sess

run:
	go build -o $(BIN_DIR)/$(BINARY) $(MAIN_DIR)
	$(BIN_DIR)/$(BINARY)

migrate-up:
	@echo "Applying all migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_PATH)" -verbose up

migrate-down:
	@echo "Rolling back last migration..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_PATH)" -verbose down 1

migrate-goto:
	@echo "Podaj numer wersji (np. 1):"; \
	read version; \
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_PATH)" -verbose goto $$version

migrate-create:
	@echo "Podaj nazwę migracji (np. add_column):"; \
	read name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

del-sess:
	@echo "Truncating all sessions from DB..."
	@docker exec -i $(MYSQL_CONTAINER_NAME) \
	mysql -h $(MYSQL_HOST) -P $(MYSQL_PORT) -u $(MYSQL_USER) -p$(MYSQL_PASSWORD) $(MYSQL_DB) \
	-e "TRUNCATE TABLE fiber_storage;"

	@docker exec -i $(MYSQL_CONTAINER_NAME) \
	mysql -h $(MYSQL_HOST) -P $(MYSQL_PORT) -u $(MYSQL_USER) -p$(MYSQL_PASSWORD) $(MYSQL_DB) \
	-e "TRUNCATE TABLE interactions;"
 
