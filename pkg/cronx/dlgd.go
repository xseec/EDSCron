package cronx

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/zeromicro/go-zero/core/logx"
	"seeccloud.com/edscron/pkg/x/timex"
)

// DlgdRow 代理购电电价条目
type DlgdRow struct {
	Area       string    `json:"area"`                              // 区域名称
	StartTime  time.Time `json:"start_time"`                        // 起始时间(Unix时间戳，包含)
	EndTime    time.Time `json:"end_time"`                          // 结束时间(Unix时间戳，不包含)
	Category   string    `json:"category" dlgd:"用电分类|用电类别"`         // 用电类别
	Voltage    string    `json:"voltage" dlgd:"电压等级"`               // 电压等级
	Stage      string    `json:"stage"`                             // 阶梯电量阈值
	Fund       float64   `json:"fund" dlgd:"政府性基金及附加"`              // 政府性基金及附加
	Sharp      float64   `json:"sharp" dlgd:"尖峰" unit:"true"`       // 尖峰电价
	Peak       float64   `json:"peak" dlgd:"高峰|^峰时段" unit:"true"`   // 高峰电价，仅"峰时段"会匹配"尖峰时段"
	Flat       float64   `json:"flat" dlgd:"平段|平时段" unit:"true"`    // 平段电价
	Valley     float64   `json:"valley" dlgd:"低谷|^谷时段" unit:"true"` // 低谷电价，仅"谷时段"会匹配"深谷时段"
	Deep       float64   `json:"deep" dlgd:"深谷" unit:"true"`        // 深谷电价
	Demand     float64   `json:"demand" dlgd:"\\W*元/千瓦\\W月\\W*"`    // 需量电价
	Capacity   float64   `json:"capacity" dlgd:"\\W*元/千伏安\\W月\\W*"` // 容量电价
	SharpDate  string    `json:"sharp_date"`                        // 尖峰日期条件
	SharpHour  string    `json:"sharp_hour"`                        // 尖峰时段，如"1100-1130,2200-0700"
	PeakDate   string    `json:"peak_date"`                         // 高峰日期条件
	PeakHour   string    `json:"peak_hour"`                         // 高峰时段
	FlatDate   string    `json:"flat_date"`                         // 平段日期条件
	FlatHour   string    `json:"flat_hour"`                         // 平段时段
	ValleyDate string    `json:"valley_date"`                       // 低谷日期条件
	ValleyHour string    `json:"valley_hour"`                       // 低谷时段
	DeepDate   string    `json:"deep_date"`                         // 深谷日期条件
	DeepHour   string    `json:"deep_hour"`                         // 深谷时段
	DocNo      string    `json:"doc_no"`                            // 电价政策文档编号
}

// Run 执行代理购电任务，返回电价列表
// 参数m用于任务失败时通知系统管理员
func (d DlgdConfig) Run(m *MailConfig) (*[]DlgdRow, *[]DlgdHour, error) {
	var url, pdf, excel string
	var actionOffset int

	// 【调试模式】：正式版本注释掉以下行
	testMode := false
	// d.Dp.IsVisible = testMode
	// excel = "temp/d4a5161083iort2c3q60.xlsx"
	// actionOffset = 6

	defer func() {
		// 调试模式下保留过程文件，否则清理临时文件
		if !d.Dp.IsVisible {
			os.Remove(pdf)
			os.Remove(excel)
			os.Remove(d.Dp.DownloadDir)
		}
	}()

	rows := make([]DlgdRow, 0)
	hours := make([]DlgdHour, 0)
	ctx := context.Background()
	// 定义任务执行流程
	actions := []Action{
		crawlAndLocalizeAc(ctx, d.Dp, &pdf),                     // 0-网页抓取并下载
		cropAc(&pdf, d.TitlePat),                                // 1-裁剪处理
		thresholdAc(&pdf, d.Threshold),                          // 2-图片阈值处理
		image2PdfAc(&pdf),                                       // 3-图片转PDF
		ocrPdfAc(d.Ocr, &pdf, &url),                             // 4-OCR识别
		localizeAc(&url, &excel),                                // 5-下载OCR结果
		unexcelizeAc(&excel, d.Area, d.TitlePat, &rows, &hours), // 6-解析Excel数据
	}

	// 顺序执行各个处理步骤
	for i, ac := range actions[actionOffset:] {
		if err := ac(); err != nil {
			return nil, nil, fmt.Errorf("执行任务步骤%d失败: %w", i, err)
		}

		if testMode {
			logx.Infof("执行%sdlgd任务步骤%d/%d成功\n", d.Area, i+1, len(actions)-actionOffset)
		}
	}

	rows = specialiseDlgd(rows)
	rowss, err := emptyValueErr(m, d, &rows)
	if err != nil {
		return nil, nil, err
	}

	return rowss, &hours, nil

}

// cropAc 创建文件裁剪任务
func cropAc(name *string, pattern string) Action {
	return func() error {
		ext := strings.ToLower(path.Ext(*name))
		if ext != ".pdf" {
			return nil
		}

		pageCount, err := api.PageCountFile(*name)
		if err != nil {
			return fmt.Errorf("获取PDF页面数量失败: %w", err)
		}

		re := regexp.MustCompile(pattern)
		re2 := regexp.MustCompile(fmt.Sprintf(`%s|%s`, voltageName, voltageSZName))
		for i := 1; i <= pageCount; i++ {
			content := ""
			err = readPdf(*name, i, &content)
			if err != nil {
				return fmt.Errorf("读取PDF页面%d内容失败: %w", i, err)
			}

			if re.MatchString(content) && re2.MatchString(content) {
				err = crop(*name, *name, []string{fmt.Sprintf("%d", i)})
				if err != nil {
					return fmt.Errorf("裁剪PDF页面%d失败: %w", i, err)
				}

				break
			}
		}

		return nil
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
func unexcelizeAc(excel *string, area string, titlePat string, rows *[]DlgdRow, hours *[]DlgdHour) Action {
	return func() error {
		return unexcelize(*excel, area, titlePat, rows, hours)
	}
}

// AutoFill 为电价列表填充字段
func (d DlgdConfig) AutoFill(rows *[]DlgdRow, hours *[]DlgdHour) error {
	if rows == nil || len(*rows) == 0 || hours == nil || len(*hours) == 0 {
		return fmt.Errorf("电价或时段列表为空")
	}

	docNo := (*hours)[0].DocNo
	month := timex.MustTime(d.Month)

	hourValues := map[string][10]string{}
	for i, v := range *rows {
		v.StartTime = month
		v.EndTime = month.AddDate(0, 1, 0)
		v.DocNo = docNo

		values, ok := hourValues[v.Category]
		if !ok {
			if vs, ok := mergeHours(*hours, v.Category, int64(month.Month())); ok {
				hourValues[v.Category] = vs
				values = vs
			} else {
				return fmt.Errorf("获取电价时段信息失败: %s", v.Category)
			}
		}

		v.SharpDate = values[0]
		v.SharpHour = values[1]
		v.PeakDate = values[2]
		v.PeakHour = values[3]
		v.FlatDate = values[4]
		v.FlatHour = values[5]
		v.ValleyDate = values[6]
		v.ValleyHour = values[7]
		v.DeepDate = values[8]
		v.DeepHour = values[9]

		(*rows)[i] = v
	}

	return nil
}
