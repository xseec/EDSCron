package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmDlgdHoursLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmDlgdHoursLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmDlgdHoursLogic {
	return &ConfirmDlgdHoursLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 确认用电时段
func (l *ConfirmDlgdHoursLogic) ConfirmDlgdHours(in *cron.DlgdHourReq) (*cron.ResultRsp, error) {
	err := expx.HasZeroError(in, "Area")
	if err != nil {
		return nil, err
	}

	err = l.svcCtx.DlgdHourModel.ConfirmAll(l.ctx, in.Area, in.DocNo)
	if err != nil {
		return nil, err
	}

	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
