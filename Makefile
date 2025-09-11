
MYSQL := "root:123456@tcp(127.0.0.1:3306)/eds_cron"

.PHONY: model

## model: 根据数据库连接生成/更新 model 代码
model:
	@echo "Generating model..."
	cd model && goctl model mysql datasource -url="$(MYSQL)" -table="*"  -dir="." -c

## rpc: 根据 proto 文件生成/更新 rpc 代码
rpc:
	@echo "Generating rpc code..."
	goctl rpc protoc cron.proto --go_out=. --go-grpc_out=. --zrpc_out=.

## test: 运行测试
test:
	@echo "Running tests..."
	go test -v ./...