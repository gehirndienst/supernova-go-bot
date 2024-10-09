PHONY: build run clean test dep lint

run:
	go run cmd/runner/run.go -env-file .env

migrate-up:
	go run cmd/migration/migrate.go -direction up -migration-path migrations -env-file .env

migrate-down:
	go run cmd/migration/migrate.go -direction down -migration-path migrations -env-file .env

clean:
	go clean
	rm ${BINARY_NAME}-l

test:
	go test -v ./...

test_coverage:
	go test -v ./... -coverprofile=cov.out

dep:
	go mod download

lint:
	golangci-lint run
