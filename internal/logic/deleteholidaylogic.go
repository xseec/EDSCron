package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteHolidayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteHolidayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteHolidayLogic {
	return &DeleteHolidayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除假日
func (l *DeleteHolidayLogic) DeleteHoliday(in *cron.DelReq) (*cron.ResultRsp, error) {
	if err := expx.HasZeroError(in, "Id"); err != nil {
		return nil, err
	}

	err := l.svcCtx.CronModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	l.svcCtx.StartCron()
	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
