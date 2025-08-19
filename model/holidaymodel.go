package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/vars"
)

var (
	_                                   HolidayModel = (*customHolidayModel)(nil)
	cacheEdsEnergyHolidayAreaYearPrefix              = "cache:edsEnergy:holiday:area:year:"
)

type (
	// HolidayModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHolidayModel.
	HolidayModel interface {
		holidayModel
		FindAllByAreaYear(ctx context.Context, area string, year int) (*[]Holiday, error)
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
	key := fmt.Sprintf(cacheEdsEnergyHolidayAreaYearPrefix+"%s:%d", area, year)
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

func (m *customHolidayModel) AddMany(ctx context.Context, area string, holidays *[]Holiday) error {
	years := map[int]any{}
	for _, hol := range *holidays {
		t, err := time.Parse(vars.DateFormat, hol.Date)
		if err != nil {
			return err
		}

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
		_, err = m.Insert(ctx, &hol)
		if err != nil {
			return err
		}
	}

	// 编辑后需同步删除缓存
	for y, _ := range years {
		key := fmt.Sprintf(cacheEdsEnergyHolidayAreaYearPrefix+"%s:%d", area, y)
		m.DelCacheCtx(ctx, key)
	}

	return nil
}
