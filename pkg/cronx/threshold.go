package cronx

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"strings"
)

// Thresholding 去除图片底纹功能
//
// 参数:
//   - name: 本地图片路径
//   - th: 底纹阈值(0-255)，建议值230以上
//
// 返回值:
//   - error: 处理过程中出现的错误
//
// 功能说明:
//  1. 读取图片并转换为灰度图
//  2. 根据阈值将像素二值化
//  3. 保存处理后的图片
//  4. 支持PNG和JPEG格式
func thresholding(name string, th int) error {
	// 检查阈值有效性
	if th <= 0 {
		return nil // 阈值为0或负数时不做处理
	}

	// 获取文件扩展名并转换为小写
	ext := strings.ToLower(path.Ext(name))

	// 跳过PDF文件处理
	if ext == ".pdf" {
		return nil
	}

	// 打开图片文件
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	// 根据文件类型解码图片
	var originalImage image.Image
	switch ext {
	case ".png":
		originalImage, err = png.Decode(file)
	default: // 默认处理JPEG格式
		originalImage, err = jpeg.Decode(file)
	}
	if err != nil {
		return err
	}

	// 创建二值化图像
	binaryImage := image.NewGray(originalImage.Bounds())
	threshold := uint8(th)
	bounds := originalImage.Bounds()

	// 遍历每个像素进行二值化处理
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 获取原始像素并转换为灰度
			oldPixel := originalImage.At(x, y)
			grayPixel := color.GrayModel.Convert(oldPixel).(color.Gray)

			// 根据阈值设置新像素值
			newPixel := color.Gray{Y: 0} // 默认黑色
			if grayPixel.Y > threshold {
				newPixel.Y = 255 // 超过阈值设为白色
			}
			binaryImage.Set(x, y, newPixel)
		}
	}

	// 创建输出文件
	outputFile, err := os.Create(name)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// 根据格式编码保存图片
	switch ext {
	case ".png":
		return png.Encode(outputFile, binaryImage)
	default:
		return jpeg.Encode(outputFile, binaryImage, &jpeg.Options{Quality: 100})
	}
}
