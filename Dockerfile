# 第一阶段：构建阶段, 替换镜像源
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:alpine3.21 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=1 \
    GOPROXY=https://goproxy.cn,direct

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache \
        gcc \
        musl-dev \
        upx && \   
    rm -rf /var/cache/apk/*

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/cron cron.go && \
    upx --lzma /app/cron

# 第二阶段：运行阶段 - 使用更小的基础镜像
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/alpine:3.21

# 安装运行时依赖并清理缓存, 固定chromium版本，避免更新导致的兼容性问题
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache \
        chromium=136.0.7103.113-r0 \
        chromium-chromedriver=136.0.7103.113-r0 \  
        nss=3.109-r0 \
        freetype \
        freetype-dev \
        harfbuzz \
        ca-certificates \
        ttf-freefont \
        libstdc++ \
        tzdata && \
    rm -rf /var/cache/apk/* && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_LAUNCHER_OPTS="--headless --disable-gpu --no-sandbox --disable-dev-shm-usage"


WORKDIR /app
COPY --from=builder /app/cron /app/cron
COPY ./etc /app/etc

CMD ["./cron", "-f", "etc/cron.yaml"]