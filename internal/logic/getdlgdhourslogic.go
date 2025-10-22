package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDlgdHoursLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDlgdHoursLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDlgdHoursLogic {
	return &GetDlgdHoursLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用电时段
func (l *GetDlgdHoursLogic) GetDlgdHours(in *cron.DlgdHourReq) (*cron.DlgdHoursRsp, error) {
	err := expx.HasZeroError(in, "Area")
	if err != nil {
		return nil, err
	}

	all, err := l.svcCtx.DlgdHourModel.QueryAll(l.ctx, in.Area, in.DocNo)
	if err != nil {
		return nil, err
	}

	var hours []*cron.DlgdHour
	copierx.Copy(&hours, all)

	return &cron.DlgdHoursRsp{
		Hours: hours,
	}, nil
}
