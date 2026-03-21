.PHONY: dev down migrate-up migrate-down migrate-status

dev:
	docker-compose up --build

down:
	docker-compose down

migrate-up:
	export $$(grep -v '^#' .env | xargs) && \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$${DB_URL}" goose -dir internal/migrations up

migrate-down:
	export $$(grep -v '^#' .env | xargs) && \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$${DB_URL}" goose -dir internal/migrations down

migrate-status:
	export $$(grep -v '^#' .env | xargs) && \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$${DB_URL}" goose -dir internal/migrations status
