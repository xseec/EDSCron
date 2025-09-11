package svc

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
)

// GetPrice 获取指定时间点的电价信息
func (svc *ServiceContext) GetDlgdPrice(t time.Time, one *model.Dlgd) (*cronx.Period, error) {

	ctx := context.Background()
	dateCondition := one.DeepDate + one.SharpDate + one.ValleyDate + one.PeakDate + one.FlatDate

	holiday := cronx.HolidayNull
	subs := regexp.MustCompile("holiday([^,]*)").FindStringSubmatch(dateCondition)
	if len(subs) == 2 {
		hol, err := svc.HolidayModel.FindOneByAreaDateCache(ctx, string(cronx.ChinaArea), t.Format(vars.DateFormat))
		if err == nil {
			if hol.Category == string(cronx.HolidayOn) {
				holiday = cronx.HolidayOn
			}

			if hol.Category == string(cronx.HolidayOff) {
				if len(subs[1]) == 0 || strings.Contains(subs[1], hol.Detail) {
					holiday = cronx.HolidayOff
				}
			}
		}
	}

	isWeatherActived := false
	subs = regexp.MustCompile(`weather:(([^>≥<≤=]+)[>≥<≤=\d]+)`).FindStringSubmatch(dateCondition)
	if len(subs) == 3 {
		wea, err := svc.WeatherModel.FindOneByDateCity(ctx, t.Format(vars.DateFormat), subs[2])
		if err == nil {

			// 创建条件表达式
			expr, err := govaluate.NewEvaluableExpression(subs[1])
			if err != nil {
				return nil, err
			}

			// 评估温度条件
			result, _ := expr.Evaluate(map[string]any{subs[2]: wea.DayTemp})
			if isMatch, ok := result.(bool); ok {
				isWeatherActived = isMatch
			}
		}
	}

	price := one.GetPrice(t, holiday, isWeatherActived)
	return &price, nil
}
