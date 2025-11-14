package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/slicex"
	"seeccloud.com/edscron/pkg/x/timex"
)

// Category 定义任务类型枚举
type CronCategory string

var (
	_                CronModel    = (*customCronModel)(nil)
	twCarbon                      = "tw-carbon"
	CategoryDlgd     CronCategory = "dlgd"      // 代理购电
	CategoryHoliday  CronCategory = "holiday"   // 节假日
	CategoryWeather  CronCategory = "weather"   // 天气
	CategoryCarbon   CronCategory = "carbon"    // 大陆碳排因子
	CategoryTwCarbon CronCategory = "tw-carbon" // 台湾碳排因子
	CategoryTwdl     CronCategory = "twdl"      // 台湾电价
	CategoryReDlgd   CronCategory = "re-dlgd"   // 重试代理购电
)

type (
	// CronModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCronModel.
	CronModel interface {
		cronModel
		FindAll(ctx context.Context) (*[]Cron, error)
		BatchInsert(ctx context.Context, data []Cron) (sql.Result, error)
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

func (m *customCronModel) BatchInsert(ctx context.Context, data []Cron) (sql.Result, error) {
	crons, err := m.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var lastResult sql.Result
	for _, v := range data {
		if slices.ContainsFunc(*crons, func(c Cron) bool { return c.Category == v.Category && c.Task == v.Task }) {
			continue
		}

		lastResult, err = m.Insert(ctx, &v)
		if err != nil {
			return nil, err
		}
	}

	return lastResult, nil
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

	c.StartTime = timex.Add(c.StartTime, times[0], times[1], times[2], times[3], times[4], times[5])

	// 预防冷却太久，下期起始时间小于当前时间
	if c.StartTime.Before(time.Now()) {
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		c.StartTime = timex.Add(startOfDay, times[0], times[1], times[2], times[3], times[4], times[5])
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
		t = timex.Add(time.Now(), times[0], times[1], times[2], times[3], times[4], times[5])
	}

	return []byte(strings.ReplaceAll(c.Task, c.Time, t.Format(c.Time))), nil
}

func NewCron(category CronCategory, address Address) Cron {
	cron := Cron{
		Category:  string(category),
		StartTime: time.Now(),
	}

	s := rand.Intn(60) // 随机秒
	m := rand.Intn(60) // 随机分钟
	h := rand.Intn(24) // 随机小时
	// scheduler 格式：秒 分 时 日 月 周
	// delta_time 格式：年,月,日,时,分,秒
	switch category {
	case CategoryDlgd:
		// 每月末(27-31号), 每日多次执行
		h = rand.Intn(12)
		cron.Scheduler = fmt.Sprintf("%d %d %d-23/6 27-31 * *", s, m, h)
		// 执行成功后，下月再执行
		cron.DeltaTime = "0,1,0,0,0,0"
		cron.Time = "2006年1月"
		minDlgd := cronx.MiniDlgdConfig{
			Province: address.Province,
			City:     address.City,
			Area:     address.Area,
			Month:    cron.Time,
		}
		task, _ := json.Marshal(minDlgd)
		cron.Task = string(task)
	case CategoryHoliday:
		// 每年12月，每日随机执行
		cron.Scheduler = fmt.Sprintf("%d %d %d 1-31 12 *", s, m, h)
		// 限定12月, +1月即可避免重复执行
		cron.DeltaTime = "0,1,0,0,0,0"
		cron.Time = "2006"
		cron.Task = cronx.DefaultHolidayGovTask()
	case CategoryWeather:
		// 每日19:30后执行，每10分钟一次
		m = 30 + rand.Intn(10)
		cron.Scheduler = fmt.Sprintf("%d %d-59/10 19-23 * * *", s, m)
		// 执行成功后，明天再执行
		cron.DeltaTime = "0,0,1,0,0,0"
		task, _ := json.Marshal(address)
		cron.Task = string(task)
	case CategoryCarbon:
		// 每月1/11/21号任意时间执行
		cron.Scheduler = fmt.Sprintf("%d %d %d 1-21/10 * *", s, m, h)
		// 固定且低频，每月固定执行
		cron.DeltaTime = "0,0,0,0,0,0"
		cron.Time = "2006"
		cron.Task = cronx.DefaultCarbonGovTask()
	case CategoryTwCarbon:
		// 每月1~26号8点执行
		cron.Scheduler = fmt.Sprintf("%d %d 8 1-26 * *", s, m)
		// 固定且低频，每月固定执行
		cron.DeltaTime = "0,0,0,0,0,0"
		cron.Time = "2006"
		cron.Task = cronx.DefaultTwCarbonTask()
	case CategoryTwdl:
		// 每月1~26号7点执行
		cron.Scheduler = fmt.Sprintf("%d %d 7 1-26 * *", s, m)
		// 固定且低频，每月固定执行
		cron.DeltaTime = "0,0,0,0,0,0"
		cron.Task = cronx.DefaultTwdlTask()
	case CategoryReDlgd:
		// 每月1~26号前1~6点执行
		cron.Scheduler = fmt.Sprintf("%d %d 1-6 1-26 * *", s, m)
		// 固定且低频，每月固定执行
		cron.DeltaTime = "0,0,0,0,0,0"
		cron.Task = ""
	}

	return cron
}
