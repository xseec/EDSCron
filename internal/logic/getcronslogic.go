package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/copierx"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCronsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCronsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCronsLogic {
	return &GetCronsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取任务列表
func (l *GetCronsLogic) GetCrons(in *cron.CronsReq) (*cron.CronsRsp, error) {
	crons, err := l.svcCtx.CronModel.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	var values []*cron.CronBody
	copierx.Copy(&values, crons)

	return &cron.CronsRsp{
		Crons: values,
	}, nil
}
