package main

import (
	"flag"
	"fmt"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/config"
	"seeccloud.com/edscron/internal/data"
	"seeccloud.com/edscron/internal/server"
	"seeccloud.com/edscron/internal/svc"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/cron.yaml", "the config file")

func main() {
	// load .env if exist
	_ = godotenv.Load()

	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.Log)

	// 执行数据库迁移
	if err := data.Migrate(c.MySql.DataSource, c.Migrate.Dir, c.Migrate.AutoRun); err != nil {
		logx.Errorf("数据库迁移失败: %v", err)
		return
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		cron.RegisterCronServer(grpcServer, server.NewCronServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	ctx.StartCron()

	fmt.Printf("Starting eds cron rpc server at %s...\n", c.ListenOn)
	s.Start()
}
