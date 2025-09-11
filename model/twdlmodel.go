package model

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/timex"
)

var (
	_                                          TwdlModel = (*customTwdlModel)(nil)
	cacheEdsCronTwdlDayStartTimeCategoryPrefix           = "cache:edsCron:twdl:dayStartTime:category:"
)

type (
	// TwdlModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTwdlModel.
	TwdlModel interface {
		twdlModel
		FindOneByDayStartTimeCategory(ctx context.Context, time string, category string) (*Twdl, error)
		FindCategories(ctx context.Context) (*[]string, error)
	}

	customTwdlModel struct {
		*defaultTwdlModel
	}
)

// NewTwdlModel returns a model for the database table.
func NewTwdlModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TwdlModel {
	return &customTwdlModel{
		defaultTwdlModel: newTwdlModel(conn, c, opts...),
	}
}

func (m *customTwdlModel) FindCategories(ctx context.Context) (*[]string, error) {
	var values []string
	query := fmt.Sprintf("select distinct `category` from %s", m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &values, query)
	if err != nil {
		return nil, err
	}
	return &values, nil
}

func (m *customTwdlModel) FindOneByDayStartTimeCategory(ctx context.Context, dayStart string, category string) (*Twdl, error) {
	edsCronTwdlTimeCategoryKey := fmt.Sprintf("%s%v:%v", cacheEdsCronTwdlDayStartTimeCategoryPrefix, dayStart, category)
	var value Twdl
	err := m.GetCacheCtx(ctx, edsCronTwdlTimeCategoryKey, &value)
	if err == nil {
		return &value, nil
	} else if err != ErrNotFound {
		return nil, err
	}

	var values []Twdl
	// 基于start_time和category查询出夏月和非夏月的两条记录
	query := fmt.Sprintf("select * from %s where unix_timestamp(`start_time`) <= unix_timestamp(?) and `category` = ? order by abs(unix_timestamp(`start_time`) - unix_timestamp(?)) limit 2", m.table)
	err = m.QueryRowsNoCacheCtx(ctx, &values, query, dayStart, category, dayStart)
	if err != nil {
		return nil, err
	}

	t, err := time.ParseInLocation(vars.DatetimeFormat, dayStart, time.Local)
	if err != nil {
		return nil, err
	}

	// 基于日期选择夏月或非夏月记录
	for _, v := range values {
		if timex.IsDateInRange(t, v.Date) {
			m.SetCacheCtx(ctx, edsCronTwdlTimeCategoryKey, v)
			return &v, nil
		}
	}

	return nil, ErrNotFound
}

func (d *Twdl) GetPrice(now string, isOffPeakDay bool) cronx.Period {
	var period cronx.Period
	t, err := time.ParseInLocation(vars.DatetimeFormat, now, time.Local)
	if err != nil {
		return period
	}

	// 周日或离峰日
	if t.Weekday() == time.Sunday || isOffPeakDay {
		if d.SunOffPeak > 0 {
			period = cronx.PeriodSunOffPeak
			period.Price = d.SunOffPeak
			return period
		}
	}

	// 周六
	if t.Weekday() == time.Saturday {
		if timex.IsHourInRange(t, d.SatOffPeakHour) {
			if d.SatOffPeak > 0 {
				period = cronx.PeriodSatOffPeak
				period.Price = d.SatOffPeak
				return period
			}
		} else if timex.IsHourInRange(t, d.SatSemiPeakHour) {
			if d.SatSemiPeak > 0 {
				period = cronx.PeriodSatSemiPeak
				period.Price = d.SatSemiPeak
				return period
			}
		}
	}

	// 周一至周五
	if t.Weekday() >= time.Monday && t.Weekday() <= time.Friday {
		if timex.IsHourInRange(t, d.WeekdayOffPeakHour) {
			if d.WeekdayOffPeak > 0 {
				period = cronx.PeriodWeekdayOffPeak
				period.Price = d.WeekdayOffPeak
				return period
			}
		} else if timex.IsHourInRange(t, d.WeekdaySemiPeakHour) {
			if d.WeekdaySemiPeak > 0 {
				period = cronx.PeriodWeekdaySemiPeak
				period.Price = d.WeekdaySemiPeak
				return period
			}
		} else if timex.IsHourInRange(t, d.WeekdayPeakHour) {
			if d.WeekdayPeak > 0 {
				period = cronx.PeriodWeekdayPeak
				period.Price = d.WeekdayPeak
				return period
			}
		}
	}

	// 基准电价
	period = cronx.PeriodStandard
	if d.Standard > 0 {
		period.Price = d.Standard
		return period
	}

	// 阶梯电价之首阶
	subs := regexp.MustCompile(`:(\d+(?:.\d+)?)`).FindStringSubmatch(d.Stage)
	if len(subs) == 2 {
		value, _ := strconv.ParseFloat(subs[1], 64)
		if value > 0 {
			period.Price = value
			return period
		}
	}

	return period
}
