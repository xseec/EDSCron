package cronx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"seeccloud.com/edscron/pkg/chromedpx"
	"seeccloud.com/edscron/pkg/x/expx"
)

var (
	// 生态环境部: 首页 > 政策文件 > 部文件 > 公告, %s: ""/_1/_2/_3
	carbonGovHost = "https://www.mee.gov.cn"
	carbonGovPat  = "https://www.mee.gov.cn/zcwj/bwj/gg/index%s.shtml"
	// 用“电力二氧化碳排放因子”，不用“电力碳足迹因子”
	carbonGovSel = `//a[contains(text(), "%s年电力二氧化碳排放因子")]`
	// carbonGovSel = `//a[contains(text(), "%s年") and contains(text(), "电力") and contains(text(), "碳") and contains(text(), "因子")]`
	// 主要用于产品碳足迹核算，不常用
	// CarbonFootprint = "电力碳足迹因子"
	// 主要用于核算电力消费的二氧化碳排放量，帮助企业进行碳管理和合规
	// CarbonEmission = "电力二氧化碳排放因子"
)

// CarbonFactor 表示一个碳排放因子记录
type CarbonFactor struct {
	Area  string  `json:"area"`  // 区域名称（如"全国"、"华东"、"福建省"）
	Year  int64   `json:"year"`  // 年份（如2020）
	Value float64 `json:"value"` // 净购入电力碳排放因子（单位：kgCO₂/kWh）
}

type CarbonGovConfig struct {
	Dp   chromedpx.DP `json:"dp"`
	Year int64        `json:"year"`
}

func DefaultCarbonGovTask() string {
	cfg := CarbonGovConfig{
		Year: 2006,
	}
	task, _ := json.Marshal(cfg)
	return string(task)
}

func (c CarbonGovConfig) Run(m *MailConfig) (*[]CarbonFactor, error) {
	factors := []CarbonFactor{}
	var pdfUrl, pdfPath, pdfText string
	defer func() {
		os.Remove(pdfPath)
	}()

	// 2021前，不发布数据；今年数据只能明年以后发布
	if c.Year < 2021 || c.Year >= int64(time.Now().Year()) {
		c.Year = 0
	}

	// 1. 捕获公告文件
	//    1.1 自定义选择器(当数据源变化，匹配失败时用
	//    1.2 默认选择器
	if len(c.Dp.Urls) > 0 {
		if err := c.Dp.Run(context.Background(), &pdfUrl); err != nil {
			return nil, err
		}
	} else {
		c.mustRun(m, &pdfUrl, &pdfText)

		if len(pdfText) > 0 {
			adjustCarbons(&factors, pdfText)
			return &factors, nil
		}
	}

	// 2. 下载公告文件
	if err := localize(pdfUrl, &pdfPath); err != nil {
		return nil, err
	}

	// 3. 解析公告文件
	if err := readPdf(pdfPath, -1, &pdfText); err != nil {
		return nil, err
	}

	// 4. 获取碳排因子表
	adjustCarbons(&factors, pdfText)

	return &factors, nil
}

func adjustCarbons(factors *[]CarbonFactor, value string) {
	value = regexp.MustCompile(`\s+`).ReplaceAllString(value, "")
	years := regexp.MustCompile(`(\d{4})年`).FindStringSubmatch(value)
	if len(years) != 2 {
		return
	}

	year, _ := strconv.ParseInt(years[1], 10, 64)
	// 格式如：全国0.5000,黑龙江0.6000
	subs := regexp.MustCompile(`(\p{Han}{2,3})(0\.\d{4})`).FindAllStringSubmatch(value, -1)
	facMap := map[string]any{}
	for _, sub := range subs {
		// 忽略：全国电力平均二氧化碳排放因子（不包括市场化交易的非化石能源电量）
		// 或：全国化石能源电力二氧化碳排放因子
		if _, ok := facMap[sub[1]]; ok {
			continue
		}

		facMap[sub[1]] = nil

		value, _ := strconv.ParseFloat(sub[2], 64)
		*factors = append(*factors, CarbonFactor{
			Year:  year,
			Area:  sub[1],
			Value: value,
		})
	}
}

func (c CarbonGovConfig) mustRun(m *MailConfig, carbonUrl *string, carbonText *string) {

	// 未指定有效年份时，从公告首页中捕获最新碳公告
	pageNum := expx.If(c.Year == 0, 1, 10)
	sel := fmt.Sprintf(carbonGovSel, expx.If(c.Year == 0, "", fmt.Sprintf("%d", c.Year)))
	ctx := context.Background()
	for i := range pageNum {

		// 1. 从公告列表中捕获公告页
		url := fmt.Sprintf(carbonGovPat, expx.If(i == 0, "", fmt.Sprintf("_%d", i)))
		dp := chromedpx.DP{
			Urls: []chromedpx.DPUrl{
				{
					Url: url,
				},
			},
			Outer: chromedpx.DPOuter{
				Host:     carbonGovHost,
				Selector: sel,
				Pattern:  `href="([^"]+)"`,
			},
		}
		if err := dp.Run(ctx, &url); err != nil {
			continue
		}

		// 2. 从公告页中捕获“附件”
		// 2.1 附件是PDF链接
		dp = chromedpx.DP{
			Urls: []chromedpx.DPUrl{
				{
					Url:    url,
					Clicks: []string{sel},
				},
			},
		}
		if err := dp.Run(ctx, carbonUrl); err == nil {
			return
		}

		// 2.2 附件是文本区块
		dp = chromedpx.DP{
			Urls: []chromedpx.DPUrl{
				{
					Url: url,
				},
			},
			Outer: chromedpx.DPOuter{
				OnlyText: true,
				Selector: `div.content_body`,
			},
		}
		// 2.3 数据源格式变化通知
		err := dp.Run(ctx, carbonText)
		if err != nil && m != nil {
			m.Send(EmptyDataTemplate{
				TaskDetail: fmt.Sprintf("数据源：%s", url),
			})
		}

		return
	}
}
