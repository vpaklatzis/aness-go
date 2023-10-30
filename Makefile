.PHONY: run-prod run-dev build proto run-implant-dev run-implant-prod

run-dev:
	ENVIRONMENT=dev go run main.go

run-prod:
	ENVIRONMENT=test go run main.go

build:
	go build

migrateup:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -verbose down

proto:
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/implant.proto

run-implant-dev:
	ENVIRONMENT=dev go run implant/implant.go

run-implant-prod:
	ENVIRONMENT=test go run /implant/implant.go