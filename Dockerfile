FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o dmp-master .
RUN go build -o dmp-gateway ./cmd/gateway
# RUN go test -tags integration -c -o dmp-device ./test   # 集成测试设备

FROM alpine:latest

RUN apk --no-redis add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/dmp-master .
COPY --from=builder /app/dmp-gateway .
# COPY --from=builder /app/dmp-device .   # 集成测试设备

COPY master.yaml master.yaml
COPY dbus.yaml dbus.yaml

EXPOSE 8080 8081 1883

CMD ["./dmp-master"]
