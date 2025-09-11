package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	_ AreaModel = (*customAreaModel)(nil)

	cacheEdsCronAreaParentPrefix = "cache:edsCron:area:parent:"
)

type (
	// AreaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAreaModel.
	AreaModel interface {
		areaModel
		FindParent(ctx context.Context, area string) (*Area, error)
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

func (m *customAreaModel) FindParent(ctx context.Context, area string) (*Area, error) {
	edsCronAreaAddressKey := fmt.Sprintf("%s%v", cacheEdsCronAreaParentPrefix, area)
	var resp, parent Area
	err := m.GetCacheCtx(ctx, edsCronAreaAddressKey, &parent)
	if err == nil {
		if len(parent.Name) == 0 {
			return nil, ErrNotFound
		}
		return &parent, nil
	}

	err = m.QueryRowCtx(ctx, &resp, edsCronAreaAddressKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `name` like ? limit 1", areaRows, m.table)
		err := conn.QueryRowCtx(ctx, &resp, query, "%"+area+"%")
		if err != nil {
			return err
		}

		if resp.Parent == -1 {
			return ErrNotFound
		}

		one, err := m.FindOne(ctx, uint64(resp.Parent))
		if err != nil {
			return err
		}

		parent = *one
		return nil
	})

	switch err {
	case nil:
		m.SetCacheCtx(ctx, edsCronAreaAddressKey, parent)
		return &parent, nil
	case sqlc.ErrNotFound:
		m.SetCacheCtx(ctx, edsCronAreaAddressKey, nil)
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
