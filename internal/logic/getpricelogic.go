package logic

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"
	"seeccloud.com/edscron/pkg/x/slicex"
	"seeccloud.com/edscron/pkg/x/stringx"

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

	// 代理购电的category以省市开头
	area, _ := l.svcCtx.AreaModel.FindAddress(l.ctx, regexp.MustCompile(`\p{Han}+`).FindString(in.Category))
	if len(area) == 0 {
		return GetTwdlPrice(in.Category, t, l)
	}

	return GetDlgdPrice(in.Category, t, l, in)
}

func GetTwdlPrice(category string, t time.Time, l *GetPriceLogic) (*cron.PriceRsp, error) {

	dayStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	one, err := l.svcCtx.TwdlModel.FindOneByDayStartTimeCategory(l.ctx, dayStart, category)
	if err != nil {
		return nil, err
	}

	holiday, _ := l.svcCtx.HolidayModel.FindOneByAreaDateCache(l.ctx, string(cronx.TaiwanArea), t.Format(vars.DateFormat))

	price := one.GetPrice(t, holiday != nil && holiday.Category == string(cronx.HolidayPeakOff))
	var rsp cron.PriceRsp
	copier.Copy(&rsp, price)
	return &rsp, nil
}

func GetDlgdPrice(category string, t time.Time, l *GetPriceLogic, in *cron.PriceReq) (*cron.PriceRsp, error) {
	infos := strings.Split(category, cronx.CategorySep)
	if len(infos) < 3 {
		return nil, fmt.Errorf("req.Category格式错误, 正确格式: %s", model.CategoryFormatTip)
	}

	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	all, err := l.svcCtx.DlgdModel.FindAllByAreaStartTimeCategoryVoltage(l.ctx, infos[0], start, infos[1], infos[2])
	if err != nil {
		return nil, err
	}

	var one model.Dlgd
	if len(*all) == 0 {
		nearlyOne, err := l.svcCtx.DlgdModel.FindOneByAreaCategoryVoltageAtNearlyStartTime(l.ctx, infos[0], start, infos[1], infos[2])
		if err != nil || nearlyOne == nil {
			return nil, fmt.Errorf("未找到%s>%d年%d月>%s>%s电价表", infos[0], t.Year(), t.Month(), infos[1], infos[2])
		}

		one = *nearlyOne
		one.StartTime = start
		one.EndTime = start.AddDate(0, 1, 0)
	} else {
		one = (*all)[0]
		// 阶梯电价取最低档位那条（深圳）
		if len(*all) > 1 {
			one = slicex.FirstOrDefFunc(*all, one, func(o model.Dlgd) bool {
				return stringx.ContainsAny(o.Stage, "以下", "<", "<=", "≤")
			})
		}
	}

	return l.svcCtx.GetDlgdPrice(t, &one)
}
