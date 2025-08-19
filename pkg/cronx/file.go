package cronx

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"code.sajari.com/docconv"
	"github.com/rs/xid"

	pdf "github.com/pdfcpu/pdfcpu/pkg/api"
)

const (
	tempDir   = "temp"
	fileMode  = 0666
	dirMode   = 0755
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
)

var (
	fileFormats = map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"pdf":  true,
		"xlsx": true,
		"docx": true,
	}

	urlRegex     = regexp.MustCompile(`^(?:https?:\/\/)?(?:[\w-]+\.)+[\w-]+(?:\/[\p{L}\w\-._~:/?#[\]%@!$&'()*+,;=.]+)*$`)
	docTypeRegex = regexp.MustCompile(`(?i)documentType=(\w+)[&]*`)
	fileExtRegex = regexp.MustCompile(`\.(\w+)($|\?)`)
)

// 图片转PDF
func image2Pdf(in string, out *string) error {
	if strings.EqualFold(filepath.Ext(in), ".pdf") {
		return nil
	}

	if err := ensureDir(tempDir); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	outFile := filepath.Join(tempDir, fmt.Sprintf("%s.pdf", xid.New().String()))
	if err := pdf.ImportImagesFile([]string{in}, outFile, nil, nil); err != nil {
		return fmt.Errorf("图片转PDF失败: %w", err)
	}

	*out = outFile
	return nil
}

// 文件本地化
func localize(url string, name *string) error {
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("无效的URL格式: %s", url)
	}

	ext := extractFileExtension(url)
	if !fileFormats[ext] {
		return fmt.Errorf("不支持的文件格式: %s", ext)
	}

	if err := ensureDir(tempDir); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	localPath := filepath.Join(tempDir, fmt.Sprintf("%s.%s", xid.New().String(), ext))
	if err := download(url, localPath); err != nil {
		return fmt.Errorf("文件下载失败: %w", err)
	}

	*name = localPath
	return nil
}

// 下载文件
func download(url, name string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: time.Second * 30}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求返回非200状态码: %d", resp.StatusCode)
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %w", err)
	}

	if err := os.WriteFile(name, fileData, fileMode); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// 提取文件扩展名
func extractFileExtension(url string) string {
	// [台]经济部能源署链接格式如下，无法基于文本判定文件格式，站点默认pdf格式
	// https://www.moeaea.gov.tw/ecw/populace/content/wHandMenuFile.ashx?file_id=16728
	if strings.Contains(url, "www.moeaea.gov.tw") {
		return "pdf"
	}

	if matches := docTypeRegex.FindStringSubmatch(url); len(matches) >= 2 {
		return strings.ToLower(matches[1])
	}

	if matches := fileExtRegex.FindStringSubmatch(url); len(matches) >= 2 {
		return strings.ToLower(matches[1])
	}

	return ""
}

// 确保目录存在
func ensureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, dirMode)
	}
	return nil
}

// 获取word文件(限定.docx)文本
func readWord(path string, content *string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	res, err := docconv.Convert(f, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true)
	if err != nil {
		return err
	}

	*content = regexp.MustCompile(`\s+`).ReplaceAllString(res.Body, "")
	return nil
}
