.PHONY: build run test clean docker-build docker-up docker-down

# 构建应用
build:
	go build -o bin/short-link ./cmd/server

# 构建admin管理工具
build-admin:
	go build -o bin/nsl-admin ./cmd/admin

# 运行应用
run:
	go run ./cmd/server

# 运行测试
test:
	go test -v ./...

# 清理构建文件
clean:
	rm -rf bin/
	rm -f *.log

# Docker构建
docker-build:
	docker-compose build

# Docker启动
docker-up:
	docker-compose up -d

# Docker停止
docker-down:
	docker-compose down

# 查看日志
docker-logs:
	docker-compose logs -f app

