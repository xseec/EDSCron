package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/cronx"
)

var (
	DlgdHourConfirmCode                 int64         = 1
	_                                   DlgdHourModel = (*customDlgdHourModel)(nil)
	cacheEdsCronDlgdHourAreaDocNoPrefix               = "cache:edsCron:dlgdHour:area:docNo:"
)

type (
	// DlgdHourModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDlgdHourModel.
	DlgdHourModel interface {
		dlgdHourModel
		InsertAll(ctx context.Context, area string, comment string) error
		MustInertAll(ctx context.Context, hours *[]DlgdHour) error
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
		query := fmt.Sprintf(`update %s set confirm = %d where area = ? and doc_no = ?`, m.table, DlgdHourConfirmCode)
		return conn.ExecCtx(ctx, query, area, docNo)
	})
	return err
}

func (m *customDlgdHourModel) InsertAll(ctx context.Context, area string, comment string) error {
	hours := cronx.NewDlgdHours(area, comment)
	for _, hour := range hours {
		data := DlgdHour{}
		copierx.MustCopy(&data, hour)
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

func (m *customDlgdHourModel) MustInertAll(ctx context.Context, hours *[]DlgdHour) error {
	if hours == nil || len(*hours) == 0 {
		return nil
	}

	area := (*hours)[0].Area
	docNo := (*hours)[0].DocNo
	key := fmt.Sprintf("%s%s:%s", cacheEdsCronDlgdHourAreaDocNoPrefix, area, docNo)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `area` = ? and `doc_no` = ?", m.table)
		return conn.ExecCtx(ctx, query, area, docNo)
	}, key)
	if err != nil {
		return err
	}

	for _, hour := range *hours {
		if _, err := m.Insert(ctx, &hour); err != nil {
			return err
		}
	}
	return nil
}

func FormatHtmlDlgdHours(hours *[]DlgdHour) string {
	if hours == nil || len(*hours) == 0 {
		return "记录为空"
	}

	var builder strings.Builder
	builder.WriteString("<ul>")
	for _, hour := range *hours {

		builder.WriteString(fmt.Sprintf("<li><p><em>%s</em> : %s</p>", hour.Name, hour.Value))

		if len(hour.Temp) > 0 {
			builder.WriteString(fmt.Sprintf("<p><em>最高气温</em> : %s</p>", hour.Temp))
		}

		if len(hour.Months) > 0 {
			builder.WriteString(fmt.Sprintf("<p><em>生效月份</em> : %s</p>", hour.Months))
		}

		if len(hour.WeekendMonths) > 0 {
			builder.WriteString(fmt.Sprintf("<p><em>当月假日</em> : %s</p>", hour.WeekendMonths))
		}

		if len(hour.Holidays) > 0 {
			builder.WriteString(fmt.Sprintf("<p><em>法定假日</em> : %s</p>", hour.Holidays))
		}

		if len(hour.Categories) > 0 {
			builder.WriteString(fmt.Sprintf("<p><em>用电类别</em> : %s</p>", hour.Categories))
		}

		builder.WriteString("</li>")
	}

	builder.WriteString("</ul>")
	return builder.String()
}
