package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/stringx"
)

var (
	itemSplit                                         = ","
	_                                   DlgdHourModel = (*customDlgdHourModel)(nil)
	cacheEdsCronDlgdHourAreaDocNoPrefix               = "cache:edsCron:dlgdHour:area:docNo:"
)

type (
	// DlgdHourModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDlgdHourModel.
	DlgdHourModel interface {
		dlgdHourModel
		InsertAll(ctx context.Context, area string, comment string) error
		ConfirmAll(ctx context.Context, area string, docNo string) error
		QueryAll(ctx context.Context, area string, docNo string) (*[]DlgdHour, error)
	}

	customDlgdHourModel struct {
		*defaultDlgdHourModel
	}
)

// NewDlgdHourModel returns a model for the database table.
func NewDlgdHourModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) DlgdHourModel {
	return &customDlgdHourModel{
		defaultDlgdHourModel: newDlgdHourModel(conn, c, opts...),
	}
}

func (m *customDlgdHourModel) ConfirmAll(ctx context.Context, area string, docNo string) error {
	key := fmt.Sprintf("%s%s:%s", cacheEdsCronDlgdHourAreaDocNoPrefix, area, docNo)
	m.DelCacheCtx(ctx, key)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf(`update %s set confirm = 1 where area = ? and doc_no = ?`, m.table)
		return conn.ExecCtx(ctx, query, area, docNo)
	})
	return err
}

func (m *customDlgdHourModel) InsertAll(ctx context.Context, area string, comment string) error {
	period := cronx.NewPeriod(comment)
	for _, hour := range period.Hours {
		data := DlgdHour{
			Area:          area,
			DocNo:         period.DocNo,
			Name:          hour.Name,
			Value:         hour.Value,
			Temp:          hour.Temp,
			Months:        stringx.Join(hour.Months, itemSplit),
			WeekendMonths: stringx.Join(hour.WeekendMonths, itemSplit),
			Holidays:      strings.Join(hour.Holidays, itemSplit),
			Categories:    strings.Join(hour.Categories, itemSplit),
		}
		if _, err := m.Insert(ctx, &data); err != nil {
			return err
		}
	}

	return nil
}

func (m *customDlgdHourModel) QueryAll(ctx context.Context, area string, docNo string) (*[]DlgdHour, error) {
	key := fmt.Sprintf("%s%s:%s", cacheEdsCronDlgdHourAreaDocNoPrefix, area, docNo)
	hours := make([]DlgdHour, 0)
	err := m.GetCacheCtx(ctx, key, &hours)
	if err == nil {
		return &hours, nil
	} else if err != ErrNotFound {
		return nil, err
	}

	query := fmt.Sprintf(`select %s from %s where area = ? and doc_no = ?`, dlgdHourRows, m.table)
	err = m.QueryRowsNoCacheCtx(ctx, &hours, query, area, docNo)
	if err != nil {
		return nil, err
	}

	m.SetCacheCtx(ctx, key, hours)
	return &hours, nil
}
