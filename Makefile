APP_PORT := $(or ${port},${port},"8080")

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

run: ## Run app. default port is `8080`. You can change it with `make port=8081 run`
	@echo "Running on port ${APP_PORT}"
	go mod vendor
	go run ./cmd/main.go -port=${APP_PORT}

lint: ## Check code style
	golangci-lint run --verbose

test: ## Run tests
	go test -race ./...

curl: ## Run sample request synchronously
	curl -X POST -H "Content-Type: application/json" -d @req.json http://localhost:8080/v1/resize

curl_async: ## Run sample request asynchronously
	curl -X POST -H "Content-Type: application/json" -d @req.json http://localhost:8080/v1/resize?async=true