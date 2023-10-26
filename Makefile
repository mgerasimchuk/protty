test-unit: vendor-install
	docker run -v $(shell pwd):/app -w /app golang:1.20.0 /bin/bash \
	-c 'go test -v ./internal/... ./pkg/...'
lint-golangci: vendor-install
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.55.1 \
	golangci-lint run --timeout 5m30s -v
vendor-install:
	@if [ -d "vendor" ]; then echo "Vendor folder already exists. Skip vendor installing."; else docker run --rm -v $(shell pwd):/app -w /app golang:1.20.0 /bin/bash -c "go mod tidy && go mod vendor"; fi
