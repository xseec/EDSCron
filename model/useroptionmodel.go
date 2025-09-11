package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	_ UserOptionModel = (*customUserOptionModel)(nil)

	cacheEdsCronUserOptionAccountNearlyAreaPrefix = "cache:edsCron:userOption:account:nearlyArea:"
)

type (
	// UserOptionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserOptionModel.
	UserOptionModel interface {
		userOptionModel
		FindOneByAccountNearlyArea(ctx context.Context, account string, area string) (*UserOption, error)
	}

	customUserOptionModel struct {
		*defaultUserOptionModel
	}
)

// NewUserOptionModel returns a model for the database table.
func NewUserOptionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserOptionModel {
	return &customUserOptionModel{
		defaultUserOptionModel: newUserOptionModel(conn, c, opts...),
	}
}

func (m *customUserOptionModel) FindOneByAccountNearlyArea(ctx context.Context, account string, area string) (*UserOption, error) {
	key := fmt.Sprintf("%s%v:%v", cacheEdsCronUserOptionAccountNearlyAreaPrefix, account, area)
	var resp UserOption
	err := m.QueryRowIndexCtx(ctx, &resp, key, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (any, error) {
		query := fmt.Sprintf("SELECT %s FROM %s WHERE `account` = ? AND `area` LIKE ? limit 1", userOptionRows, m.table)
		err := conn.QueryRowCtx(ctx, &resp, query, account, "%"+area+"%")
		if err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
