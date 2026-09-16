.PHONY: build run test vet clean verify format

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -count=1 -v ./...

vet:
	go vet ./...

format:
	gofmt -w .
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Unformatted files found:"; \
		gofmt -l .; \
		exit 1; \
	fi

verify: format vet test

clean:
	rm -rf bin data

