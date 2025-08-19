package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/vars"

	"github.com/jinzhu/copier"

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
	err := svc.Format(in)
	if err != nil {
		return nil, err
	}

	var data model.Cron
	copier.Copy(&data, in)
	data.Id = in.Id
	err = l.svcCtx.CronModel.Update(l.ctx, &data)
	if err != nil {
		return nil, err
	}

	l.svcCtx.StartCron()
	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
