package timex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TwYear 将公历年份转换为民国年份(台湾地区使用的年份系统)
//
// 参数:
//
//	y - 公历年份
//
// 返回值:
//
//	民国年份(公历年份减去1911)
func TwYear(y int) int {
	return y - 1911
}

// Year 将民国年份转换为公历年份
//
// 参数:
//
//	twY - 民国年份
//
// 返回值:
//
//	公历年份(民国年份加上1911)
func Year(twY int) int {
	if twY >= 1000 {
		return twY
	}

	return twY + 1911
}

// NowTwYear 获取当前的民国年份
//
// 返回值:
//
//	当前时间的民国年份
func NowTwYear() int {
	return TwYear(time.Now().Year())
}

// ParseTwDate 解析民国年份的日期字符串为时间对象
//
// 参数:
//
//	layout - 日期格式布局(如"2006/01/02")
//	value - 包含民国年份的日期字符串(如"113年10月1日")
//
// 返回值:
//
//	时间对象和可能的错误
func ParseTwDate(layout, value string) (time.Time, error) {
	// 使用正则表达式提取民国年份和剩余部分
	re := regexp.MustCompile(`^(\d{3})(.*)$`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return time.Time{}, fmt.Errorf("无效的民国日期格式: %s", value)
	}

	// 转换民国年份为公元年份
	twYear, _ := strconv.Atoi(matches[1])
	adYear := twYear + 1911

	// 构建公元日期字符串
	adDate := fmt.Sprintf("%d%s", adYear, matches[2])

	// 解析日期
	t, err := time.Parse(layout, adDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("日期解析失败: %w", err)
	}

	return t, nil
}

// FormatTwDate 将时间格式化为包含民国年份的字符串（民国年份 = 公历年份 - 1911）
//
// 参数:
//
//	t       - 时间对象
//	layout  - 日期格式布局（如 "2006/01/02"）
//
// 返回值:
//
//	格式化后的日期字符串（民国年份替代公历年份）
func FormatTwDate(t time.Time, layout string) string {
	const twEraOffset = 1911 // 民国元年偏移量（1912年为民国1年）
	twYear := t.Year() - twEraOffset
	if twYear <= 0 {
		twYear = 1 // 处理1911年之前的年份
	}

	// 替换布局中的年份占位符
	layout = strings.Replace(layout, "2006", strconv.Itoa(twYear), 1)
	return t.Format(layout)
}
