# ── 构建阶段 ────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder
WORKDIR /src

# 依赖层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 纯静态二进制;数据资产经 go:embed 打进二进制,运行时零外部文件依赖
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X github.com/dawang20250107/ziweidoushu/internal/httpapi.Version=$(cat VERSION 2>/dev/null || echo 2.0.0)" \
    -o /out/ziweidoushu ./cmd/server

# ── 运行阶段 ────────────────────────────────────────────────────
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=builder /out/ziweidoushu /app/ziweidoushu

ENV LISTEN_ADDR=:8080
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/ziweidoushu"]
