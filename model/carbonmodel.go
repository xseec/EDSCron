package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	_                              CarbonModel = (*customCarbonModel)(nil)
	cacheEdsEnergyCarbonAreaPrefix             = "cache:edsEnergy:carbon:area:"
)

type (
	// CarbonModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCarbonModel.
	CarbonModel interface {
		carbonModel
		FindOneByArea(ctx context.Context, area string) (*Carbon, error)
	}

	customCarbonModel struct {
		*defaultCarbonModel
	}
)

// NewCarbonModel returns a model for the database table.
func NewCarbonModel(conn sqlx.SqlConn, c cache.CacheConf) CarbonModel {
	return &customCarbonModel{
		defaultCarbonModel: newCarbonModel(conn, c),
	}
}

func (m *customCarbonModel) FindOneByArea(ctx context.Context, area string) (*Carbon, error) {
	key := fmt.Sprintf("%s%s", cacheEdsEnergyCarbonAreaPrefix, area)
	var c Carbon
	err := m.QueryRowCtx(ctx, &c, key, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `area` = ? order by `year` desc limit 1", carbonRows, m.table)
		return conn.QueryRowCtx(ctx, v, query, area)
	})

	return &c, err
}
