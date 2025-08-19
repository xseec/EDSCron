package model

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/x/slicex"
)

var (
	_        CronModel = (*customCronModel)(nil)
	reDlgd             = "re-dlgd"
	twCarbon           = "tw-carbon"
)

type (
	// CronModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCronModel.
	CronModel interface {
		cronModel
		FindAll(ctx context.Context) (*[]Cron, error)
		InsertOrIgnore(ctx context.Context, data *Cron) (sql.Result, error)
	}

	customCronModel struct {
		*defaultCronModel
	}
)

// NewCronModel returns a model for the database table.
func NewCronModel(conn sqlx.SqlConn, c cache.CacheConf) CronModel {
	return &customCronModel{
		defaultCronModel: newCronModel(conn, c),
	}
}

func (m *customCronModel) FindAll(ctx context.Context) (*[]Cron, error) {
	crons := make([]Cron, 0)
	query := fmt.Sprintf("select %s from %s", cronRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &crons, query)

	switch err {
	case nil:
		return &crons, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customCronModel) InsertOrIgnore(ctx context.Context, data *Cron) (sql.Result, error) {
	if data.Category != reDlgd {
		return m.Insert(ctx, data)
	}

	crons, err := m.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if slicex.Any(*crons, func(c Cron) bool { return c.Category == reDlgd }) {
		return nil, fmt.Errorf("%s 已存在", reDlgd)
	}

	return m.Insert(ctx, data)
}

// 获取下期任务，替换cron.task中的时间参数cron.time为下期时间
func (c *Cron) Deltas() ([]int, error) {
	nums := regexp.MustCompile(`\d+`).FindAllString(c.DeltaTime, -1)
	if len(nums) != 6 {
		return nil, fmt.Errorf("delta_time(%s) 格式错误", c.DeltaTime)
	}

	times := slicex.MapFunc(nums, func(s string) int {
		i, _ := strconv.Atoi(s)
		return i
	})

	return times, nil
}

func (c *Cron) NextStartTime() error {
	times, err := c.Deltas()
	if err != nil {
		return err
	}

	c.StartTime = time.Unix(c.StartTime, 0).AddDate(times[0], times[1], times[2]).Add(time.Duration(times[3]) * time.Hour).Add(time.Duration(times[4]) * time.Minute).Add(time.Duration(times[5]) * time.Second).Unix()

	// 预防冷却太久，下期起始时间小于当前时间
	if c.StartTime < time.Now().Unix() {
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		c.StartTime = startOfDay.AddDate(times[0], times[1], times[2]).Add(time.Duration(times[3]) * time.Hour).Add(time.Duration(times[4]) * time.Minute).Add(time.Duration(times[5]) * time.Second).Unix()
	}
	return nil
}

func (c *Cron) NextTask() ([]byte, error) {
	times, err := c.Deltas()
	if err != nil {
		return nil, err
	}

	var t time.Time
	if c.Category == twCarbon {
		// (台湾)碳排查的是去年的数据
		t = time.Now().AddDate(-1, 0, 0)
	} else {
		t = time.Now().AddDate(times[0], times[1], times[2]).Add(time.Duration(times[3]) * time.Hour).Add(time.Duration(times[4]) * time.Minute).Add(time.Duration(times[5]) * time.Second)
	}

	return []byte(strings.ReplaceAll(c.Task, c.Time, t.Format(c.Time))), nil
}
