package cronx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	pdf "github.com/pdfcpu/pdfcpu/pkg/api"
)

// https://tools.pdf24.org/zh/pdf-to-excel
// pdf24.org将PDF转换为Excel后能准确保留单元格背景色等信息

var (
	pdf24SubUrls = map[OutputFormat]string{
		formatExcel: "Excel",
		formatWord:  "Word",
		formatImage: "Png",
	}

	exts = map[OutputFormat]string{
		formatExcel: ".xlsx",
		formatWord:  ".docx",
		formatImage: ".png",
	}
)

// pdf24Convert 将PDF文件转换为指定格式并下载
//
// 参数:
//   - inPath: 输入PDF文件路径
//   - format: 输出格式，支持Excel、Word、图片
//   - outPath: 输出已下载文件路径指针
//
// 返回值:
//   - error: 错误信息
func pdf24Convert(inPath string, format OutputFormat, outPath *string) error {
	inExt := filepath.Ext(inPath)
	if strings.ToLower(inExt) != ".pdf" {
		return errors.New("输入文件格式错误")
	}

	// chromedp上传文件需绝对路径
	if !filepath.IsAbs(inPath) {
		inPath, _ = filepath.Abs(inPath)
	}

	downloadDir, _ := filepath.Abs(tempDir)
	os.MkdirAll(downloadDir, 0755)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("download.default_directory", downloadDir),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		// chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// 准备下载监听
	downloadComplete := make(chan struct{})
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*browser.EventDownloadProgress); ok {
			if ev.State == browser.DownloadProgressStateCompleted {
				close(downloadComplete)
			}
		}
	})

	// pdf24.org网址大小写敏感
	url := fmt.Sprintf("https://tools.pdf24.org/zh/pdf-to-%s", pdf24SubUrls[format])
	uploadId := fmt.Sprintf("#pdfTo%s > input", pdf24SubUrls[format])
	err := chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
			WithDownloadPath(downloadDir).
			WithEventsEnabled(true),
		chromedp.Navigate(strings.ToLower(url)),
		chromedp.SetUploadFiles(uploadId, []string{inPath}),
		chromedp.WaitNotPresent("button.btn.action.convert.disabled"),
		chromedp.Click("button.btn.action.convert"),
		chromedp.Click("#downloadTool"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			select {
			case <-downloadComplete:
				return nil
			case <-time.After(5 * time.Minute):
				return fmt.Errorf("下载超时")
			}
		}),
	)

	if err != nil {
		return err
	}

	fileName := strings.TrimRight(filepath.Base(inPath), inExt)
	ext := exts[format]
	if format == formatImage {
		cnt, _ := pdf.PageCountFile(inPath)
		if cnt > 1 {
			ext = ".zip"
		}
	}

	fileName = filepath.Join(tempDir, fileName+ext)
	if _, err := os.Stat(fileName); err != nil {
		return err
	}

	*outPath = fileName

	return nil

}
