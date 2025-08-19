package cronx

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"seeccloud.com/edscron/pkg/x/slicex"

	pdf "github.com/pdfcpu/pdfcpu/pkg/api"
)

const (
	jpegQuality = 100 // JPEG 输出质量
)

// crop 裁剪文档，支持PDF和图片格式(JPG/JPEG/PNG)
//
// 参数:
//   - input: 输入文件路径
//   - output: 输出文件路径
//   - crops: 裁剪参数
//     PDF文档: 页面范围参数(从1开始)，如["1"]表示第一页
//     图片文档: 4个整数参数，表示基于原始图片尺寸的百分比，格式为["x1", "y1", "x2", "y2"]
//     对应裁剪矩形对角点: pt1(x1*wid/100, y1*hei/100), pt2(x2*wid/100, y2*hei/100)
//
// 返回:
//   - error: 裁剪过程中的错误
func crop(input, output string, crops []string) error {
	if len(crops) == 0 {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(input))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return cropImage(input, output, crops)
	default:
		return pdf.TrimFile(input, output, crops, nil)
	}
}

// cropImage 裁剪图片文件
func cropImage(input, output string, crops []string) error {
	if len(crops) != 4 {
		return nil
	}

	// 打开图片文件
	file, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("打开图片文件失败: %w", err)
	}
	defer file.Close()

	// 解码图片
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %w", err)
	}

	// 创建新图片并复制内容
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)

	// 转换裁剪参数为整数
	cropParams := slicex.MapFunc(crops, func(s string) int {
		val, _ := strconv.Atoi(s)
		return val
	})

	// 计算裁剪区域
	rect := image.Rect(
		bounds.Dx()*cropParams[0]/100,
		bounds.Dy()*cropParams[1]/100,
		bounds.Dx()*cropParams[2]/100,
		bounds.Dy()*cropParams[3]/100,
	)

	// 执行裁剪
	subImg := newImg.SubImage(rect)

	// 创建输出文件
	outFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)

	// 根据文件类型编码保存
	ext := strings.ToLower(filepath.Ext(output))
	switch ext {
	case ".png":
		err = png.Encode(writer, subImg)
	default:
		err = jpeg.Encode(writer, subImg, &jpeg.Options{Quality: jpegQuality})
	}

	if err != nil {
		return fmt.Errorf("图片编码失败: %w", err)
	}

	// 确保数据写入磁盘
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
