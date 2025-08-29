package cronx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
)

const (
	// 国家温室气体排放因子数据库API
	carbonFactorAPI = "https://data.ncsc.org.cn/factoryes/api/factor/metaData/getFactorTables?pkid=82"
)

var (
	// 特殊区域名称映射
	specialAreaMapping = map[string]string{
		"全国电力平均二氧化碳排放因子":                   "中国",
		"全国电力平均二氧化碳排放因子（不包括市场化交易的非化石能源电量）": "中国非化石能源",
		"全国化石能源电力二氧化碳排放因子":                 "中国化石能源",
	}

	// 年份匹配正则
	yearRegex = regexp.MustCompile(`(\d{4})年`)
)

// CarbonFactor 表示一个碳排放因子记录
type CarbonFactor struct {
	Area  string  `json:"area"`  // 区域名称（如"全国"、"华东"、"福建省"）
	Year  int64   `json:"year"`  // 年份（如2020）
	Value float64 `json:"value"` // 净购入电力碳排放因子（单位：kgCO₂/kWh）
}

// CarbonConfig 碳排放因子获取配置
type CarbonConfig struct{}

// Run 获取历年全国、区域和省份的净购入电力碳排放因子
//
// 参数:
//   - m: 邮件配置（用于错误处理）
//
// 返回:
//   - []CarbonFactor: 碳排放因子列表
//   - error: 错误信息
func (c CarbonConfig) Run(m *MailConfig) (*[]CarbonFactor, error) {
	// 创建HTTP客户端并设置超时
	client := &http.Client{Timeout: 30 * time.Second}

	// 发送HTTP请求获取数据
	resp, err := client.Get(carbonFactorAPI)
	if err != nil {
		return nil, fmt.Errorf("请求碳排放因子API失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回非200状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取API响应失败: %w", err)
	}

	// 解析JSON数据
	js, err := simplejson.NewJson(body)
	if err != nil {
		return nil, fmt.Errorf("解析JSON数据失败: %w", err)
	}

	// 提取结果数据
	resultData, err := js.Get("data").Get("result").MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("提取结果数据失败: %w", err)
	}

	// 反序列化为CarbonTable数组
	var tables []struct {
		Result [][]struct {
			Value string `json:"value"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resultData, &tables); err != nil {
		return nil, fmt.Errorf("反序列化结果数据失败: %w", err)
	}

	// 处理数据并构建碳排放因子列表
	factors, err := processCarbonTables(tables)
	if err != nil {
		return nil, err
	}

	// 处理空值错误并返回结果
	return emptyValueErr(m, c, &factors)
}

// processCarbonTables 处理碳排放因子表格数据
func processCarbonTables(tables []struct {
	Result [][]struct {
		Value string `json:"value"`
	} `json:"result"`
}) ([]CarbonFactor, error) {
	var factors []CarbonFactor

	for _, table := range tables {
		if len(table.Result) == 0 {
			continue
		}

		// 提取年份信息
		years := extractYears(table.Result[0])
		if len(years) == 0 {
			continue
		}

		// 处理每一行数据
		for i := 1; i < len(table.Result); i++ {
			row := table.Result[i]
			if len(row) <= 1 {
				continue
			}

			// 获取区域名称（应用特殊映射）
			area := getMappedArea(row[0].Value)

			// 处理每个年份的数据
			for j := 1; j < len(row) && j-1 < len(years); j++ {
				value, err := strconv.ParseFloat(row[j].Value, 64)
				if err != nil {
					continue // 跳过无效数值
				}

				factors = append(factors, CarbonFactor{
					Area:  area,
					Year:  years[j-1],
					Value: value,
				})
			}
		}
	}

	return factors, nil
}

// extractYears 从表头行提取年份信息
func extractYears(header []struct {
	Value string `json:"value"`
}) []int64 {
	var years []int64

	for _, col := range header[1:] { // 跳过第一列（名称列）
		matches := yearRegex.FindStringSubmatch(col.Value)
		if len(matches) == 2 {
			if year, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				years = append(years, year)
			}
		}
	}

	return years
}

// getMappedArea 获取映射后的区域名称
func getMappedArea(original string) string {
	if mapped, ok := specialAreaMapping[original]; ok {
		return mapped
	}
	return original
}
