package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QuickStartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQuickStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuickStartLogic {
	return &QuickStartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 快速开始
func (l *QuickStartLogic) QuickStart(in *cron.QuickStartReq) (*cron.ResultRsp, error) {
	// todo: add your logic here and delete this line

	return &cron.ResultRsp{}, nil
}
