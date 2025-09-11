package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"
	"seeccloud.com/edscron/pkg/x/slicex"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPriceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPriceLogic {
	return &GetPriceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取电价
func (l *GetPriceLogic) GetPrice(in *cron.PriceReq) (*cron.PriceRsp, error) {
	if err := expx.HasZeroError(in, "Category", "Time"); err != nil {
		return nil, err
	}

	t, err := time.ParseInLocation(vars.DatetimeFormat, in.Time, time.Local)
	if err != nil {
		return nil, err
	}

	if slicex.Contains(cronx.TwdlCategories, in.Category) {
		return GetTwdlPrice(in.Category, t, l)
	}

	return GetDlgdPrice(in.Category, t, l)
}

func GetTwdlPrice(category string, t time.Time, l *GetPriceLogic) (*cron.PriceRsp, error) {

	dayStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	one, err := l.svcCtx.TwdlModel.FindOneByDayStartTimeCategory(l.ctx, dayStart.Format(vars.DatetimeFormat), category)
	if err != nil {
		return nil, err
	}

	holiday, _ := l.svcCtx.HolidayModel.FindOneByAreaDateCache(l.ctx, string(cronx.TaiwanArea), t.Format(vars.DateFormat))

	price := one.GetPrice(t.Format(vars.DatetimeFormat), holiday != nil && holiday.Category == string(cronx.HolidayPeakOff))
	var rsp cron.PriceRsp
	copier.Copy(&rsp, price)
	return &rsp, nil
}

func GetDlgdPrice(category string, t time.Time, l *GetPriceLogic) (*cron.PriceRsp, error) {
	infos := strings.Split(category, cronx.CategorySep)
	if len(infos) < 3 {
		return nil, fmt.Errorf("req.Category格式错误, 正确格式: %s", model.CategoryFormatTip)
	}

	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	one, err := l.svcCtx.DlgdModel.FindFirstByAreaStartTimeCategoryVoltage(l.ctx, infos[0], start.Format(vars.DatetimeFormat), infos[1], infos[2])
	if err == model.ErrNotFound {
		nearlyOne, err := l.svcCtx.DlgdModel.FindOneByAreaCategoryVoltageAtNearlyStartTime(l.ctx, infos[0], start.Format(vars.DatetimeFormat), infos[1], infos[2])
		if err != nil || nearlyOne == nil {
			return nil, fmt.Errorf("未找到%s>%d年%d月>%s>%s电价表", infos[0], t.Year(), t.Month(), infos[1], infos[2])
		}

		one = nearlyOne
		one.StartTime = start
		one.EndTime = start.AddDate(0, 1, 0)
	} else if err != nil {
		return nil, err
	}

	period, err := l.svcCtx.GetDlgdPrice(t, one)
	if err != nil {
		return nil, err
	}
	var rsp cron.PriceRsp
	copier.Copy(&rsp, period)
	return &rsp, nil
}
