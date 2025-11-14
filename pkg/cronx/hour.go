package cronx

import (
	"regexp"
	"strconv"
	"strings"

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
