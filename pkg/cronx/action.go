package cronx

import (
	"context"
	"fmt"
	"path/filepath"

	"seeccloud.com/edscron/pkg/chromedpx"

	aliapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/rs/xid"
)

// Action 定义任务处理步骤类型
type Action func() error

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
func ocrPdfAc(config aliapi.Config, inPath, outUrl *string) Action {
	return func() error {
		return aliConvertPDF(config, *inPath, outUrl, FormatExcel)
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
