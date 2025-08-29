package timex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func SubTomorrow() time.Duration {
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return tomorrow.Sub(now)
}

func Add(t time.Time, intervals ...int) time.Time {
	result := t
	operations := []func(time.Time, int) time.Time{
		func(t time.Time, v int) time.Time { return t.AddDate(v, 0, 0) },
		func(t time.Time, v int) time.Time { return t.AddDate(0, v, 0) },
		func(t time.Time, v int) time.Time { return t.AddDate(0, 0, v) },
		func(t time.Time, v int) time.Time { return t.Add(time.Duration(v) * time.Hour) },
		func(t time.Time, v int) time.Time { return t.Add(time.Duration(v) * time.Minute) },
		func(t time.Time, v int) time.Time { return t.Add(time.Duration(v) * time.Second) },
	}

	for i, value := range intervals {
		if i >= len(operations) {
			break
		}
		if value != 0 {
			result = operations[i](result, value)
		}
	}

	return result
}

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
	t, err := time.ParseInLocation(layout, adDate, time.Local)
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

// IsDateInRange 检查给定的时间是否在指定的日期范围内
//
// 参数:
//
//	t - 要检查的时间
//	dates - 日期范围字符串，格式为 "MMDD-MMDD,MMDD-MMDD"
//
// 返回值:
//
//	如果时间在范围内，返回 true；否则返回 false
func IsDateInRange(t time.Time, dates string) bool {
	// 处理空字符串
	if dates == "" {
		return false
	}

	currentMMDD := t.Format("0102")

	dateRanges := regexp.MustCompile(`\d{4}`).FindAllString(dates, -1)
	if len(dateRanges)%2 != 0 {
		return false
	}

	for i := 0; i < len(dateRanges); i += 2 {
		start, end := dateRanges[i], dateRanges[i+1]
		if currentMMDD >= start && currentMMDD <= end {
			return true
		}
	}

	return false
}

// IsHourInRange 检查给定的时间是否在指定的时间范围内
//
// 参数:
//
//	t - 要检查的时间
//	timeRanges - 时间范围字符串，格式为 "HHMM-HHMM,HHMM-HHMM"
//
// 返回值:
//
//	如果时间在范围内，返回 true；否则返回 false
func IsHourInRange(t time.Time, timeRanges string) bool {
	// 处理空字符串
	if timeRanges == "" {
		return false
	}

	currentTime := t.Format("1504") // 当前时间的 "HHMM" 格式

	// 处理特殊情况：全天24小时
	if timeRanges == "0000-2400" {
		return true
	}

	// 使用正则表达式提取所有时间点
	timePoints := regexp.MustCompile(`\d{4}`).FindAllString(timeRanges, -1)
	if len(timePoints)%2 != 0 {
		return false // 时间点数量必须是偶数
	}

	for i := 0; i < len(timePoints); i += 2 {
		start, end := timePoints[i], timePoints[i+1]

		// 处理跨天时段（如 2200-0700）
		if start > end {
			// 情况1：当前时间在当天晚上部分（start到2400）
			if currentTime >= start && currentTime <= "2400" {
				return true
			}
			// 情况2：当前时间在次日凌晨部分（0000到end）
			if currentTime >= "0000" && currentTime <= end {
				return true
			}
		} else {
			// 正常时段（不跨天）
			// 包含起始时间，不包含结束时间
			if currentTime >= start && currentTime < end {
				return true
			}
		}
	}

	return false
}
