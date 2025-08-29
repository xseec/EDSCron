package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/slicex"
	"seeccloud.com/edscron/pkg/x/timex"
)

var (
	_                                                          DlgdModel = (*customDlgdModel)(nil)
	CategoryFormat                                                       = "%s>%s>%s"
	CategoryFormatTip                                                    = "area>category>voltage"
	cacheEdsCronDlgdAreaStartTimeCategoryVoltagePrefix                   = "cache:edsCron:dlgd:area:startTime:category:voltage:"
	cacheEdsCronDlgdAreaCategoryVoltageAtNearlyStartTimePrefix           = "cache:edsCron:dlgd:area:category:voltage:atNearlyStartTime:"
)

type (
	// DlgdModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDlgdModel.
	DlgdModel interface {
		dlgdModel
		FindCategoriesByAreas(ctx context.Context, areas ...string) (*[]string, error)
		FindAllByAreaStartTimeCategoryVoltage(ctx context.Context, area string, startTime time.Time, category string, voltage string) (*[]Dlgd, error)
		FindOneByAreaCategoryVoltageAtNearlyStartTime(ctx context.Context, area string, startTime time.Time, category string, voltage string) (*Dlgd, error)
	}

	customDlgdModel struct {
		*defaultDlgdModel
	}

	Category struct {
		Area     string `db:"area"`     // 区域
		Category string `db:"category"` // 用电分类，如单一制、两部制
		Voltage  string `db:"voltage"`  // 电压等级，如1-10（20）千伏
	}
)

// NewDlgdModel returns a model for the database table.
func NewDlgdModel(conn sqlx.SqlConn, c cache.CacheConf) DlgdModel {
	return &customDlgdModel{
		defaultDlgdModel: newDlgdModel(conn, c),
	}
}

func (m *customDlgdModel) FindCategoriesByAreas(ctx context.Context, areas ...string) (*[]string, error) {
	// 请求不频繁，无需缓存
	for _, area := range areas {
		if len(area) == 0 {
			continue
		}

		query := fmt.Sprintf("select distinct `area`, `category`, `voltage` from %s where `area` like '%%%s%%'", m.table, area)
		var values []Category
		err := m.QueryRowsNoCacheCtx(ctx, &values, query)
		if err != nil || len(values) > 0 {
			categories := slicex.MapFunc(values, func(c Category) string {
				return fmt.Sprintf(CategoryFormat, c.Area, c.Category, c.Voltage)
			})
			return &categories, err
		}
	}

	return nil, nil
}

func (m *customDlgdModel) FindAllByAreaStartTimeCategoryVoltage(ctx context.Context, area string, startTime time.Time, category string, voltage string) (*[]Dlgd, error) {
	// 存在阶梯电价时返回超过一条记录（常为2条），引入缓存机制
	key := fmt.Sprintf("%s%v:%v:%v:%v", cacheEdsCronDlgdAreaStartTimeCategoryVoltagePrefix, area, startTime, category, voltage)
	var all []Dlgd
	err := m.GetCacheCtx(ctx, key, &all)
	if err == nil {
		return &all, nil
	} else if err != sqlx.ErrNotFound {
		return nil, err
	}

	q := fmt.Sprintf("select %s from %s where `area` = ? and `start_time` = ? and `category` = ? and `voltage` = ?", dlgdRows, m.table)
	err = m.QueryRowsNoCacheCtx(ctx, &all, q, area, startTime, category, voltage)
	if err != nil {
		return nil, err
	}

	// 默认缓存7天过长，可能在此期间记录已更新
	err = m.SetCacheWithExpireCtx(ctx, key, &all, time.Hour)
	if err != nil {
		return nil, err
	}

	return &all, nil
}

func (m *customDlgdModel) FindOneByAreaCategoryVoltageAtNearlyStartTime(ctx context.Context, area string, startTime time.Time, category string, voltage string) (*Dlgd, error) {
	// 时间没法匹配时选最接近的一条
	key := fmt.Sprintf("%s%v:%v:%v:%v", cacheEdsCronDlgdAreaCategoryVoltageAtNearlyStartTimePrefix, area, startTime, category, voltage)
	var one Dlgd
	err := m.GetCacheCtx(ctx, key, &one)
	if err == nil {
		return &one, nil
	} else if err != sqlx.ErrNotFound {
		return nil, err
	}

	q := fmt.Sprintf("select %s from %s where `area` = ? and `category` = ? and `voltage` = ? order by abs(unix_timestamp(`start_time`) - unix_timestamp(?)) limit 1", dlgdRows, m.table)
	err = m.QueryRowNoCacheCtx(ctx, &one, q, area, category, voltage, startTime)
	if err != nil {
		return nil, err
	}

	// 默认缓存7天过长，可能在此期间记录已更新
	err = m.SetCacheWithExpireCtx(ctx, key, &one, time.Hour)
	if err != nil {
		return nil, err
	}

	return &one, nil
}

func (d *Dlgd) GetPrice(t time.Time, holiday cronx.HolidayCategory, isWeatherActived bool) cronx.Period {
	var period cronx.Period
	// 深谷、尖段优先级高于谷段和峰段
	if isDateActived(t, d.DeepDate, holiday, isWeatherActived) && timex.IsHourInRange(t, d.DeepHour) {
		period = cronx.PeriodDeep
		period.Value = d.Deep
		return period
	}

	if isDateActived(t, d.SharpDate, holiday, isWeatherActived) && timex.IsHourInRange(t, d.SharpHour) {
		period = cronx.PeriodSharp
		period.Value = d.Sharp
		return period
	}

	if isDateActived(t, d.ValleyDate, holiday, isWeatherActived) && timex.IsHourInRange(t, d.ValleyHour) {
		period = cronx.PeriodValley
		period.Value = d.Valley
		return period
	}

	if isDateActived(t, d.PeakDate, holiday, isWeatherActived) && timex.IsHourInRange(t, d.PeakHour) {
		period = cronx.PeriodPeak
		period.Value = d.Peak
		return period
	}

	period = cronx.PeriodFlat
	period.Value = d.Flat
	return period
}

func isDateActived(t time.Time, dateCondition string, holiday cronx.HolidayCategory, isWeatherActived bool) bool {
	// 空值表示不受限，总是生效
	if len(dateCondition) == 0 {
		return true
	}

	// 检查假期条件
	if strings.Contains(dateCondition, "holiday") && holiday == cronx.HolidayOff {
		return true
	}

	// 检查周末条件(需排除调休工作日)
	if strings.Contains(dateCondition, "weekend") {
		isWeekend := t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
		if isWeekend && holiday != cronx.HolidayOn {
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

	// 检查天气条件(温度), 格式："weather:广州>=35"
	if strings.Contains(dateCondition, "weather") && isWeatherActived {
		return true
	}

	return false
}
