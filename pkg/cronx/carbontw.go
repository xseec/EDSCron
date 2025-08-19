package cronx

import (
	"context"
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
	MaxRunYear int          // 已执行的年份(公历)
	Dp         chromedpx.DP `json:"dp"` // 网页爬虫配置
}

// Run 获取台湾的电力排碳系数
//
// 参数:
//   - m: 邮件配置（用于错误处理）
//
// 返回:
//   - []CarbonFactor: 碳排放因子列表
//   - error: 错误信息
func (c *TwCarbonConfig) Run(m *MailConfig) (*[]CarbonFactor, error) {
	var url, pdfPath, wordPath string
	var year int
	var value float64

	// 处理完成后清理临时文件（调试模式下保留）
	defer func() {
		if !c.Dp.IsVisible {
			os.Remove(pdfPath)
			os.Remove(wordPath)
		}
	}()

	// 数据库或UI提供公历(2024年)，数据源用民国年份(113年)
	mustAdjustTwYear(&c.Dp, &year)
	// 已执行或新年份，无需执行
	if year == c.MaxRunYear || year >= time.Now().Year() {
		return nil, nil
	}

	ctx := context.Background()
	actions := []Action{
		crawlAc(ctx, c.Dp, &url),
		localizeAc(&url, &pdfPath),
		pdfConvertAc(&pdfPath, &wordPath, FormatWord),
		extractTwCarbonAc(&wordPath, &value),
	}

	// 顺序执行各个处理步骤
	for i, ac := range actions {
		if err := ac(); err != nil {
			return nil, fmt.Errorf("执行任务步骤%d失败: %w", i, err)
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
			Area:  "台湾",
			Value: value,
		},
	}, nil
}

func extractTwCarbonAc(path *string, value *float64) Action {
	return func() error {
		var content string
		if err := readWord(*path, &content); err != nil {
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
