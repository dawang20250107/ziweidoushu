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

# 数据库集成测试(需本机 PostgreSQL;TEST_DATABASE_URL 可自定义)
test-db:
	@test -n "$$TEST_DATABASE_URL" || (echo "请设置 TEST_DATABASE_URL,如 postgres://postgres@localhost:5432/postgres" && exit 1)
	go test ./internal/auth/ ./internal/store/ -v

docker:
	docker build -t ziweidoushu:latest .

clean:
	rm -rf bin
