package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	_ AreaModel = (*customAreaModel)(nil)

	cacheEdsCronAreaAddressPrefix = "cache:edsCron:area:address:"
)

type (
	// AreaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAreaModel.
	AreaModel interface {
		areaModel
		FindAddress(ctx context.Context, area string) (string, error)
	}

	customAreaModel struct {
		*defaultAreaModel
	}
)

// NewAreaModel returns a model for the database table.
func NewAreaModel(conn sqlx.SqlConn, c cache.CacheConf) AreaModel {
	return &customAreaModel{
		defaultAreaModel: newAreaModel(conn, c),
	}
}

func (m *customAreaModel) FindAddress(ctx context.Context, area string) (string, error) {
	edsCronAreaAddressKey := fmt.Sprintf("%s%v", cacheEdsCronAreaAddressPrefix, area)
	var address string
	err := m.GetCacheCtx(ctx, edsCronAreaAddressKey, &address)
	if err == nil {
		return address, nil
	} else if err != ErrNotFound {
		return "", err
	}

	var resp Area
	query := fmt.Sprintf("select %s from %s where `name` like ? limit 1", areaRows, m.table)
	err = m.QueryRowNoCacheCtx(ctx, &resp, query, "%"+area+"%")
	if err == ErrNotFound {
		m.SetCacheCtx(ctx, edsCronAreaAddressKey, nil)
	}

	if err != nil {
		return "", err
	}

	address = resp.Name
	// 避免数据库数据错误进入死循环(id=100, parent=100)
	for range 5 {
		if resp.Parent == 0 {
			break
		}

		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", areaRows, m.table)
		err := m.QueryRowNoCacheCtx(ctx, &resp, query, resp.Parent)
		if err != nil {
			return "", err
		}

		address = resp.Name + address
	}

	m.SetCacheCtx(ctx, edsCronAreaAddressKey, address)
	return address, nil
}
