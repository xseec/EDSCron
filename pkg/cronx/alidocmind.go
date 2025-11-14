package cronx

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	aliapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alidoc "github.com/alibabacloud-go/docmind-api-20220711/client"
	"github.com/alibabacloud-go/tea-utils/v2/service"
)

/*
使用示例:
err := aliConvertPDF(config, inPath, &outPath, FormatExcel)  // 转换为Excel
err := aliConvertPDF(config, inPath, &outPath, FormatWord)   // 转换为Word
err := aliConvertPDF(config, inPath, &outPath, FormatImage)  // 转换为图片

参数说明:
- config: 阿里云API配置
- inPath: 本地PDF文件路径
- outPath: 用于接收结果URL的指针
- format: 输出格式枚举(FormatExcel/FormatWord/FormatImage)
*/

// 文档转换响应结构体
type AliDocConvertResp struct {
	Data struct {
		Id string `json:"id"` // 任务ID
	} `json:"data"`
	Code    string `json:"code"`    // 错误码
	Message string `json:"message"` // 错误信息
}

// 转换结果响应结构体
type AliDocResultResp struct {
	Completed bool   `json:"completed"` // 是否完成
	Status    string `json:"status"`    // 状态(Success/Fail)
	Code      string `json:"code"`      // 错误码
	Message   string `json:"message"`   // 错误信息
	Data      []struct {
		Type string `json:"type"` // 文件类型
		Url  string `json:"url"`  // 文件URL
	} `json:"data"`
}

type AliOcr struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
}

// aliConvertPDF 使用阿里云文档智能服务转换PDF文件
func aliConvertPDF(cfg AliOcr, inPath string, outUrl *string, format OutputFormat) error {
	// 初始化客户端
	config := aliapi.Config{
		Endpoint:        &cfg.Endpoint,
		AccessKeyId:     &cfg.AccessKeyId,
		AccessKeySecret: &cfg.AccessKeySecret,
	}
	client, err := alidoc.NewClient(&config)
	if err != nil {
		return fmt.Errorf("初始化阿里云客户端失败: %w", err)
	}

	// 打开输入文件
	file, err := os.Open(inPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 根据格式选择不同的转换请求
	var responseBody []byte
	switch format {
	case formatExcel:
		request := alidoc.SubmitConvertPdfToExcelJobAdvanceRequest{
			FileName:      &inPath,
			FileUrlObject: file,
		}
		response, err := client.SubmitConvertPdfToExcelJobAdvance(&request, &service.RuntimeOptions{})
		if err != nil {
			return fmt.Errorf("提交Excel转换失败: %w", err)
		}
		responseBody = []byte(response.Body.String())

	case formatWord:
		request := alidoc.SubmitConvertPdfToWordJobAdvanceRequest{
			FileName:      &inPath,
			FileUrlObject: file,
		}
		response, err := client.SubmitConvertPdfToWordJobAdvance(&request, &service.RuntimeOptions{})
		if err != nil {
			return fmt.Errorf("提交Word转换失败: %w", err)
		}
		responseBody = []byte(response.Body.String())

	case formatImage:
		request := alidoc.SubmitConvertPdfToImageJobAdvanceRequest{
			FileName:      &inPath,
			FileUrlObject: file,
		}
		response, err := client.SubmitConvertPdfToImageJobAdvance(&request, &service.RuntimeOptions{})
		if err != nil {
			return fmt.Errorf("提交Image转换失败: %w", err)
		}
		responseBody = []byte(response.Body.String())

	default:
		return fmt.Errorf("不支持的输出格式: %s", format)
	}

	// 处理转换响应
	var resp AliDocConvertResp
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Data.Id == "" {
		return fmt.Errorf("阿里云服务错误: %s - %s", resp.Code, resp.Message)
	}

	return getDocResult(client, resp.Data.Id, outUrl)
}

// 获取转换结果
func getDocResult(client *alidoc.Client, id string, outPath *string) error {
	const (
		timeout  = 5 * time.Minute  // 超时时间
		interval = 10 * time.Second // 轮询间隔
	)

	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-time.After(timeout):
			return errors.New("转换超时(5分钟)")
		case <-timer.C:
			response, err := client.GetDocumentConvertResult(&alidoc.GetDocumentConvertResultRequest{
				Id: &id,
			})
			if err != nil {
				return fmt.Errorf("查询结果失败: %w", err)
			}

			responseBody := []byte(response.Body.String())
			var resp AliDocResultResp
			if err := json.Unmarshal(responseBody, &resp); err != nil {
				return fmt.Errorf("解析结果失败: %w", err)
			}

			if !resp.Completed {
				timer.Reset(interval)
				continue
			}

			if resp.Status == "Fail" {
				return fmt.Errorf("转换失败: %s - %s", resp.Code, resp.Message)
			}

			if len(resp.Data) == 0 {
				return errors.New("空结果")
			}

			*outPath = resp.Data[0].Url
			return nil
		}
	}
}
