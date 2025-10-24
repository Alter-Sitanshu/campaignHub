-include .env
MIGRATION_PATH = backend/database/migrations

.PHONY: startdb migratedown migrateup migrateforce migration server test
startdb:
	docker run project

migrateup:
	@migrate -path $(MIGRATION_PATH) -database $(MAKEDB) -verbose up

migratedown:
	@migrate -path $(MIGRATION_PATH) -database $(MAKEDB) -verbose down

migrateforce:
	@migrate -path ${MIGRATION_PATH} -database ${MAKEDB} force $(filter-out $@, $(MAKECMDGOALS))

migration:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(filter-out $@, $(MAKECMDGOALS))

test:
	go test -v ./backend/...

build:
	go build -o ./bin/main.exe main.go
	
server:
	go run main.go

%:
	@: