package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
	"seeccloud.com/edscron/pkg/cronx"
)

type Config struct {
	zrpc.RpcServerConf
	MySql struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	Mail       cronx.MailConfig
	Ocr        struct {
		Endpoint        string
		AccessKeyId     string
		AccessKeySecret string
	}
}
