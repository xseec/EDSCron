package cronx

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"seeccloud.com/edscron/pkg/x/timex"
)

type SummerRange string

var (
	// 夏月常带日期范围，如：“夏月(5月16日至10月15日)”
	summerRangeReg = regexp.MustCompile(`^夏.*(\d+)月(\d+)日至(\d+)月(\d+)日`)
)

// formatTwHour 格式化台湾电价时间范围字符串，输出如：“00:00-24:00”
func formatTwHour(s string) string {
	if s == "全日" {
		return "0000-2400"
	}

	subs := regexp.MustCompile(`\d{2}`).FindAllString(s, -1)
	if len(subs)%4 != 0 {
		return ""
	}

	s = ""
	for i := 0; i < len(subs); i += 4 {
		s = s + "," + fmt.Sprintf("%s%s-%s%s", subs[i], subs[i+1], subs[i+2], subs[i+3])
	}

	return strings.Trim(s, ",")
}

// MustRange 从季节范围字符串解析出夏季和非夏季日期范围
func (s SummerRange) MustRange() (summer, nonSummer string) {
	if len(s) == 0 {
		return
	}

	matches := summerRangeReg.FindStringSubmatch(string(s))
	if len(matches) != 5 {
		return
	}

	// 解析夏季开始和结束日期
	startMonth, _ := strconv.Atoi(matches[1])
	startDay, _ := strconv.Atoi(matches[2])
	endMonth, _ := strconv.Atoi(matches[3])
	endDay, _ := strconv.Atoi(matches[4])

	// 构造 time.Time 对象（年份用固定值，如 2000，仅用于计算）
	startDate := time.Date(2000, time.Month(startMonth), startDay, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2000, time.Month(endMonth), endDay, 0, 0, 0, 0, time.UTC)

	// 格式化夏季范围
	summer = fmt.Sprintf("%02d%02d-%02d%02d", startMonth, startDay, endMonth, endDay)

	// 计算非夏季范围（分两段）
	var nonSummerRanges []string

	// 第一段：1月1日 ~ 夏季开始前一日
	if !(startMonth == 1 && startDay == 1) {
		prevDay := startDate.AddDate(0, 0, -1) // 计算前一天
		nonSummerRanges = append(nonSummerRanges,
			fmt.Sprintf("0101-%02d%02d", int(prevDay.Month()), prevDay.Day()))
	}

	// 第二段：夏季结束次日 ~ 12月31日
	if !(endMonth == 12 && endDay == 31) {
		nextDay := endDate.AddDate(0, 0, 1) // 计算后一天
		nonSummerRanges = append(nonSummerRanges,
			fmt.Sprintf("%02d%02d-1231", int(nextDay.Month()), nextDay.Day()))
	}

	nonSummer = strings.Join(nonSummerRanges, ",")
	return
}

// InSummer 判断给定时间是否在夏季范围内
func (s SummerRange) InSummer(t time.Time) bool {
	summer, _ := s.MustRange()
	if summer == "" {
		return false
	}

	// 分割夏季范围（如 "0601-0930"）
	parts := strings.Split(summer, "-")
	if len(parts) != 2 {
		return false
	}
	start, end := parts[0], parts[1]

	// 获取当前日期的 "MMDD" 格式字符串
	current := fmt.Sprintf("%02d%02d", int(t.Month()), t.Day())

	// 直接比较字符串（"MMDD" 格式可正确比较大小）
	return current >= start && current <= end
}

// mustAdjustTwYear 调整结构体中的年份字段为民国年份
func mustAdjustTwYear(v any, year *int) {
	buf, err := json.Marshal(v)
	if err != nil {
		return
	}

	buf = regexp.MustCompile(`\d{4}年`).ReplaceAllFunc(buf, func(match []byte) []byte {
		*year, _ = strconv.Atoi(string(match[:4]))
		return []byte(fmt.Sprintf("%d年", timex.TwYear(*year)))
	})

	json.Unmarshal(buf, v)
}
