package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddHolidaysLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddHolidaysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddHolidaysLogic {
	return &AddHolidaysLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建/更新假日
func (l *AddHolidaysLogic) AddHolidays(in *cron.AddHolidaysReq) (*cron.ResultRsp, error) {
	if err := expx.HasZeroError(in, "Address", "Holidays"); err != nil {
		return nil, err
	}

	area := cronx.EitherChinaOrTaiwan(in.Address)

	var many []model.Holiday
	copierx.MustCopy(&many, in.Holidays)

	if err := l.svcCtx.HolidayModel.AddMany(l.ctx, string(area), &many); err != nil {
		return nil, err
	}

	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
