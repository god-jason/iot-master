#!/bin/bash

# 编译Linux版本
export GOOS=linux

go build -o master.run .
go build -o gateway.run ./cmd/gateway/
# 集成测试设备（需要 -tags integration，默认跳过）
# go test -tags integration -c -o test-device.run ./test/

# 编译Windows版本
export GOOS=windows
