package svc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/internal/config"
	"seeccloud.com/edscron/model"
)

type ServiceContext struct {
	Config       config.Config
	CronModel    model.CronModel
	CarbonModel  model.CarbonModel
	DlgdModel    model.DlgdModel
	HolidayModel model.HolidayModel
	WeatherModel model.WeatherModel
	AreaModel    model.AreaModel
	Cr           *cron.Cron
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySql.DataSource)
	return &ServiceContext{
		Config:       c,
		CronModel:    model.NewCronModel(conn, c.CacheRedis),
		CarbonModel:  model.NewCarbonModel(conn, c.CacheRedis),
		DlgdModel:    model.NewDlgdModel(conn, c.CacheRedis),
		HolidayModel: model.NewHolidayModel(conn, c.CacheRedis),
		WeatherModel: model.NewWeatherModel(conn, c.CacheRedis),
		AreaModel:    model.NewAreaModel(conn, c.CacheRedis),
		Cr:           cron.New(),
	}
}

// Todo 执行单个定时任务
func (svc *ServiceContext) Todo(c *model.Cron, execTime string) {
	ctx := context.Background()
	// 替换任务中的时间变量为实际执行时间
	task := []byte(strings.ReplaceAll(c.Task, c.Time, execTime))

	var err error
	switch c.Category {
	case "re-dlgd": // 重试电量购电任务
		err = runReDlgd(ctx, svc)
	case "dlgd": // 电量购电任务
		err = runDlgd(ctx, svc, task)
	case "holiday": // 节假日任务
		err = runHoliday(ctx, svc, task)
	case "weather": // 天气任务
		err = runWeather(ctx, svc, task)
	case "carbon": // 碳排放任务
		err = runCarbon(ctx, svc, task)
	case "tw-carbon": // 台湾碳排放任务
		err = runTwCarbon(ctx, svc, &task)
	default:
		err = fmt.Errorf("未知的任务类型: %s", c.Category)
	}

	if err != nil {
		logx.Errorf("执行%s任务失败: %v", c.Category, err)
	}
}

// StartCron 启动所有定时任务
// 注意：当任务有增删改时，需要调用此方法重启所有任务
func (svc *ServiceContext) StartCron() {
	// 停止现有任务调度
	svc.Cr.Stop()
	// 创建新的调度器实例
	svc.Cr = cron.New()

	ctx := context.Background()
	// 获取所有定时任务
	crons, err := svc.CronModel.FindAll(ctx)
	if err != nil {
		logx.Errorf("获取定时任务列表失败: %v", err)
		return
	}

	// 遍历所有任务并添加到调度器
	for _, task := range *crons {
		// 使用局部变量避免闭包问题
		currentTask := task

		err = svc.Cr.AddFunc(currentTask.Scheduler, func() {
			// 检查任务是否已到开始执行时间
			if time.Now().Unix() < currentTask.StartTime {
				return
			}

			// 获取任务数据
			taskData, err := currentTask.NextTask()
			if err != nil {
				logx.Errorf("获取任务数据失败(ID=%d): %v", currentTask.Id, err)
				return
			}

			// 根据任务类型执行对应处理
			var execErr error
			switch currentTask.Category {
			case "re-dlgd":
				execErr = runReDlgd(ctx, svc)
			case "dlgd":
				execErr = runDlgd(ctx, svc, taskData)
			case "holiday":
				execErr = runHoliday(ctx, svc, taskData)
			case "weather":
				execErr = runWeather(ctx, svc, taskData)
			case "carbon":
				execErr = runCarbon(ctx, svc, taskData)
			case "tw-carbon":
				// 将最近执行年份信息写入任务数据库中
				execErr = runTwCarbon(ctx, svc, &taskData)
				currentTask.Task = string(taskData)
			default:
				execErr = fmt.Errorf("未知的任务类型: %s", currentTask.Category)
			}

			if execErr != nil {
				logx.Errorf("执行任务失败(ID=%d): %v", currentTask.Id, execErr)
				return
			}

			// 更新任务的下次执行时间
			currentTask.NextStartTime()
			if err := svc.CronModel.Update(ctx, &currentTask); err != nil {
				logx.Errorf("更新任务时间失败(ID=%d): %v", currentTask.Id, err)
			}
		})

		if err != nil {
			logx.Errorf("添加定时任务到调度器失败(ID=%d): %v", task.Id, err)
		}
	}

	// 启动定时任务调度器
	svc.Cr.Start()
	logx.Info("定时任务调度器已启动")
}
