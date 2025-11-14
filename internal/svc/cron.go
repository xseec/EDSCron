package svc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/expx"
	"seeccloud.com/edscron/pkg/x/slicex"
)

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
		return c.Category == string(model.CategoryDlgd) && (now.Year() != c.StartTime.Year() || now.Month() != c.StartTime.Month())
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
	var mini cronx.MiniDlgdConfig

	// 解析配置
	if err := json.Unmarshal(task, &mini); err != nil {
		return fmt.Errorf("解析代理购电任务配置失败: %v", err)
	}

	cfg := cronx.NewDlgdConfig(mini, cronx.AliOcr{
		Endpoint:        svc.Config.Ocr.Endpoint,
		AccessKeyId:     svc.Config.Ocr.AccessKeyId,
		AccessKeySecret: svc.Config.Ocr.AccessKeySecret,
	})

	// 执行任务
	dlgdRows, dlgdHours, err := cfg.Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行代理购电任务失败: %v", err)
	}

	// 时段校验
	docNo := expx.If(len(*dlgdHours) > 0, (*dlgdHours)[0].DocNo, "")
	oldHours, _ := svc.DlgdHourModel.QueryAll(ctx, cfg.Area, docNo)
	if oldHours == nil || len(*oldHours) == 0 || (*oldHours)[0].Confirm != model.DlgdHourConfirmCode {
		hours := []model.DlgdHour{}
		copierx.MustCopy(&hours, dlgdHours)
		if err := svc.DlgdHourModel.MustInertAll(ctx, &hours); err != nil {
			return fmt.Errorf("保存时段划分结果失败: %v", err)
		}

		svc.Config.Mail.Send(cronx.DlgdHourConfirmTemplate{
			Area:   cfg.Area,
			Month:  cfg.Month,
			DocNo:  docNo,
			Detail: template.HTML(model.FormatHtmlDlgdHours(&hours)),
		})

		return errors.New("时段划分待确认")
	}

	// 保存电价记录
	confirmHours := []cronx.DlgdHour{}
	copierx.MustCopy(&confirmHours, oldHours)
	cfg.AutoFill(dlgdRows, &confirmHours)

	for _, r := range *dlgdRows {
		row := model.Dlgd{}
		copierx.MustCopy(&row, &r)
		old, _ := svc.DlgdModel.FindOneByAreaStartTimeCategoryVoltageStage(ctx, row.Area, row.StartTime, row.Category, row.Voltage, row.Stage)
		if old != nil {
			row.Id = old.Id
			err = svc.DlgdModel.Update(ctx, &row)
		} else {
			_, err = svc.DlgdModel.Insert(ctx, &row)
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
	cfg.Ocr = cronx.AliOcr{
		Endpoint:        svc.Config.Ocr.Endpoint,
		AccessKeyId:     svc.Config.Ocr.AccessKeyId,
		AccessKeySecret: svc.Config.Ocr.AccessKeySecret,
	}

	// 执行任务
	rsts, days, err := (&cfg).Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行台湾电力任务失败: %v", err)
	}

	// 转换结果
	copierx.MustCopy(&rows, rsts)

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
	copierx.MustCopy(&holidays, days)

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
	var cfg cronx.CarbonGovConfig
	var rows []model.Carbon

	if err := json.Unmarshal(task, &cfg); err != nil {
		return fmt.Errorf("解析碳排放配置失败: %v", err)
	}

	rsts, err := cfg.Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行碳排放任务失败: %v", err)
	}

	copierx.MustCopy(&rows, rsts)

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

	copierx.MustCopy(&rows, rsts)

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
	var cfg cronx.HolidayGovConfig
	var rows []model.Holiday

	if err := json.Unmarshal(task, &cfg); err != nil {
		return fmt.Errorf("解析节假日配置失败: %v", err)
	}

	rsts, err := cfg.Run(&svc.Config.Mail)
	if err != nil {
		return fmt.Errorf("执行节假日任务失败: %v", err)
	}

	copierx.MustCopy(&rows, rsts)

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

	copierx.MustCopy(&rows, rsts)

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
