.PHONY: build run test test-race vet fmt bench docker clean

BINARY := bin/ziweidoushu

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/server

run: build
	./$(BINARY)

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal data

# 排盘引擎微基准
bench:
	go test ./internal/ziwei/ -bench . -benchmem -run '^$$'

docker:
	docker build -t ziweidoushu:latest .

clean:
	rm -rf bin
