include .env

# Variables
PROJECT_NAME = simplebank
POSTGRES_CONTAINER = $(PROJECT_NAME)-postgres
POSTGRES_USER ?= root
POSTGRES_DB ?= simple_bank

createdb:
	docker exec -it $(POSTGRES_CONTAINER) createdb --username=$(POSTGRES_USER) --owner=${POSTGRES_USER} ${POSTGRES_DB}

dropdb:
	docker exec -it $(POSTGRES_CONTAINER) dropdb ${POSTGRES_DB}

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/${POSTGRES_DB}?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/${POSTGRES_DB}?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PNOHY: createdb dropdb migrateup migratedown sqlc test
