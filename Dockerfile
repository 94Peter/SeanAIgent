FROM golang:1.24-alpine AS builder

# 設置環境變數
ENV GO111MODULE=on \
GOPROXY=https://proxy.golang.org,direct \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

WORKDIR /build

COPY . .
RUN apk add --no-cache tzdata
RUN go mod download

RUN go build -o bot ./main.go

###################
# multi-stage build
###################
FROM scratch
# 1. 複製 SSL/TLS 憑證
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 複製時區資料
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app
COPY ./configs /app/configs
COPY ./assets /app/assets

COPY --from=builder /build/bot /app/

# Expose both HTTP and gRPC ports
EXPOSE 8081

ENTRYPOINT ["/app/bot"]
CMD ["serve", "--config", "/app/conf/config.yaml"]