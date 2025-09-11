package cronx

import (
	"context"
	"fmt"
	"html/template"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	aliapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	pdf "github.com/pdfcpu/pdfcpu/pkg/api"
	"seeccloud.com/edscron/pkg/chromedpx"
	"seeccloud.com/edscron/pkg/x/slicex"
)

// TwdlConfig 台湾电价获取配置
type TwdlConfig struct {
	FileSize string        `json:"file_size"` // 电价表文件大小，因页面上无任何时间标签，基于文件大小判定是否发布新电价
	Ocr      aliapi.Config `json:"ocr"`       // 阿里云OCR配置
	Dp       chromedpx.DP  `json:"dp"`        // 网页爬虫配置
}

func (c *TwdlConfig) Run(m *MailConfig) (*[]TwdlRow, *[]Holiday, error) {

	var url, calUrl, calPdfPath, calExcelPath, pdfPath, wordPath, firstPafPath, subsPdfPath, excelPath string
	startDate := ""
	fileSize := c.FileSize
	page := make([]string, 1)
	values := make([]TwdlRow, 0)
	offPeakDays := make([]string, 0)

	defer func() {
		if !c.Dp.IsVisible {
			os.Remove(pdfPath)
			os.Remove(wordPath)
			os.Remove(firstPafPath)
			os.Remove(subsPdfPath)
			os.Remove(excelPath)
			os.Remove(calPdfPath)
			os.Remove(calExcelPath)
		}
	}()

	ctx := context.Background()

	preloadActions := []Action{
		cropPdfAc(&pdfPath, &firstPafPath, &[]string{"1"}),
		pdfConvertAc(&firstPafPath, &wordPath, formatWord),
		extractFirstPageAc(&wordPath, &startDate, &page),
	}

	calActions := []Action{
		localizeAc(&calUrl, &calPdfPath),
		pdfConvertAc(&calPdfPath, &calExcelPath, formatExcel),
		excelizeCalAc(&calExcelPath, &offPeakDays),
	}

	actions := []Action{
		crawlAc(ctx, c.Dp, &url),
		extractContentAc(&fileSize, &url, &calUrl),
		adjustCalendarAc(&calUrl, &calActions),
		localizeAc(&url, &pdfPath),
		adjustPreloadAc(&pdfPath, &page, &preloadActions),
		cropPdfAc(&pdfPath, &subsPdfPath, &page),
		ocrPdfAc(c.Ocr, &subsPdfPath, &url),
		localizeAc(&url, &excelPath),
		excelizeTwdlAc(&excelPath, &startDate, &values),
	}

	for i, ac := range actions {
		if err := ac(); err != nil {
			return nil, nil, fmt.Errorf("执行任务步骤%d失败: %w", i, err)
		}
	}

	if len(values) > 0 {
		// 邮件含附近，用go会被中断
		m.Send(TwdlTemplate{
			Calendar:    strings.Join(offPeakDays, ", "),
			StartDate:   startDate,
			RecordCount: len(values),
			TargetCount: len(TwdlCategories) * 2,
			Details:     template.HTML(formatTwdls(values, true)),
		}, calPdfPath, calExcelPath, pdfPath, wordPath, excelPath)

		c.FileSize = fileSize

		days := slicex.MapFunc(offPeakDays, func(s string) Holiday {
			return Holiday{
				Area:     string(TaiwanArea),
				Alias:    taiwanAreaName,
				Date:     s,
				Category: string(HolidayPeakOff),
			}
		})
		return &values, &days, nil
	}

	return nil, nil, fmt.Errorf("空数据错误: %v", c)
}

func extractContentAc(fileSize, url, calendarUrl *string) Action {
	return func() error {
		content := regexp.MustCompile(`[\r\n]+`).ReplaceAllString(*url, "")
		reg1 := regexp.MustCompile(`(?s)<a\s+[^>]*href=["']([^"']*簡要電價表[^"']*)["'][^>]*>.*?(\d+(?:\.\d+)?[KMkm][Bb]).*?</a>`)
		reg2 := regexp.MustCompile(`(?s)<a\s+[^>]*href=["']([^"']*日曆表[^"']*)["'][^>]*>.*?</a>`)

		matches := reg1.FindStringSubmatch(content)
		if len(matches) != 3 {
			return fmt.Errorf("找不到含'簡要電價表'和文件大小的链接(<a>)")
		}

		// 未发布新电价
		if strings.EqualFold(matches[2], *fileSize) {
			return fmt.Errorf("未发布新电价")
		}

		// 发布新电价
		*fileSize = matches[2]
		*url = matches[1]

		// 电价日历
		matches = reg2.FindStringSubmatch(content)
		if len(matches) == 2 {
			*calendarUrl = matches[1]
		}

		return nil

	}
}

func adjustPreloadAc(pdfPath *string, bodyPages *[]string, actions *[]Action) Action {

	return func() error {
		cnt, err := pdf.PageCountFile(*pdfPath)
		if err != nil {
			return err
		}

		pageNum := math.Min(float64(cnt), float64(maxPage))
		(*bodyPages)[0] = fmt.Sprintf("1-%d", int(pageNum))

		if cnt > maxPage {
			for _, ac := range *actions {
				if err := ac(); err != nil {
					return fmt.Errorf("执行目录页任务失败: %w", err)
				}
			}
		}

		return nil
	}
}

// extractFirstPageAc 提取目录页信息任务
func extractFirstPageAc(wordPath, startDate *string, page *[]string) Action {
	// 提取任务允许失败
	return func() error {
		var content string
		if err := readWord(*wordPath, &content); err != nil {
			return nil
		}

		if match := twStartDateReg.FindStringSubmatch(content); len(match) == 4 {
			*startDate = fmt.Sprintf("%s年%s月%s日", match[1], match[2], match[3])
		}

		if subs := twPageNumReg.FindStringSubmatch(content); len(subs) == 3 {
			// "壹、調整後各類用電電價表...1"，基于目录页偏移+1
			start, _ := strconv.Atoi(subs[1])
			// "貳、凍漲行業適用電價表...10"，不含“10”
			end, _ := strconv.Atoi(subs[2])
			(*page)[0] = fmt.Sprintf("%d-%d", start+1, end)
		}

		return nil
	}
}

func excelizeTwdlAc(excelPath, startDate *string, twdls *[]TwdlRow) Action {
	return func() error {
		return excelizeTwdl(*excelPath, *startDate, twdls)
	}
}

func adjustCalendarAc(calUrl *string, action *[]Action) Action {
	return func() error {
		if len(*calUrl) == 0 {
			return nil
		}

		for _, ac := range *action {
			if err := ac(); err != nil {
				return err
			}
		}

		return nil
	}
}

func excelizeCalAc(excelPath *string, dates *[]string) Action {
	return func() error {
		dts, err := ExcelizeCalendar(*excelPath)
		if err != nil {
			return err
		}

		*dates = append(*dates, dts...)
		return nil
	}
}
