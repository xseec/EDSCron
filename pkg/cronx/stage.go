package cronx

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/mathx"
)

// Range 定义数值范围的结构体
type Range struct {
	Min int // 最小值(包含)
	Max int // 最大值(包含)
}

// Contains 检查给定值是否在范围内
func (r Range) Contains(v int) bool {
	if r.Min <= v && v <= r.Max {
		return true
	}
	return false
}

// Stage 计算给定值在范围内的阶段跨度
func (r Range) Stage(n int) int {
	if n < r.Min {
		return 0
	}

	// 因r.Min包含，需要+1
	return mathx.MinInt(r.Max-r.Min, n-r.Min) + 1
}

// adjustBoundarySteps 自适应阶梯电价的边界描述
func adjustBoundarySteps(text, stageSpe string) []string {
	steps := strings.Split(text, stageSpe)

	// 台湾电力阶梯电价的边界描述不够严谨，原文中“120度以下部分”或“1001度以上部分”均应包括自身
	if len(steps) >= 2 {
		first, second, secondLast, last := steps[0], steps[1], steps[len(steps)-2], steps[len(steps)-1]
		firsts := regexp.MustCompile(`(\d+)[^及]*?以下`).FindStringSubmatch(first)
		seconds := regexp.MustCompile(`(\d+)[^\d]+?(\d+)`).FindStringSubmatch(second)
		if len(firsts) == 2 && len(seconds) == 3 {
			first, _ := strconv.Atoi(firsts[1])
			second, _ := strconv.Atoi(seconds[1])
			if first+1 == second {
				steps[0] = strings.ReplaceAll(steps[0], "以下", "及以下")
			}
		}

		lasts := regexp.MustCompile(`(\d+)[^及]*?以上`).FindStringSubmatch(last)
		secondLasts := regexp.MustCompile(`(\d+)[^\d]+?(\d+)`).FindStringSubmatch(secondLast)
		if len(lasts) == 2 && len(secondLasts) == 3 {
			last, _ := strconv.Atoi(lasts[1])
			secondLast, _ := strconv.Atoi(secondLasts[2])
			if last == secondLast+1 {
				steps[len(steps)-1] = strings.ReplaceAll(steps[len(steps)-1], "以上", "及以上")
			}
		}
	}

	return steps
}

// GetStageFee 计算阶梯电价的额外费用(不含首阶)
func GetStageFee(stage string, total float64) float64 {
	steps := adjustBoundarySteps(stage, fieldSubSep)
	valueReg := regexp.MustCompile(`\d+(.\d+)?`)
	var baseValue, stageFee float64
	for _, step := range steps {
		kv := strings.Split(step, fieldKeyValueSep)
		if len(kv) != 2 {
			continue
		}

		key, value := kv[0], kv[1]
		rng, ok := ExtractRange(key)
		if !ok {
			continue
		}

		str := valueReg.FindString(value)
		if len(str) == 0 {
			continue
		}

		stepVal, _ := strconv.ParseFloat(str, 64)

		if rng.Min == math.MinInt {
			baseValue = stepVal
		} else {
			stageFee += (stepVal - baseValue) * float64(rng.Stage(int(total)))
		}
	}
	return stageFee
}

func GetStagePrice(text, stageSpe, valueSep string, total int) (float64, bool) {
	steps := adjustBoundarySteps(text, stageSpe)
	for _, step := range steps {
		values := strings.Split(step, valueSep)
		if len(values) != 2 {
			continue
		}

		rng, ok := ExtractRange(values[0])
		if !ok {
			continue
		}

		if total == 1001 {
			fmt.Println(rng, total, rng.Contains(total))
		}

		if rng.Contains(total) {
			val, err := strconv.ParseFloat(values[1], 64)
			if err != nil {
				continue
			}
			return val, true
		}

	}

	return 0, false
}

// ExtractRange 从自然语言文本中提取数值范围
func ExtractRange(text string) (Range, bool) {
	if len(text) == 0 {
		return Range{}, false
	}

	// 预处理文本：统一符号格式，去除干扰字符
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, " ", "")  // 去除空格
	text = strings.ReplaceAll(text, "，", ",") // 全角逗号转半角
	text = strings.ReplaceAll(text, "－", "-") // 全角连字符转半角
	text = strings.ReplaceAll(text, "～", "~") // 全角波浪线转半角
	text = strings.ReplaceAll(text, ":", "~") // 冒号转波浪线
	text = strings.ReplaceAll(text, "—", "-") // 破折号转连字符

	// 先尝试匹配范围模式（如100-200）
	if r, ok := matchRangePatterns(text); ok {
		return r, true
	}

	// 再尝试匹配单值模式（如100以上）
	if r, ok := matchSingleValuePatterns(text); ok {
		return r, true
	}

	return Range{}, false
}

// matchRangePatterns 匹配各种范围表达式
func matchRangePatterns(text string) (Range, bool) {
	// 定义各种范围模式及其处理函数
	patterns := []struct {
		regex   *regexp.Regexp               // 正则表达式
		handler func([]string) (Range, bool) // 处理函数
	}{
		// 简单范围格式：100-200、100~200、100至200、100到200
		{regexp.MustCompile(`(\d+)(?:-|~|至|到)(\d+)[^\d]*`), handleSimpleRange},
		// 完整范围格式：100以上~200以下
		{regexp.MustCompile(`(\d+)以上(?:-|~|至|到)(\d+)以下[^\d]*`), handleFullRange},
		// 带单位的范围格式：100~200kg
		{regexp.MustCompile(`(\d+)(?:-|~|至|到)(\d+)([^\d]+)`), handleRangeWithUnit},
	}

	// 尝试匹配每种模式
	for _, p := range patterns {
		if matches := p.regex.FindStringSubmatch(text); matches != nil {
			return p.handler(matches)
		}
	}

	return Range{}, false
}

// matchSingleValuePatterns 匹配各种单值表达式
func matchSingleValuePatterns(text string) (Range, bool) {
	// 定义各种单值模式及其处理函数
	patterns := []struct {
		regex   *regexp.Regexp
		handler func([]string) (Range, bool)
	}{
		// 100及以上
		{regexp.MustCompile(`(\d+)[^及以上下]*及以上[^\d]*`), handleAboveInclusive},
		// 100以上
		{regexp.MustCompile(`(\d+)[^及以上下]*以上[^\d]*`), handleAbove},
		// 100及以下
		{regexp.MustCompile(`(\d+)[^及以上下]*及以下[^\d]*`), handleBelowInclusive},
		// 100以下
		{regexp.MustCompile(`(\d+)[^及以上下]*以下[^\d]*`), handleBelow},
		// 超过100
		{regexp.MustCompile(`(?:超过|超過)(\d+)[^\d]*`), handleAbove},
		// 不足100
		{regexp.MustCompile(`不[满足滿](\d+)[^\d]*`), handleBelow},
		// 纯数字
		{regexp.MustCompile(`(\d+)[^\d]*`), handleSingleValue},
	}

	// 尝试匹配每种模式
	for _, p := range patterns {
		if matches := p.regex.FindStringSubmatch(text); matches != nil {
			return p.handler(matches)
		}
	}

	return Range{}, false
}

// handleSimpleRange 处理简单范围格式：100-200
func handleSimpleRange(matches []string) (Range, bool) {
	min, err1 := strconv.Atoi(matches[1])
	max, err2 := strconv.Atoi(matches[2])
	if err1 != nil || err2 != nil {
		return Range{}, false
	}
	return Range{
		Min: min,
		Max: max,
	}, true
}

// handleFullRange 处理完整范围格式：100以上~200以下
func handleFullRange(matches []string) (Range, bool) {
	min, err1 := strconv.Atoi(matches[1])
	max, err2 := strconv.Atoi(matches[2])
	if err1 != nil || err2 != nil {
		return Range{}, false
	}
	return Range{
		Min: min,
		Max: max,
	}, true
}

// handleRangeWithUnit 处理带单位的范围格式：100~200kg
func handleRangeWithUnit(matches []string) (Range, bool) {
	min, err1 := strconv.Atoi(matches[1])
	max, err2 := strconv.Atoi(matches[2])
	if err1 != nil || err2 != nil {
		return Range{}, false
	}
	return Range{
		Min: min,
		Max: max,
	}, true
}

// handleAboveInclusive 处理"大于等于"表达式：100及以上
func handleAboveInclusive(matches []string) (Range, bool) {
	val, err := strconv.Atoi(matches[1])
	if err != nil {
		return Range{}, false
	}
	return Range{
		Min: val,
		Max: math.MaxInt,
	}, true
}

// handleAbove 处理"大于"表达式：100以上、超过100
func handleAbove(matches []string) (Range, bool) {
	val, err := strconv.Atoi(matches[1])
	if err != nil {
		return Range{}, false
	}
	return Range{
		Min: val + 1, // 大于100实际是从101开始
		Max: math.MaxInt,
	}, true
}

// handleBelowInclusive 处理"小于等于"表达式：100及以下
func handleBelowInclusive(matches []string) (Range, bool) {
	val, err := strconv.Atoi(matches[1])
	if err != nil {
		return Range{}, false
	}
	return Range{
		Min: math.MinInt,
		Max: val,
	}, true
}

// handleBelow 处理"小于"表达式：100以下、不足100
func handleBelow(matches []string) (Range, bool) {
	val, err := strconv.Atoi(matches[1])
	if err != nil {
		return Range{}, false
	}
	return Range{
		Min: math.MinInt,
		Max: val - 1, // 小于100实际是到99
	}, true
}

// handleSingleValue 处理纯数字表达式：100
func handleSingleValue(matches []string) (Range, bool) {
	val, err := strconv.Atoi(matches[1])
	if err != nil {
		return Range{}, false
	}
	return Range{
		Min: val,
		Max: val,
	}, true
}
