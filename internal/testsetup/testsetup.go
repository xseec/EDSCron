package testsetup

import (
	"context"
	"flag"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"seeccloud.com/edscron/internal/config"
	"seeccloud.com/edscron/internal/svc"
)

var configFile = flag.String("f", "etc/cron.yaml", "the config file")

type TestSetup struct {
	Ctx    context.Context
	Ctrl   *gomock.Controller
	SvcCtx *svc.ServiceContext
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// cron > internal > logic(测试文件)
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func SetupTest(t *testing.T) *TestSetup {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	rootDir := getProjectRoot()

	// load .env if exist
	_ = godotenv.Load(filepath.Join(rootDir, ".env"))

	flag.Parse()
	var c config.Config
	conf.MustLoad(filepath.Join(rootDir, *configFile), &c, conf.UseEnv())
	svcCtx := svc.NewServiceContext(c)

	return &TestSetup{
		Ctx:    ctx,
		Ctrl:   ctrl,
		SvcCtx: svcCtx,
	}
}
