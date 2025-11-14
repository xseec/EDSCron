package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/timex"
)

var (
	_                                        HolidayModel = (*customHolidayModel)(nil)
	cacheEdsCronHolidayAreaYearPrefix                     = "cache:edsCron:holiday:area:year:"
	cacheEdsCronHolidayOffSizeAreaDatePrefix              = "cache:edsCron:holidayOffSize:area:date:"
)

type (
	// HolidayModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHolidayModel.
	HolidayModel interface {
		holidayModel
		FindAllByAreaYear(ctx context.Context, area string, year int) (*[]Holiday, error)
		FindOneByAreaDateCache(ctx context.Context, area string, date string) (*Holiday, error)
		GetHolidayOffSizeByAreaDate(ctx context.Context, area string, date string) (int64, error)
		AddMany(ctx context.Context, area string, holidays *[]Holiday) error
	}

	customHolidayModel struct {
		*defaultHolidayModel
	}
)

// NewHolidayModel returns a model for the database table.
func NewHolidayModel(conn sqlx.SqlConn, c cache.CacheConf) HolidayModel {
	return &customHolidayModel{
		defaultHolidayModel: newHolidayModel(conn, c),
	}
}

func (m *customHolidayModel) FindAllByAreaYear(ctx context.Context, area string, year int) (*[]Holiday, error) {
	// 假日编辑场景较少，加入缓存安全，编辑后需同步删除缓存
	key := fmt.Sprintf(cacheEdsCronHolidayAreaYearPrefix+"%s:%d", area, year)
	var days []Holiday
	err := m.GetCacheCtx(ctx, key, &days)
	if err == nil {
		return &days, nil
	} else if err != ErrNotFound {
		return nil, err
	}

	query := fmt.Sprintf("select %s from %s where `area` = ? and `date` like '%d%%'", holidayRows, m.table, year)
	err = m.QueryRowsNoCacheCtx(ctx, &days, query, area)
	if err != nil {
		return nil, err
	}

	err = m.SetCacheCtx(ctx, key, &days)
	if err != nil {
		return nil, err
	}

	return &days, nil
}

func (m *customHolidayModel) FindOneByAreaDateCache(ctx context.Context, area string, date string) (*Holiday, error) {
	// 查询日期假期属高频场景，空纪录也加入缓存1天
	one, err := m.FindOneByAreaDate(ctx, area, date)
	if err == ErrNotFound {
		key := fmt.Sprintf(cacheEdsCronHolidayAreaDatePrefix+"%s:%s", area, date)
		m.SetCacheWithExpireCtx(ctx, key, nil, time.Hour*24)
	}

	return one, err
}

func (m *customHolidayModel) GetHolidayOffSizeByAreaDate(ctx context.Context, area string, date string) (int64, error) {
	key := fmt.Sprintf(cacheEdsCronHolidayOffSizeAreaDatePrefix+"%s:%s", area, date)

	var size int64
	err := m.GetCacheCtx(ctx, key, &size)
	if err == nil {
		return size, nil
	}

	one, err := m.FindOneByAreaDateCache(ctx, area, date)
	if err != nil {
		return 0, err
	}

	if one.Category != string(cronx.HolidayOff) {
		return 0, nil
	}

	size = 1
	// 解析基础日期
	baseDay := timex.MustDate(date)

	// 向前扩展（过去的日期）
	current := baseDay.AddDate(0, 0, -1) // 从指定日期前一天开始
	for {
		currentStr := current.Format(vars.DateFormat)
		prev, err := m.FindOneByAreaDateCache(ctx, area, currentStr)
		if err != nil || prev.Category != string(cronx.HolidayOff) {
			break // 遇到非假期则停止向前扩展
		}
		// 向前扩展的日期需要插入到列表头部以保持顺序
		size++
		current = current.AddDate(0, 0, -1)
	}

	// 向后扩展（未来的日期）
	current = baseDay.AddDate(0, 0, 1) // 从指定日期后一天开始
	for {
		currentStr := current.Format(vars.DateFormat)
		next, err := m.FindOneByAreaDateCache(ctx, area, currentStr)
		if err != nil || next.Category != string(cronx.HolidayOff) {
			break // 遇到非假期则停止向后扩展
		}
		size++
		current = current.AddDate(0, 0, 1)
	}

	err = m.SetCacheCtx(ctx, key, &size)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (m *customHolidayModel) AddMany(ctx context.Context, area string, holidays *[]Holiday) error {
	years := map[int]any{}
	for _, hol := range *holidays {
		t := timex.MustTime(hol.Date)

		years[t.Year()] = nil

		if one, err := m.FindOneByAreaDate(ctx, area, hol.Date); err == nil {
			one.Category = hol.Category
			one.Detail = hol.Detail
			err = m.Update(ctx, one)
			if err != nil {
				return err
			}

			continue
		}

		hol.Area = area
		if _, err := m.Insert(ctx, &hol); err != nil {
			return err
		}
	}

	// 编辑后需同步删除缓存
	for y := range years {
		key := fmt.Sprintf(cacheEdsCronHolidayAreaYearPrefix+"%s:%d", area, y)
		m.DelCacheCtx(ctx, key)
	}

	return nil
}
