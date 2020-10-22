.PHONY: migrate drop sqlc test run

migrate:
	migrate -path db/migrations -database ${DATABASE_URL} -verbose up

drop:
	migrate -path db/migrations -database ${DATABASE_URL} -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

run:
	go run main.go