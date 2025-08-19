package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/x/slicex"
)

var (
	_                                                    DlgdModel = (*customDlgdModel)(nil)
	CategoryFormat                                                 = "%s>%s>%s"
	CategoryFormatTip                                              = "area>category>voltage"
	CategorySep                                                    = ">"
	cacheEdsEnergyDlgdAreaStartTimeCategoryVoltagePrefix           = "cache:edsEnergy:dlgd:area:startTime:category:voltage:"
)

type (
	// DlgdModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDlgdModel.
	DlgdModel interface {
		dlgdModel
		FindCategoriesByAreas(ctx context.Context, areas ...string) (*[]string, error)
		FindAllByAreaStartTimeCategoryVoltage(ctx context.Context, area string, startTime int64, category string, voltage string) (*[]Dlgd, error)
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

func (m *customDlgdModel) FindAllByAreaStartTimeCategoryVoltage(ctx context.Context, area string, startTime int64, category string, voltage string) (*[]Dlgd, error) {
	// 存在阶梯电价时返回超过一条记录（常为2条），引入缓存机制
	key := fmt.Sprintf("%s%v:%v:%v:%v", cacheEdsEnergyDlgdAreaStartTimeCategoryVoltagePrefix, area, startTime, category, voltage)
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
