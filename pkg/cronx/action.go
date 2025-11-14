package cronx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"seeccloud.com/edscron/pkg/chromedpx"

	"github.com/rs/xid"
)

// Action 定义任务处理步骤类型
type Action func() error

func crawlAndLocalizeAc(ctx context.Context, dp chromedpx.DP, path *string) Action {
	return func() error {
		url := ""
		err := dp.Run(ctx, &url)
		if err == nil {
			return localize(url, path)
		}

		if entries, err1 := os.ReadDir(dp.DownloadDir); err1 == nil && len(entries) > 0 {
			*path = filepath.Join(dp.DownloadDir, entries[0].Name())
			return nil
		}

		return err
	}
}

// crawlAc 创建网页抓取任务
func crawlAc(ctx context.Context, dp chromedpx.DP, name *string) Action {
	return func() error {
		return dp.Run(ctx, name)
	}
}

// localizeAc 创建文件下载任务
func localizeAc(url, path *string) Action {
	return func() error {
		return localize(*url, path)
	}
}

// ocrPdfAc 创建OCR识别任务
func ocrPdfAc(config AliOcr, inPath, outUrl *string) Action {
	return func() error {
		return aliConvertPDF(config, *inPath, outUrl, formatExcel)
	}
}

// pdfConvertAc 创建PDF转换任务
func pdfConvertAc(inPath, outUrl *string, format OutputFormat) Action {

	return func() error {
		return pdf24Convert(*inPath, format, outUrl)
	}
}

// cropContentAc 获取目录页任务
func cropPdfAc(inPath, outPath *string, crops *[]string) Action {
	*outPath = filepath.Join(tempDir, fmt.Sprintf("%s.pdf", xid.New().String()))
	return func() error {
		return crop(*inPath, *outPath, *crops)
	}
}
