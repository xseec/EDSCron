package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/slicex"
)

// Category 定义任务类型枚举
type Category string

const (
	CategoryDlgd     Category = "dlgd"      // 代理购电
	CategoryHoliday  Category = "holiday"   // 节假日
	CategoryWeather  Category = "weather"   // 天气
	CategoryCarbon   Category = "carbon"    // 大陆碳排因子
	CategoryTwCarbon Category = "tw-carbon" // 台湾碳排因子
	CategoryTwdl     Category = "twdl"      // 台湾电价
	CategoryReDlgd   Category = "re-dlgd"   // 重试代理购电
)

// Format 格式化并验证Cron任务配置
func Format(c *cron.CronBody) error {
	// 设置默认任务参数
	if len(c.Task) == 0 {
		c.Task = "{}"
	}

	// 设置默认开始时间
	if c.StartTime == 0 {
		c.StartTime = time.Now().Unix()
	}

	// 设置默认时间间隔
	if len(regexp.MustCompile(`\d+`).FindAllString(c.DeltaTime, -1)) != 6 {
		switch Category(c.Category) {
		case CategoryDlgd:
			c.DeltaTime = "0,1,0,0,0,0" // 1月
		case CategoryHoliday:
			c.DeltaTime = "0,1,0,0,0,0" // 1月(+1月即可跳过12月)
		case CategoryWeather:
			c.DeltaTime = "0,0,1,0,0,0" // 1天
		case CategoryReDlgd, CategoryCarbon, CategoryTwdl, CategoryTwCarbon:
			c.DeltaTime = "0,0,0,0,0,0"
		default:
			return fmt.Errorf("无效的任务类型: %s, 可选值: dlgd, holiday, weather, carbon, re-dlgd, tw-carbon, twdl", c.Category)
		}
	}

	// 设置默认调度时间
	if len(c.Scheduler) == 0 {
		s := rand.Intn(60) // 随机秒
		m := rand.Intn(60) // 随机分钟
		h := rand.Intn(24) // 随机小时

		switch Category(c.Category) {
		// 基于http请求x3
		case CategoryCarbon:
			// 每月1/11/21号任意时间执行
			c.Scheduler = fmt.Sprintf("%d %d %d 1-21/10 * *", s, m, h)
		case CategoryHoliday:
			// 每年12月每日执行
			c.Scheduler = fmt.Sprintf("%d %d %d 1-31 12 *", s, m, h)
		case CategoryWeather:
			// 每日19:30后执行，每10分钟一次
			m = 30 + rand.Intn(10)
			c.Scheduler = fmt.Sprintf("%d %d-59/10 19-23 * * *", s, m)

		// 基于chromedp请求x4，需注意错开时间
		case CategoryDlgd:
			// 每月27-31号任意时间，多省份，一天多次
			h = rand.Intn(12)
			c.Scheduler = fmt.Sprintf("%d %d %d-23/6 27-31 * *", s, m, h)
		case CategoryReDlgd:
			// 每月1~26号前1~6点执行
			c.Scheduler = fmt.Sprintf("%d %d 1-6 1-26 * *", s, m)
		case CategoryTwdl:
			// 每月1~26号7点执行
			c.Scheduler = fmt.Sprintf("%d %d 7 1-26 * *", s, m)
		case CategoryTwCarbon:
			// 每月1~26号8点执行
			c.Scheduler = fmt.Sprintf("%d %d 8 1-26 * *", s, m)
		}
	}

	// 设置默认时间格式
	if len(c.Time) == 0 {
		switch Category(c.Category) {
		case CategoryDlgd:
			c.Time = "2006年1月" // 年月格式
		case CategoryHoliday:
			c.Time = "2006" // 年格式
		case CategoryTwCarbon:
			c.Time = "2006年"
		}
	}

	// 特殊验证：电量购电任务需要检查价格时段
	if c.Category == string(CategoryDlgd) {
		var dlgd cronx.DlgdConfig
		if err := json.Unmarshal([]byte(c.Task), &dlgd); err != nil {
			return fmt.Errorf("解析代理购电配置失败: %v", err)
		}
		return dlgd.CheckPriceHour()
	}

	return nil
}

// runReDlgd 执行重试电量购电任务
func runReDlgd(ctx context.Context, svc *ServiceContext) error {
	// 获取所有定时任务
	crons, err := svc.CronModel.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("获取定时任务列表失败: %v", err)
	}

	now := time.Now()
	// 查找上月底未执行成功的电量购电任务中随机一个，避免单个数据源异常卡住
	cron, ok := slicex.RandomFunc(*crons, func(c model.Cron) bool {
		return c.Category == string(CategoryDlgd) && (now.Year() != c.StartTime.Year() || now.Month() != c.StartTime.Month())
	})

	if !ok {
		return nil // 没有需要重试的任务
	}

	// 执行重试
	task := []byte(strings.ReplaceAll(cron.Task, cron.Time, now.Format(cron.Time)))
	if err := runDlgd(ctx, svc, task); err != nil {
		return fmt.Errorf("执行重试代理购电任务失败: %v", err)
	}

	// 更新任务开始时间
	cron.StartTime = time.Now()
	if err := svc.CronModel.Update(ctx, &cron); err != nil {
		return fmt.Errorf("更新任务开始时间失败: %v", err)
	}

	return nil
}

// runDlgd 执行电量购电任务
func runDlgd(ctx context.Context, svc *ServiceContext, task []byte) error {
	var cfg cronx.DlgdConfig
	var rows []model.Dlgd

	// 解析配置
	if err := json.Unmarshal(task, &cfg); err != nil {
		return fmt.Errorf("解析代理购电任务配置失败: %v", err)
	}

	// 设置OCR配置
	cfg.Ocr.Endpoint = &svc.Config.Ocr.Endpoint
	cfg.Ocr.AccessKeyId = &svc.Config.Ocr.AccessKeyId
	cfg.Ocr.AccessKeySecret = &svc.Config.Ocr.AccessKeySecret

	// 执行任务
	rsts, err := cfg.Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行代理购电任务失败: %v", err)
	}

	// 转换结果
	if err := copier.Copy(&rows, rsts); err != nil {
		return fmt.Errorf("转换代理购电结果失败: %v", err)
	}

	// 保存结果
	for _, v := range rows {
		old, _ := svc.DlgdModel.FindOneByAreaStartTimeCategoryVoltageStage(ctx, v.Area, v.StartTime, v.Category, v.Voltage, v.Stage)
		if old != nil {
			v.Id = old.Id
			err = svc.DlgdModel.Update(ctx, &v)
		} else {
			_, err = svc.DlgdModel.Insert(ctx, &v)
		}

		if err != nil {
			return fmt.Errorf("保存代理购电结果失败: %v", err)
		}
	}

	return nil
}

// runTwdl 执行台湾电量购电任务
func runTwdl(ctx context.Context, svc *ServiceContext, task *[]byte) error {
	var cfg cronx.TwdlConfig
	var rows []model.Twdl
	var holidays []model.Holiday

	// 解析配置
	if err := json.Unmarshal(*task, &cfg); err != nil {
		return fmt.Errorf("解析台湾电力配置失败: %v", err)
	}

	// 设置OCR配置
	cfg.Ocr.Endpoint = &svc.Config.Ocr.Endpoint
	cfg.Ocr.AccessKeyId = &svc.Config.Ocr.AccessKeyId
	cfg.Ocr.AccessKeySecret = &svc.Config.Ocr.AccessKeySecret

	// 执行任务
	rsts, days, err := (&cfg).Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行台湾电力任务失败: %v", err)
	}

	// 转换结果
	if err := copier.Copy(&rows, rsts); err != nil {
		return fmt.Errorf("转换台湾电力结果失败: %v", err)
	}

	// 保存结果
	for _, v := range rows {
		old, _ := svc.TwdlModel.FindOneByStartTimeCategoryDate(ctx, v.StartTime, v.Category, v.Date)
		if old != nil {
			v.Id = old.Id
			err = svc.TwdlModel.Update(ctx, &v)
		} else {
			_, err = svc.TwdlModel.Insert(ctx, &v)
		}

		if err != nil {
			return fmt.Errorf("保存台湾电力结果失败: %v", err)
		}
	}

	// 保存离峰日结果
	if err := copier.Copy(&holidays, days); err != nil {
		return fmt.Errorf("转换离峰日结果失败: %v", err)
	}

	for _, v := range holidays {
		old, _ := svc.HolidayModel.FindOneByAreaDate(ctx, v.Area, v.Date)
		if old != nil {
			v.Id = old.Id
			err = svc.HolidayModel.Update(ctx, &v)
		} else {
			_, err = svc.HolidayModel.Insert(ctx, &v)
		}

		if err != nil {
			return fmt.Errorf("保存离峰日结果失败: %v", err)
		}
	}

	buf, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化台湾电力配置失败: %v", err)
	}

	*task = buf

	return nil
}

// runCarbon 执行碳排放任务
func runCarbon(ctx context.Context, svc *ServiceContext, task []byte) error {
	var cfg cronx.CarbonConfig
	var rows []model.Carbon

	if err := json.Unmarshal(task, &cfg); err != nil {
		return fmt.Errorf("解析碳排放配置失败: %v", err)
	}

	rsts, err := cfg.Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行碳排放任务失败: %v", err)
	}

	if err := copier.Copy(&rows, rsts); err != nil {
		return fmt.Errorf("转换碳排放结果失败: %v", err)
	}

	for _, v := range rows {
		old, _ := svc.CarbonModel.FindOneByAreaYear(ctx, v.Area, v.Year)
		if old != nil {
			v.Id = old.Id
			err = svc.CarbonModel.Update(ctx, &v)
		} else {
			_, err = svc.CarbonModel.Insert(ctx, &v)
		}

		if err != nil {
			return fmt.Errorf("保存碳排放结果失败: %v", err)
		}
	}

	return nil
}

// runTwCarbon 执行台湾碳排放任务
func runTwCarbon(ctx context.Context, svc *ServiceContext, task *[]byte) error {
	var cfg cronx.TwCarbonConfig
	var rows []model.Carbon

	if err := json.Unmarshal(*task, &cfg); err != nil {
		return fmt.Errorf("解析碳排放配置失败: %v", err)
	}

	rsts, err := (&cfg).Run(&svc.Config.Mail)

	if err != nil {
		return fmt.Errorf("执行碳排放任务失败: %v", err)
	}

	if err := copier.Copy(&rows, rsts); err != nil {
		return fmt.Errorf("转换碳排放结果失败: %v", err)
	}

	for _, v := range rows {
		old, _ := svc.CarbonModel.FindOneByAreaYear(ctx, v.Area, v.Year)
		if old != nil {
			v.Id = old.Id
			err = svc.CarbonModel.Update(ctx, &v)
		} else {
			_, err = svc.CarbonModel.Insert(ctx, &v)
		}

		if err != nil {
			return fmt.Errorf("保存碳排放结果失败: %v", err)
		}
	}

	buf, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化碳排放配置失败: %v", err)
	}

	*task = buf

	return nil
}

// runHoliday 执行节假日任务
func runHoliday(ctx context.Context, svc *ServiceContext, task []byte) error {
	var cfg cronx.HolidayConfig
	var rows []model.Holiday

	if err := json.Unmarshal(task, &cfg); err != nil {
		return fmt.Errorf("解析节假日配置失败: %v", err)
	}

	rsts, err := cfg.Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行节假日任务失败: %v", err)
	}

	if err := copier.Copy(&rows, rsts); err != nil {
		return fmt.Errorf("转换节假日结果失败: %v", err)
	}

	for _, v := range rows {
		old, _ := svc.HolidayModel.FindOneByAreaDate(ctx, v.Area, v.Date)
		if old != nil {
			v.Id = old.Id
			err = svc.HolidayModel.Update(ctx, &v)
		} else {
			_, err = svc.HolidayModel.Insert(ctx, &v)
		}

		if err != nil {
			return fmt.Errorf("保存节假日结果失败: %v", err)
		}
	}

	return nil
}

// runWeather 执行天气任务
func runWeather(ctx context.Context, svc *ServiceContext, task []byte) error {
	var cfg cronx.WeatherConfig
	var rows []model.Weather

	if err := json.Unmarshal(task, &cfg); err != nil {
		return fmt.Errorf("解析天气配置失败: %v", err)
	}

	rsts, err := cfg.Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行天气任务失败: %v", err)
	}

	if err := copier.Copy(&rows, rsts); err != nil {
		return fmt.Errorf("转换天气结果失败: %v", err)
	}

	for _, v := range rows {
		old, _ := svc.WeatherModel.FindOneByDateCity(ctx, v.Date, v.City)
		if old != nil {
			v.Id = old.Id
			err = svc.WeatherModel.Update(ctx, &v)
		} else {
			_, err = svc.WeatherModel.Insert(ctx, &v)
		}

		if err != nil {
			return fmt.Errorf("保存天气结果失败: %v", err)
		}
	}

	return nil
}
