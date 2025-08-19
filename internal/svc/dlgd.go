package svc

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"
	"seeccloud.com/edscron/pkg/x/slicex"
	"seeccloud.com/edscron/pkg/x/stringx"
)

// GetPrice 获取指定时间点的电价信息
// 峰谷优先次序：深谷 > 尖峰 > 峰时 = 谷时 > 平时
func (svc *ServiceContext) GetPrice(t time.Time, dlgd *model.Dlgd) *cron.PriceRsp {
	// 定义电价时段类型及其优先级
	periodTypes := []struct {
		name  string
		date  string
		hour  string
		price float64
	}{
		{"deep", dlgd.DeepDate, dlgd.DeepHour, dlgd.Deep},         // 深谷
		{"sharp", dlgd.SharpDate, dlgd.SharpHour, dlgd.Sharp},     // 尖峰
		{"peak", dlgd.PeakDate, dlgd.PeakHour, dlgd.Peak},         // 峰时
		{"valley", dlgd.ValleyDate, dlgd.ValleyHour, dlgd.Valley}, // 谷时
		{"flat", dlgd.FlatDate, dlgd.FlatHour, dlgd.Flat},         // 平时
	}

	area := cronx.EitherChinaOrTaiwan(dlgd.Area)
	dailyPrices := make(map[int]*cron.PriceRsp) // 存储每天48个半小时点的电价(24小时*2)

	// 按优先级处理各时段电价
	for _, pt := range periodTypes {
		// 跳过无效电价(价格为0表示无此时段)
		if pt.price <= 0 {
			continue
		}

		// 检查当前日期是否适用此时段
		if svc.isDateEnabled(t, pt.date, area) {
			// 获取此时段包含的半小时点索引
			halfHourIndexes := cronx.GetHalfHourIndexs(pt.hour)

			// 为每个半小时点设置电价(高优先级时段会覆盖低优先级时段)
			slicex.EachFunc(halfHourIndexes, func(index int) {
				if _, exists := dailyPrices[index]; !exists {
					dailyPrices[index] = &cron.PriceRsp{
						Value:  pt.price,
						Period: pt.name,
					}
				}
			})
		}
	}

	// 计算当前时间对应的半小时点索引(0-47)
	currentIndex := t.Hour()*2 + t.Minute()/30
	return dailyPrices[currentIndex]
}

// isDateEnabled 检查指定日期是否符合条件
func (svc *ServiceContext) isDateEnabled(t time.Time, dateCondition string, area string) bool {
	// 空条件表示总是生效
	if len(dateCondition) == 0 {
		return true
	}

	ctx := context.Background()
	var holidayInfo *model.Holiday

	// 如果条件包含假日或周末信息，先查询假日数据
	if stringx.ContainsAny(dateCondition, "holiday", "weekend") {
		dateStr := t.Format(vars.DateFormat)
		holidayInfo, _ = svc.HolidayModel.FindOneByAreaDate(ctx, area, dateStr)
	}

	// 检查假日条件
	if holidayMatch := regexp.MustCompile("holiday([^,]*)").FindStringSubmatch(dateCondition); len(holidayMatch) == 2 {
		if holidayInfo != nil && holidayInfo.Category == cronx.HolidayText {
			// 无具体假日要求或匹配具体假日
			if len(holidayMatch[1]) == 0 || strings.Contains(holidayMatch[1], holidayInfo.Detail) {
				return true
			}
		}
	}

	// 检查周末条件(需排除调休工作日)
	if strings.Contains(dateCondition, "weekend") {
		isWeekend := t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
		isWorkday := holidayInfo != nil && holidayInfo.Category == cronx.WeekendWorkdayText
		if isWeekend && !isWorkday {
			return true
		}
	}

	// 检查周六条件
	if strings.Contains(dateCondition, "sat") && t.Weekday() == time.Saturday {
		return true
	}

	// 检查周日条件
	if strings.Contains(dateCondition, "sun") && t.Weekday() == time.Sunday {
		return true
	}

	// 检查天气条件(温度)
	if weatherMatch := regexp.MustCompile(`weather:(([^>≥<≤=]+)[>≥<≤=\d]+)`).FindStringSubmatch(dateCondition); len(weatherMatch) == 3 {
		city := weatherMatch[2]
		weather, err := svc.WeatherModel.FindOneByDateCity(ctx, t.Format(vars.DateFormat), city)
		if err != nil {
			return false
		}

		// 创建条件表达式
		expr, err := govaluate.NewEvaluableExpression(weatherMatch[1])
		if err != nil {
			return false
		}

		// 根据条件选择白天或夜间温度
		temp := expx.If(stringx.ContainsAny(weatherMatch[2], ">", "≥"), weather.DayTemp, weather.NightTemp)

		// 评估温度条件
		result, _ := expr.Evaluate(map[string]any{weatherMatch[2]: temp})
		if isMatch, ok := result.(bool); ok {
			return isMatch
		}
	}

	return false
}
