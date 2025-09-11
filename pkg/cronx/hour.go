package cronx

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"seeccloud.com/edscron/pkg/x/slicex"
)

// GetHalfHourIndexs 将时段字符串转换为48点时段索引
// 参数s格式如"1200-1300,2300-0130"，返回[24,25,46,47,0,1,2]
func GetHalfHourIndexs(s string) []int {
	indexs := make([]int, 0)
	items := strings.Split(s, ",")

	for _, item := range items {
		reg := regexp.MustCompile(`(\d{2})(\d{2})-(\d{2})(\d{2})`)
		subs := reg.FindStringSubmatch(item)
		if len(subs) != 5 {
			continue
		}

		// 转换时间点为48点制索引
		nums := slicex.MapFunc(subs[1:], func(sub string) int {
			num, _ := strconv.Atoi(sub)
			return num
		})

		from := nums[0]*2 + nums[1]/30 // 起始点索引
		to := nums[2]*2 + nums[3]/30   // 结束点索引

		// 生成连续的时段索引
		for i := from; i != to; {
			indexs = append(indexs, i)
			if i == 47 { // 处理跨天情况
				i = 0
			} else {
				i++
			}
		}
	}

	return indexs
}

// getDuration 计算时段字符串的总秒数
// 参数s格式如"1200-1300,2200-0800"表示12:00-13:00和22:00-08:00
func getDuration(s string) (int, error) {
	s = strings.ReplaceAll(s, "2400", "0000") // 标准化时间格式
	durs := regexp.MustCompile(`(\d{4})\s*[~-]+\s*(\d{4})`).FindAllStringSubmatch(s, -1)
	total := 0

	for _, dur := range durs {
		// 解析时间点
		t1, err := time.ParseInLocation("1504", dur[1], time.Local)
		if err != nil {
			return 0, fmt.Errorf("解析起始时间失败: %w", err)
		}

		t2, err := time.ParseInLocation("1504", dur[2], time.Local)
		if err != nil {
			return 0, fmt.Errorf("解析结束时间失败: %w", err)
		}

		// 处理跨天情况
		if t2.Unix() <= t1.Unix() {
			t2 = t2.AddDate(0, 0, 1)
		}

		total += int(t2.Unix()) - int(t1.Unix())
	}

	return total, nil
}
