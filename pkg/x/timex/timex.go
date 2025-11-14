package timex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const twEraOffset = 1911 // 民国元年偏移量（1912年为民国1年）

// SubTomorrow 计算当到明天零时时间差，用于缓存本日
func SubTomorrow() time.Duration {
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return tomorrow.Sub(now)
}

// Add 时间偏移, intervals 顺序: 年, 月, 日, 时, 分, 秒
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

func TwYear(y int) int {
	return y - twEraOffset
}

func Year(twY int) int {
	if twY >= 1000 {
		return twY
	}

	return twY + twEraOffset
}

func NowTwYear() int {
	return TwYear(time.Now().Year())
}

func FormatTwDate(t time.Time, layout string) string {
	twYear := t.Year() - twEraOffset
	if twYear <= 0 {
		twYear = 1 // 处理1911年之前的年份
	}

	// 替换布局中的年份占位符
	layout = strings.Replace(layout, "2006", strconv.Itoa(twYear), 1)
	return t.Format(layout)
}

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

func IsHourInRange(t time.Time, timeRanges string) bool {
	// 处理空字符串
	if timeRanges == "" {
		return false
	}

	timeRanges = regexp.MustCompile(`(\d{2}):(\d{2})`).ReplaceAllString(timeRanges, "$1$2")

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

// MustTime 解析多种时间格式
func MustTime(value string) time.Time {
	if len(value) == 0 {
		return time.Now()
	}

	layouts := []string{
		"2006-1-2 15:4:5",
		"2006-1-2 15:4",
		"2006-1-2 15",
		"2006-1-2",
		"2006-1",
		"2006",

		"20060102150405",
		"20060102",

		"2006年1月2日 15:4:5",
		"2006年1月2日 15:4",
		"2006年1月2日 15时4分5秒",
		"2006年1月2日 15时4分",
		"2006年1月2日 15时",
		"2006年1月2日",
		"2006年1月",
		"2006年",

		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05+08:00",
	}

	value = strings.ReplaceAll(value, "点", "时")
	value = regexp.MustCompile(`0([1-9])([月日时分秒])`).ReplaceAllString(value, "$1$2")
	value = regexp.MustCompile(`\s+`).ReplaceAllString(value, " ")
	// 2006/1/1 → 2006-1-1
	value = regexp.MustCompile(`/(\d+)`).ReplaceAllString(value, "-$1")
	// 民国100~199年：2011年-2110年，已能满足需求
	value = regexp.MustCompile(`^1\d{2}[年\-]`).ReplaceAllStringFunc(value, func(s string) string {
		year, _ := strconv.Atoi(s[:3])
		return fmt.Sprintf("%d%s", year+twEraOffset, s[3:])
	})

	// 尝试解析处理后的value
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return t
		}
	}

	return time.Now()
}

func MustDate(value string) time.Time {
	t := MustTime(value)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func MustMonth(value string) time.Time {
	t := MustTime(value)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
}
