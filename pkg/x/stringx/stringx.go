package stringx

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// 替换字符串中所有匹配的子串
//
// 参数：
//   - s: 原始字符串
//   - old: 需要被替换的子串集合
//   - new: 替换后的新字符串
//
// 返回值：替换后的字符串
func Replace(s string, old []string, new string) string {
	for _, o := range old {
		s = strings.ReplaceAll(s, o, new)
	}
	return s
}

// 检查字符串是否包含任意给定子串
//
// 参数：
//   - s: 待检查的字符串
//   - strs: 需要检查的子串集合
//
// 返回值：是否包含任意子串
func ContainsAny(s string, strs ...string) bool {
	for _, str := range strs {
		if strings.Contains(s, str) {
			return true
		}
	}
	return false
}

// 安全格式化字符串，参数不足时循环使用
//
// 参数：
//   - format: 格式化模板字符串
//   - a: 格式化参数集合
//
// 返回值：格式化后的字符串
func MustFormat(format string, a ...interface{}) string {
	reg := regexp.MustCompile("%[^%]")
	count := len(reg.FindAllString(format, -1))
	args := make([]any, 0, count)

	for i := 0; i < count; i++ {
		if len(a) == 0 {
			args = append(args, nil)
			continue
		}
		args = append(args, a[i%len(a)])
	}
	return fmt.Sprintf(format, args...)
}

// 从URL中提取文件名部分
//
// 参数：
//   - u: 完整的URL字符串
//
// 返回值：文件名部分
func FileNameOf(u string) string {
	i := strings.LastIndex(u, "/")
	if i == -1 || i == len(u)-1 {
		return u
	}
	return u[i+1:]
}

// 使用指定分隔符集合分割字符串
//
// 参数：
//   - s: 待分割的字符串
//   - sep: 包含所有可能分隔符的字符串
//
// 返回值：分割后的字符串数组
func Split(s string, sep string) []string {
	results := make([]string, 0)
	start := 0
	ss := []rune(s)

	for i, v := range ss {
		if strings.Contains(sep, string(v)) {
			if i > start {
				results = append(results, string(ss[start:i]))
			}
			start = i + 1
		}
	}

	if start < len(ss) {
		results = append(results, string(ss[start:]))
	}

	return results
}

// 生成随机UUID字符串
//
// 返回值：生成的UUID字符串
func NewUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// 解析数字字符串，支持单个数字和范围格式
//
// 参数：
//   - s: 包含数字或范围的字符串(如"8,9,11-2")
//   - min: 允许的最小值
//   - max: 允许的最大值
//
// 返回值：解析后的整数数组
func MustInts(s string, min, max int64) []int64 {
	results := make([]int64, 0)
	strs := regexp.MustCompile(`[\d-~]+`).FindAllString(s, -1)

	for _, str := range strs {
		subs := regexp.MustCompile(`\d+`).FindAllString(str, -1)
		switch len(subs) {
		case 1:
			v, _ := strconv.ParseInt(subs[0], 10, 64)
			results = append(results, v)
		case 2:
			start, _ := strconv.ParseInt(subs[0], 10, 64)
			end, _ := strconv.ParseInt(subs[1], 10, 64)

			for {
				results = append(results, start)
				if start == end {
					break
				}

				if start == max {
					start = min
				} else {
					start++
				}
			}
		}
	}

	return results
}

// PadLeft 在字符串左侧填充指定字符直到达到指定长度
//
// 参数：
//   - s: 原始字符串
//   - padChar: 用于填充的字符（如果是多字符字符串，只使用第一个字符）
//   - totalWidth: 填充后字符串的总长度
//
// 返回值：
//   - 填充后的字符串
//   - 当 totalWidth <= len(s) 时直接返回原字符串
//   - 当 padChar 为空字符串时使用空格填充
func PadLeft(s string, padChar string, totalWidth int) string {
	if len(s) >= totalWidth {
		return s
	}

	// 确定填充字符
	char := ' '
	if len(padChar) > 0 {
		char = rune(padChar[0]) // 只取第一个字符
	}

	// 计算需要填充的长度
	paddingLength := totalWidth - len(s)
	padding := strings.Repeat(string(char), paddingLength)

	return padding + s
}
