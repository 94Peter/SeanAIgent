FROM golang:1.24-alpine AS builder

# 設置環境變數
ENV GO111MODULE=on \
GOPROXY=https://proxy.golang.org,direct \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

WORKDIR /build

COPY . .
RUN go mod download

RUN go build -o bot ./main.go

###################
# multi-stage build
###################
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/


WORKDIR /app
COPY ./configs /app/configs
COPY ./assets /app/assets

COPY --from=builder /build/bot /app/

# Expose both HTTP and gRPC ports
EXPOSE 8081

ENTRYPOINT ["/app/bot"]
CMD ["serve", "--config", "/app/conf/config.yaml"]