-include .env
MIGRATION_PATH = ./database/migrations

.PHONY: startdb migratedown migrateup migration server test
startdb:
	docker run project

migrateup:
	@migrate -path $(MIGRATION_PATH) -database $(MAKEDB) -verbose up

migratedown:
	@migrate -path $(MIGRATION_PATH) -database $(MAKEDB) -verbose down

migration:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(filter-out $@, $(MAKECMDGOALS))

test:
	go test -v ./...
	
server:
	go run main.go

