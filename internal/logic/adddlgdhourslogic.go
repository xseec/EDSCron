package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddDlgdHoursLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddDlgdHoursLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddDlgdHoursLogic {
	return &AddDlgdHoursLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 新增用电时段
func (l *AddDlgdHoursLogic) AddDlgdHours(in *cron.AddDlgdHourReq) (*cron.ResultRsp, error) {
	err := expx.HasZeroError(in, "Area", "Comment")
	if err != nil {
		return nil, err
	}

	err = l.svcCtx.DlgdHourModel.InsertAll(l.ctx, in.Area, in.Comment)
	if err != nil {
		return nil, err
	}

	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
