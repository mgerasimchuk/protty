lint-golangci: vendor-install
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.55.1 \
	golangci-lint run --timeout 5m30s -v
vendor-install:
	@if [ -d "vendor" ]; then echo "Vendor folder already exists. Skip vendor installing."; else docker run --rm -v $(shell pwd):/app -w /app golang:1.20.0 /bin/bash -c "go mod tidy && go mod vendor"; fi
test-unit: vendor-install
	docker run -v $(shell pwd):/app -w /app golang:1.20.0 /bin/bash \
	-c 'go test -covermode=count -coverprofile=assets/coverage/unit.out -tags=unit -v ./... && go tool cover -html=assets/coverage/unit.out -o=assets/coverage/unit.html'
test-integration: vendor-install
	docker run -v $(shell pwd):/app -w /app golang:1.20.0 /bin/bash \
	-c 'go test -covermode=count -coverprofile=assets/coverage/integration.out -tags=integration -v ./... && go tool cover -html=assets/coverage/integration.out -o=assets/coverage/integration.html'