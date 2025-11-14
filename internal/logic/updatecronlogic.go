package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/vars"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCronLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCronLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCronLogic {
	return &UpdateCronLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新任务
func (l *UpdateCronLogic) UpdateCron(in *cron.CronBody) (*cron.ResultRsp, error) {

	var data model.Cron
	copierx.MustCopy(&data, in)
	data.Id = in.Id
	if err := l.svcCtx.CronModel.Update(l.ctx, &data); err != nil {
		return nil, err
	}

	l.svcCtx.StartCron()
	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
