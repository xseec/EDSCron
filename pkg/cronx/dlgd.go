package cronx

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	aliapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"seeccloud.com/edscron/pkg/chromedpx"
	"seeccloud.com/edscron/pkg/x/slicex"
)

// PriceHour 定义电价时段划分规则
type PriceHour struct {
	Months []int    `json:"months"` // 适用的月份列表(1-12)
	Days   []string `json:"days"`   // 特殊日期定义，支持多种格式:
	// - 气温条件: "weather:广州>=35"
	// - 周末: "weekend:1:2:4:7" (1,2,4,7月)
	// - 星期: "sun:1:2:4:7" (1,2,4,7月的周日)
	// - 节假日: "holiday:元旦:春节"
	Sharp  string `json:"sharp"`  // 尖峰时段，格式"1100-1200,1800-1830,2200-0700"
	Peak   string `json:"peak"`   // 高峰时段
	Flat   string `json:"flat"`   // 平时段
	Valley string `json:"valley"` // 低谷时段
	Deep   string `json:"deep"`   // 深谷时段，通常与特殊日期搭配使用
}

// DlgdConfig 代理购电任务配置
type DlgdConfig struct {
	Area  string       `json:"area"`  // 区域名称
	Month string       `json:"month"` // 执行月份，格式"2025年3月"
	Dp    chromedpx.DP `json:"dp"`    // 网页爬虫配置
	Crops []string     `json:"crops"` // 裁剪范围配置:
	// PDF: [起始页,结束页] 从1开始，如[1,3]
	// 图片: [x1(%宽度),y1(%高度),x2(%宽度),y2(%高度)]
	Threshold  int           `json:"threshold"`  // 图片底纹阈值(230~245)
	Ocr        aliapi.Config `json:"ocr"`        // 阿里云OCR配置
	Validators []string      `json:"validators"` // 验证文本列表，如公文号"国办发明电〔2024〕12号"
	PriceHours []PriceHour   `json:"priceHours"` // 电价时段划分规则
}

// Dlgd 代理购电用户电价表
type Dlgd struct {
	Rows    []DlgdRow `json:"rows"`    // 电价数据行
	Comment string    `json:"comment"` // 备注信息
}

// DlgdRow 代理购电电价明细
type DlgdRow struct {
	Area      string  `json:"area"`                              // 区域名称
	StartTime int64   `json:"start_time"`                        // 起始时间(Unix时间戳，包含)
	EndTime   int64   `json:"end_time"`                          // 结束时间(Unix时间戳，不包含)
	Category  string  `json:"category" dlgd:"用电分类|用电类别"`         // 用电类别
	Voltage   string  `json:"voltage" dlgd:"电压等级"`               // 电压等级
	Stage     string  `json:"stage"`                             // 阶梯电量阈值
	Fund      float64 `json:"fund" dlgd:"政府性基金及附加"`              // 政府性基金及附加
	Sharp     float64 `json:"sharp" dlgd:"尖峰" unit:"true"`       // 尖峰电价
	Peak      float64 `json:"peak" dlgd:"高峰|峰时段" unit:"true"`    // 高峰电价
	Flat      float64 `json:"flat" dlgd:"平段|平时段" unit:"true"`    // 平段电价
	Valley    float64 `json:"valley" dlgd:"低谷|谷时段" unit:"true"`  // 低谷电价
	Deep      float64 `json:"deep" dlgd:"深谷" unit:"true"`        // 深谷电价
	Demand    float64 `json:"demand" dlgd:"\\W*元/千瓦\\W月\\W*"`    // 需量电价
	Capacity  float64 `json:"capacity" dlgd:"\\W*元/千伏安\\W月\\W*"` // 容量电价

	// 以下字段定义特殊时段的日期和时间范围
	SharpDate  string `json:"sharp_date"`  // 尖峰日期条件
	SharpHour  string `json:"sharp_hour"`  // 尖峰时段，如"1100-1130,2200-0700"
	PeakDate   string `json:"peak_date"`   // 高峰日期条件
	PeakHour   string `json:"peak_hour"`   // 高峰时段
	FlatDate   string `json:"flat_date"`   // 平段日期条件
	FlatHour   string `json:"flat_hour"`   // 平段时段
	ValleyDate string `json:"valley_date"` // 低谷日期条件
	ValleyHour string `json:"valley_hour"` // 低谷时段
	DeepDate   string `json:"deep_date"`   // 深谷日期条件
	DeepHour   string `json:"deep_hour"`   // 深谷时段
}

// Run 执行代理购电任务，返回电价列表
// 参数m用于任务失败时通知系统管理员
func (d DlgdConfig) Run(m *MailConfig) (*[]DlgdRow, error) {
	var url, file, excel string
	defer func() {
		// 调试模式下保留过程文件，否则清理临时文件
		if !d.Dp.IsVisible {
			os.Remove(file)
			os.Remove(excel)
		}
	}()

	dlgd := &Dlgd{}
	row := d.initDlgdRow()

	ctx := context.Background()
	// 定义任务执行流程
	actions := []Action{
		crawlAc(ctx, d.Dp, &url),                   // 网页抓取
		localizeAc(&url, &file),                    // 下载文件
		cropAc(&file, d.Crops),                     // 裁剪处理
		thresholdAc(&file, d.Threshold),            // 图片阈值处理
		image2PdfAc(&file),                         // 图片转PDF
		ocrPdfAc(d.Ocr, &file, &url),               // OCR识别
		localizeAc(&url, &excel),                   // 下载OCR结果
		unexcelizeAc(&excel, row, dlgd),            // 解析Excel数据
		validateAc(&dlgd.Comment, d.Validators, m), // 数据验证
	}

	// 顺序执行各个处理步骤
	for i, ac := range actions {
		if err := ac(); err != nil {
			return nil, fmt.Errorf("执行任务步骤%d失败: %w", i, err)
		}
	}

	return emptyValueErr(m, d, &dlgd.Rows)
}

// cropAc 创建文件裁剪任务
func cropAc(name *string, params []string) Action {
	return func() error {
		return crop(*name, *name, params)
	}
}

// thresholdAc 创建图片阈值处理任务
func thresholdAc(name *string, th int) Action {
	return func() error {
		return thresholding(*name, th)
	}
}

// image2PdfAc 创建图片转PDF任务
func image2PdfAc(name *string) Action {
	return func() error {
		return image2Pdf(*name, name)
	}
}

// unexcelizeAc 创建Excel解析任务
func unexcelizeAc(excel *string, row DlgdRow, dlgd *Dlgd) Action {
	return func() error {
		return unexcelize(*excel, row, dlgd)
	}
}

// validateAc 创建数据验证任务
func validateAc(text *string, validators []string, m *MailConfig) Action {
	return func() error {
		for _, v := range validators {
			// 处理OCR识别可能导致的特殊字符差异
			// 例如："闽规﹝2024﹞01号文"可能被识别为"闽规[2024]01号文"
			reg1 := regexp.MustCompile("[^\u4e00-\u9fa5a-zA-Z0-9]")
			pat := reg1.ReplaceAllString(v, `[\s\S]{1,3}`)
			reg2 := regexp.MustCompile(pat)

			if !reg2.MatchString(*text) {
				go m.Send(&DlgdWarningTemplate{
					Remark: v,
					Rule:   *text,
				})
				return fmt.Errorf("验证失败[%s]", v)
			}
		}
		return nil
	}
}

// initDlgdRow 初始化电价信息，包括区域、时间范围和峰谷时段配置
func (c DlgdConfig) initDlgdRow() (d DlgdRow) {
	d.Area = c.Area

	// 解析月份时间范围
	start, err := time.Parse("2006年1月", c.Month)
	if err != nil {
		return
	}

	d.StartTime = start.Unix()
	d.EndTime = start.AddDate(0, 1, 0).Unix()

	// 处理电价时段配置
	for _, ph := range c.PriceHours {
		currentMonth := int(start.Month())

		// 1. 首先处理月份匹配的常规时段配置
		if slices.Contains(ph.Months, currentMonth) {
			d.DeepHour = ph.Deep
			d.ValleyHour = ph.Valley
			d.FlatHour = ph.Flat
			d.PeakHour = ph.Peak
			d.SharpHour = ph.Sharp
		}

		// 2. 处理特殊日期配置
		days := make([]string, 0)
		for _, day := range ph.Days {
			parts := strings.Split(day, ":")
			if len(parts) < 2 {
				days = append(days, day)
				continue
			}

			// 处理周末/星期配置中的月份限定
			if strings.Contains(strings.ToLower(parts[0]), "weekend") ||
				strings.Contains(strings.ToLower(parts[0]), "sat") ||
				strings.Contains(strings.ToLower(parts[0]), "sun") {
				// 提取月份参数并转换为数字
				months := slicex.MapFunc(parts[1:], func(s string) int {
					m, _ := strconv.Atoi(s)
					return m
				})

				// 检查当前月份是否在限定范围内
				if slices.Contains(months, currentMonth) {
					days = append(days, parts[0])
				}
				continue
			}

			days = append(days, day)
		}

		// 如果有特殊日期配置，则更新对应时段
		if len(days) > 0 {
			date := strings.Join(days, ",")
			if len(ph.Deep) != 0 {
				d.DeepHour = ph.Deep
				d.DeepDate = date
			}
			if len(ph.Valley) != 0 {
				d.ValleyHour = ph.Valley
				d.ValleyDate = date
			}
			if len(ph.Flat) != 0 {
				d.FlatHour = ph.Flat
				d.FlatDate = date
			}
			if len(ph.Peak) != 0 {
				d.PeakHour = ph.Peak
				d.PeakDate = date
			}
			if len(ph.Sharp) != 0 {
				d.SharpHour = ph.Sharp
				d.SharpDate = date
			}
		}
	}

	return
}

// CheckPriceHour 验证峰谷时段配置是否覆盖全天24小时
func (c DlgdConfig) CheckPriceHour() error {
	for _, h := range c.PriceHours {
		hours := []string{h.Deep, h.Valley, h.Flat, h.Peak, h.Sharp}
		// 特殊情况下允许只配置一个时段(如高温天只设尖峰)
		if slicex.LenFunc(hours, func(s string) bool { return len(s) > 0 }) == 1 {
			return nil
		}

		// 计算总时长
		t, err := getDuration(strings.Join(hours, ","))
		if err != nil {
			return fmt.Errorf("时段解析失败: %w", err)
		}

		// 验证是否覆盖24小时
		if t != 3600*24 {
			return fmt.Errorf("时段配置不完整: 总时长%.1fh≠24h, 配置:%+v", float64(t)/3600.0, h)
		}
	}

	return nil
}

// GetHalfHourIndexs 将时段字符串转换为48点时段索引
// 参数s格式如"1200-1300,2300-0130"，返回[24,25,46,47,0,1,2]
func GetHalfHourIndexs(s string) []int {
	indexs := make([]int, 0)
	items := strings.Split(s, ",")

	for _, item := range items {
		reg := regexp.MustCompile(`(\d{2})(\d{2})-(\d{2})(\d{2})`)
		subs := reg.FindStringSubmatch(item)
		if len(subs) != 5 {
			continue
		}

		// 转换时间点为48点制索引
		nums := slicex.MapFunc(subs[1:], func(sub string) int {
			num, _ := strconv.Atoi(sub)
			return num
		})

		from := nums[0]*2 + nums[1]/30 // 起始点索引
		to := nums[2]*2 + nums[3]/30   // 结束点索引

		// 生成连续的时段索引
		for i := from; i != to; {
			indexs = append(indexs, i)
			if i == 47 { // 处理跨天情况
				i = 0
			} else {
				i++
			}
		}
	}

	return indexs
}

// getDuration 计算时段字符串的总秒数
// 参数s格式如"1200-1300,2200-0800"表示12:00-13:00和22:00-08:00
func getDuration(s string) (int, error) {
	s = strings.ReplaceAll(s, "2400", "0000") // 标准化时间格式
	durs := regexp.MustCompile(`(\d{4})\s*[~-]+\s*(\d{4})`).FindAllStringSubmatch(s, -1)
	total := 0

	for _, dur := range durs {
		// 解析时间点
		t1, err := time.Parse("1504", dur[1])
		if err != nil {
			return 0, fmt.Errorf("解析起始时间失败: %w", err)
		}

		t2, err := time.Parse("1504", dur[2])
		if err != nil {
			return 0, fmt.Errorf("解析结束时间失败: %w", err)
		}

		// 处理跨天情况
		if t2.Unix() <= t1.Unix() {
			t2 = t2.AddDate(0, 0, 1)
		}

		total += int(t2.Unix()) - int(t1.Unix())
	}

	return total, nil
}
