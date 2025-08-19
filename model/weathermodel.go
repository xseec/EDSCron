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
	_                                       WeatherModel = (*customWeatherModel)(nil)
	cacheEdsEnergyWeatherDateCitySizePrefix              = "cache:edsEnergy:weather:date:city:size:"
)

type (
	// WeatherModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWeatherModel.
	WeatherModel interface {
		weatherModel
		FindMoreByDateCity(ctx context.Context, date string, city string, size int64) (*[]Weather, error)
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

func (m *customWeatherModel) FindMoreByDateCity(ctx context.Context, date string, city string, size int64) (*[]Weather, error) {
	// 天气未开放人为编辑接口，加入缓存安全
	start := time.Now()
	if t, err := time.Parse(vars.DateFormat, date); err == nil {
		start = t
	}

	end := start.AddDate(0, 0, int(size))

	key := fmt.Sprintf(cacheEdsEnergyWeatherDateCitySizePrefix+"%s:%s:%d", start.Format(vars.DateFormat), city, size)
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
