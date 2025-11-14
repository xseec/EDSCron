package logic

import (
	"context"
	"time"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHolidaysLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHolidaysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHolidaysLogic {
	return &GetHolidaysLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取假日列表
func (l *GetHolidaysLogic) GetHolidays(in *cron.HolidaysReq) (*cron.HolidaysRsp, error) {
	if err := expx.HasZeroError(in, "Address"); err != nil {
		return nil, err
	}

	area := cronx.EitherChinaOrTaiwan(in.Address)
	year := expx.If(in.Year == 0, time.Now().Year(), int(in.Year))
	all, err := l.svcCtx.HolidayModel.FindAllByAreaYear(l.ctx, string(area), year)
	if err != nil {
		return nil, err
	}

	var days []*cron.Holiday
	copierx.MustCopy(&days, all)

	return &cron.HolidaysRsp{
		Holidays: days,
	}, nil
}
