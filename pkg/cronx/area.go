package cronx

import (
	"regexp"
	"strings"

	"github.com/siongui/gojianfan"
)

type AreaCategory string

const (
	TaiwanArea AreaCategory = "taiwan"
	ChinaArea  AreaCategory = "china"
)

// EitherChinaOrTaiwan 根据地址判断所属区域
//
// 参数:
//   - address: 用户地址(支持简繁体)，如"福建省厦门市集美区"或"台北市中山北路"
//
// 返回:
//   - "china": 中国大陆地址
//   - "taiwan": 台湾地区地址
//
// 注意: 目前仅支持中国大陆和台湾地区的识别
func EitherChinaOrTaiwan(address string) AreaCategory {
	province, _ := ExtractAddress(address, true)
	if strings.Contains(address, taiwanAreaName) || province == taiwanAreaName {
		return TaiwanArea
	}

	return ChinaArea
}

// ExtractAddress 从地址中提取省市级信息
//
// 参数:
//   - address: 用户地址(支持简繁体)
//   - short: 是否返回精简名称(去除"省"、"市"等后缀)
//
// 返回:
//   - province: 省级行政区名称
//   - city: 市级行政区名称
func ExtractAddress(address string, short bool) (province, city string) {
	defer func() {
		// 泉州市石狮市 → 泉州市
		cty := regexp.MustCompile(`\S+?(市|自治州|盟|地区|区|县)`).FindString(city)
		if len(cty) > 0 {
			city = cty
		}

		if short {
			// 内蒙古自治区→内蒙（95598.cn缩写）
			province = trimSuffix(province, []string{"省", "壮族自治区",
				"回族自治区", "维吾尔自治区", "特别行政区", "古自治区", "自治区", "市"})
			city = trimSuffix(city, []string{"市", "自治州", "盟", "地区", "区", "县"})
		}
	}()

	// 繁体转简体
	address = gojianfan.T2S(address)

	// 定义正则表达式模式
	const (
		provincePattern = `(北京市|天津市|上海市|重庆市|台湾|内蒙古自治区|广西壮族自治区|` +
			`西藏自治区|宁夏回族自治区|新疆维吾尔自治区|香港特别行政区|澳门特别行政区|.+?省)`
		cityPattern   = `(.+?市|.+?自治州|.+?盟|.+?地区)` //|.+?区|.+?县)`
		taiwanPattern = `^(台湾地区|台湾省|台湾)?(台北市|高雄市|台中市|台南市|` +
			`新北市|新竹市|桃园市|基隆市|嘉义市|.+?县)`
	)

	// 尝试匹配标准地址格式(省级+市级)
	if matches := regexp.MustCompile(provincePattern + cityPattern).FindStringSubmatch(address); len(matches) == 3 {
		province, city = matches[1], matches[2]
		return
	}

	// 尝试匹配台湾地区地址格式
	if matches := regexp.MustCompile(taiwanPattern).FindStringSubmatch(address); len(matches) == 3 {
		province, city = taiwanAreaName, matches[2]
		return
	}

	// 尝试仅匹配省级(如直辖市/特别行政区)
	if matches := regexp.MustCompile(provincePattern).FindStringSubmatch(address); len(matches) == 2 {
		province, city = matches[1], matches[1]
	}

	// 尝试仅匹配市级
	if matches := regexp.MustCompile(cityPattern).FindStringSubmatch(address); len(matches) == 2 {
		city = matches[1]
	}

	return
}

// trimSuffix 去除字符串中的指定后缀
func trimSuffix(s string, suffixes []string) string {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return strings.TrimSuffix(s, suffix)
		}
	}
	return s
}
