.PHONY: run build clean test

run:
	go run main.go

build:
	go build -o potato-service main.go

clean:
	rm -f potato-service

test:
	go test ./...

deps:
	go mod download

tidy:
	go mod tidy

