package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/timex"
)

var (
	_                                           WeatherModel = (*customWeatherModel)(nil)
	cacheEdsCronWeatherCityDateSizePrefix                    = "cache:edsCron:weather:city:date:size:"
	cacheEdsCronWeatherCityDateHiTempSizePrefix              = "cache:edsCron:weather:city:date:hi-temp:size:"
)

type (
	// WeatherModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWeatherModel.
	WeatherModel interface {
		weatherModel
		FindAllByDateCity(ctx context.Context, date string, city string, size int64) (*[]Weather, error)
		FindHiTempSize(ctx context.Context, date string, city string, temp float64) (int64, error)
	}

	customWeatherModel struct {
		*defaultWeatherModel
	}
)

// NewWeatherModel returns a model for the database table.
func NewWeatherModel(conn sqlx.SqlConn, c cache.CacheConf) WeatherModel {
	return &customWeatherModel{
		defaultWeatherModel: newWeatherModel(conn, c),
	}
}

func (m *customWeatherModel) FindAllByDateCity(ctx context.Context, date string, city string, size int64) (*[]Weather, error) {
	// 天气未开放人为编辑接口，加入缓存安全
	start := timex.MustTime(date)

	// `date` between "2025-08-01" and "2025-08-02" 将返回两条记录
	end := start.AddDate(0, 0, int(size-1))

	key := fmt.Sprintf(cacheEdsCronWeatherCityDateSizePrefix+"%s:%s:%d", city, start.Format(vars.DateFormat), size)
	var weas []Weather
	err := m.GetCacheCtx(ctx, key, &weas)
	if err == nil {
		return &weas, nil
	} else if err != sqlx.ErrNotFound {
		return nil, err
	}

	query := fmt.Sprintf("select %s from %s where `city` = ? and `date` between ? and ?", weatherRows, m.table)
	err = m.QueryRowsNoCacheCtx(ctx, &weas, query, city, start.Format(vars.DateFormat), end.Format(vars.DateFormat))
	if err != nil {
		return nil, err
	}

	err = m.SetCacheWithExpireCtx(ctx, key, &weas, time.Hour*24)
	if err != nil {
		return nil, err
	}

	return &weas, nil
}

func (m *customWeatherModel) FindHiTempSize(ctx context.Context, date string, city string, temp float64) (int64, error) {
	key := fmt.Sprintf(cacheEdsCronWeatherCityDateHiTempSizePrefix+"%s:%s:%f", city, date, temp)
	var size int64
	err := m.GetCacheCtx(ctx, key, &size)
	if err == nil {
		return size, nil
	}

	day := timex.MustTime(date)
	for {
		wea, err := m.FindOneByDateCity(ctx, day.Format(vars.DateFormat), city)
		if err != nil || wea == nil || (*wea).DayTemp < temp {
			break
		}

		size++
		day = day.AddDate(0, 0, -1)
	}

	err = m.SetCacheWithExpireCtx(ctx, key, &size, timex.SubTomorrow())
	if err != nil {
		return 0, err
	}

	return size, nil
}
