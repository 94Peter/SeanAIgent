FROM golang:1.24-alpine AS builder

# 設置環境變數
ENV GO111MODULE=on \
GOPROXY=https://proxy.golang.org,direct \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

WORKDIR /build

RUN apk add --no-cache tzdata
# COPY go.mod go.sum ./
COPY . .
RUN go mod download


# 加上 -ldflags="-s -w" 移除除錯資訊，Image 會再縮小約 10-20MB
RUN go build -ldflags="-s -w" -o bot ./main.go

RUN cat <<EOF > /build/config.yaml
http:
  port: 8080
  mode: debug
  csrf:
    enabled: false
    field_name: _csrf
    secret: a-32-byte-long-secret-key-1234567890
    ignore_paths:
      - /line
      - /ig/webhook
  session:
    enabled: true
    store: mongo
    cookie_name: session
    max_age: 3600
    key_pairs: ["just_secret"]
  tracer:
    enabled: true
  logger:
    enabled: false

tracing:
  endpoint: jaeger.tracing.orb.local:4318

log:
  level: debug
  dev: true

database:
  uri: "mongodb://REPLACE_ME:27017/?compressors=zstd,snappy,zlib"
  db: "REPLACE_ME"
  max_conn_idle_time: 10m
  max_pool_size: 100
  min_pool_size: 20

linebot:
  channel_secret: REPLACE_ME
  channel_access_token: REPLACE_ME+kuh6qZPiJfIa2+y++91o0vqabnEFnyI2x3D/89s7EllaUcAnGUB4DPvkQ4mSwAiazYBVBYcENzm6qnC9UHim5b1k3NSnn1MXSYt5zGwK3yfHHCtJYh+AdB04t89/1O/w1cDnyilFU=
  flex_message:
    config_file: ./configs/flex_messages.yaml
    logo: https://developers-resource.landpress.line.me/fx/clip/clip13.jpg
  message:
    config_file: ./configs/messages.yaml
  admin_user_id: REPLACE_ME
  
liffids:
    booking: "REPLACE_ME"
    checkin: "REPLACE_ME"
    training_data: "REPLACE_ME"

service:
  response_templates: ./configs/response_messages.yaml

llm:
  model: gemini-2.5-flash-lite
  memory_collection: chat_history
  config_file: ./configs/llm.yaml
  googleai:
    api_key: REPLACE_ME
  mcp_server:
  - http://REPLACE_ME:9080/mcp
  
storage:
  provider: r2
  r2:
    endpoint: https://REPLACE_ME.com
    access_key_id: REPLACE_ME
    secret_access_key: REPLACE_ME
    bucket: REPLACE_ME

cron:
  user_stats_notify_url: http://REPLACE_ME:8080/line/notification/user-appt-stats
EOF



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

COPY --from=builder /build/config.yaml /app/conf/config.yaml
COPY --from=builder /build/bot /app/

# Expose both HTTP and gRPC ports
EXPOSE 8081

ENTRYPOINT ["/app/bot"]
CMD ["serve", "--config", "/app/conf/config.yaml"]