-include ../.env
MIGRATION_PATH = database/migrations

.PHONY: startdb migratedown migrateup migrateforce migration server test fresh-test
startdb:
	docker run project

migrateup:
	@migrate -path $(MIGRATION_PATH) -database ${DB_ADDR} -verbose up

migratedown:
	@migrate -path $(MIGRATION_PATH) -database ${DB_ADDR} -verbose down

migrateforce:
	@migrate -path ${MIGRATION_PATH} -database ${DB_ADDR} force $(filter-out $@, $(MAKECMDGOALS))

migration:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(filter-out $@, $(MAKECMDGOALS))

test:
	go test -v ./...

fresh-test:
	go clean -testcache
	go test -v ./...

build:
	go build -o ./bin/main.exe main.go
	
server:
	go run main.go

%:
	@: