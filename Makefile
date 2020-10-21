.PHONY: migrate drop

migrate:
	migrate -path db/migrations -database ${DATABASE_URL} -verbose up

drop:
	migrate -path db/migrations -database ${DATABASE_URL} -verbose down

sqlc:
	sqlc generate