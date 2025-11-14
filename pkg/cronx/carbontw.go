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
)

const (
	// 匹配台湾经济部能源署温室气体文件中的碳排系数
	twCarbonFactorPattern = `(\d+\.\d+)\s*公斤\s*CO2e/度`
)

// TwCarbonConfig 碳排放因子获取配置
type TwCarbonConfig struct {
	MaxRunYear int          `json:"max_run_year"` // 已执行的年份(公历)
	Dp         chromedpx.DP `json:"dp"`           // 网页爬虫配置
}

// DefaultTwCarbonTask 默认的台湾碳排放因子获取任务
func DefaultTwCarbonTask() string {
	cfg := TwCarbonConfig{
		Dp: chromedpx.DP{
			Urls: []chromedpx.DPUrl{
				{
					Url: "https://99z.top/https://www.moeaea.gov.tw/ecw/populace/content/SubMenu.aspx?menu_id=114",
				},
			},
			Outer: chromedpx.DPOuter{
				Host:     "https://99z.top/https://www.moeaea.gov.tw/ecw/populace/content",
				Pattern:  "href=\"([^\"]*)\"",
				Selector: `a[title*='2006年度電力排碳係數']`, // 2006为年份占位符
			},
		},
	}
	task, _ := json.Marshal(cfg)
	return string(task)
}

// Run 获取台湾的电力排碳系数
func (c *TwCarbonConfig) Run(m *MailConfig) (*[]CarbonFactor, error) {
	var url, pdfPath string
	var year int
	var value float64

	// c.Dp.IsVisible = true

	// 处理完成后清理临时文件（调试模式下保留）
	defer func() {
		if !c.Dp.IsVisible {
			os.Remove(pdfPath)
		}
	}()

	// 避免mustAdjustTwYear影响c.Dp
	adjustDp := c.Dp
	// 数据库或UI提供公历(2024年)，数据源用民国年份(113年)
	mustAdjustTwYear(&adjustDp, &year)
	// 已执行或新年份，无需执行
	if year == c.MaxRunYear || year >= time.Now().Year() {
		return nil, nil
	}

	ctx := context.Background()
	actions := []Action{
		crawlAc(ctx, adjustDp, &url),
		localizeAc(&url, &pdfPath),
		extractTwCarbonAc(&pdfPath, &value),
	}

	// 顺序执行各个处理步骤
	for i, ac := range actions {
		if err := ac(); err != nil {
			return nil, fmt.Errorf("执行任务步骤%d失败: %w", i, err)
		}

		if c.Dp.IsVisible {
			fmt.Printf("执行任务步骤%d成功\n", i)
		}
	}

	m.Send(TwCarbonTemplate{
		Year:  year,
		Value: value,
	}, pdfPath)
	if c.MaxRunYear < year {
		c.MaxRunYear = year
	}

	return &[]CarbonFactor{
		{
			Year:  int64(year),
			Area:  TaiwanAreaName,
			Value: value,
		},
	}, nil
}

func extractTwCarbonAc(path *string, value *float64) Action {
	return func() error {
		var content string
		if err := readPdf(*path, 1, &content); err != nil {
			return err
		}

		subs := regexp.MustCompile(twCarbonFactorPattern).FindStringSubmatch(content)
		if len(subs) != 2 {
			return fmt.Errorf("提取碳排因子失败, word: %s", content)
		}

		*value, _ = strconv.ParseFloat(subs[1], 64)
		return nil
	}
}
