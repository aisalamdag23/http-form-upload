setup:
	@echo ">> building docker images"
	@docker-compose up -d
	@echo ">> setting environment variables"
	@-yes no | cp -i .env.example .env
	@echo ">> setting up databases"
	@-PGPASSWORD=S3cret createdb -U rootuser -h localhost -p 5432 http-form-upload_dev "development database"
	@goose --dir db_migrations postgres postgres://rootuser:S3cret@localhost:5432/http-form-upload_dev?sslmode=disable up

tools: # Install all the tools
	@go install github.com/pressly/goose/v3/cmd/goose@latest

test: 
	go test -cover -coverprofile=test-cov.out ./...

run:
	go run .

.PHONY: setup