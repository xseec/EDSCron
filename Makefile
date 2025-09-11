
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

## docker: 构建 docker 镜像
docker:
	@echo "Building docker image..."
	docker build -t cron:latest .

## docker-run: 运行 docker 容器
docker-run:
	docker start mysql
	docker start redis
	@echo "Checking for existing container..."
	-docker rm -f cron 2>NUL || cmd /c exit 0
	-docker rm -f cron 2>/dev/null || true
	@echo "Running docker container: cron..."
	docker run -d -it --name cron -p 8123:8123 --network eds -v //d/env/docker/cron/.env:/app/.env -v //d/env/docker/cron/logs:/app/logs cron

## docker-output: 导出 docker 镜像到共享文件夹
docker-output:
	@echo "Connecting to docker-machine and exporting image..."
	docker-machine ssh default "docker save -o /docker-share/cron-image.tar cron:latest"
	@echo "Image exported successfully to shared folder"
