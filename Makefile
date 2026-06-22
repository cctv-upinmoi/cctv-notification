BINARY  = cctv-notification
CMD     = ./cmd/server

.PHONY: build run tidy lint docker-build docker-run

build:
	go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

tidy:
	go mod tidy

lint:
	go vet ./...

docker-build:
	docker build -t $(BINARY):latest .

docker-run:
	docker run --rm --env-file .env $(BINARY):latest
