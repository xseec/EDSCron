package svc

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
)

// GetPrice 获取指定时间点的电价信息
func (svc *ServiceContext) GetDlgdPrice(t time.Time, one *model.Dlgd) (*cronx.Period, error) {

	ctx := context.Background()
	dateCondition := one.DeepDate + one.SharpDate + one.ValleyDate + one.PeakDate + one.FlatDate

	holiday := cronx.HolidayNull
	// holiday, holiday:3, holiday:春节,劳动节,国庆节;weekend
	subs := regexp.MustCompile("holiday([^;]*)").FindStringSubmatch(dateCondition)
	if len(subs) == 2 {
		hol, err := svc.HolidayModel.FindOneByAreaDateCache(ctx, string(cronx.ChinaArea), t.Format(vars.DateFormat))
		if err == nil {
			// 调休工作日
			if hol.Category == string(cronx.HolidayOn) {
				holiday = cronx.HolidayOn
			}

			if hol.Category == string(cronx.HolidayOff) {
				// 全年或指定假期
				if len(subs[1]) == 0 || strings.Contains(subs[1], hol.Detail) {
					holiday = cronx.HolidayOff
				}

				// 连续N天以上的假期
				if num, err := strconv.Atoi(subs[1][len(subs[1])-1:]); err == nil {
					size, _ := svc.HolidayModel.GetHolidayOffSizeByAreaDate(ctx, string(cronx.ChinaArea), t.Format(vars.DateFormat))
					if int(size) >= num {
						holiday = cronx.HolidayOff
					}
				}
			}
		}
	}

	isWeatherActived := false
	subs = regexp.MustCompile(`temp:(\d+)(?:,(\d+))?`).FindStringSubmatch(dateCondition)
	if len(subs) == 3 {
		capital, err := svc.AreaModel.GetProvincialCapital(ctx, one.Area)
		if err != nil {
			return nil, err
		}

		temp, _ := strconv.ParseFloat(subs[1], 64)
		hiTempSize, err := svc.WeatherModel.FindHiTempSize(ctx, t.Format(vars.DateFormat), capital, temp)
		if err != nil {
			return nil, err
		}

		daySize := 1
		if len(subs[2]) > 0 {
			daySize, _ = strconv.Atoi(subs[2])
		}

		if int(hiTempSize) >= daySize {
			isWeatherActived = true
		}
	}

	price := one.GetPrice(t, holiday, isWeatherActived)
	return &price, nil
}
