package cronx

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"seeccloud.com/edscron/pkg/chromedpx"
	"seeccloud.com/edscron/pkg/x/expx"
)

var (
	// 南方电网：下辖五省
	NanfangfProvinces = map[string]string{
		"深圳": "sz",
		"广东": "gd",
		"广西": "gx",
		"云南": "yn",
		"贵州": "gz",
		"海南": "hn",
	}
	yunnan   = "云南"
	shenzhen = "深圳"

	dlgdTableTitleFormat = `%s[\p{Han}（）\(\)]{1,15}代理购电工商业用户电价(?:公告)?表%s`
	specialCities        = map[string]string{
		"广州/珠海/佛山/中山/东莞": "（珠三角五市）",
		"惠州":             "（惠州市）",
		"江门":             "（江门市）",
		"汕头/潮州/揭阳/汕尾/阳江/湛江/茂名/肇庆/恩平/台山/开平": "（东西两翼地区）",
		"云浮/河源/梅州/韶关/清远":                   "（粤北山区）",
		"深圳":                               "（深圳市）",
		"深汕特别合作区":                          "（深汕特别合作区）",
		"西安/宝鸡/咸阳/铜川/渭南/延安/汉中/安康/商洛/韩城/华阴": "（不含榆林地区）",
		"榆林/神木": "（榆林地区）",
		"石家庄/保定/邢台/邯郸/沧州/衡水":                 "", //河北南网
		"唐山/秦皇岛/廊坊/承德/张家口":                   "", //冀北电网
		"通辽/兴安/呼伦贝尔/赤峰":                      "", //蒙东电网
		"呼和浩特/包头/乌海/乌兰察布/鄂尔多斯/巴彦淖尔/锡林郭勒/阿拉善": "",
	}

	nanfangfDlgdUrlFormat = "https://95598.csg.cn/#/%s/businessPage"

	// 蒙西电网：下辖八市，暨内蒙电网
	neimengxiCities = []string{"呼和浩特", "包头", "乌海", "乌兰察布", "鄂尔多斯", "巴彦淖尔", "锡林郭勒", "阿拉善"}
	// 蒙东电网：下辖四市，隶属国家电网
	neimengdongCities = []string{"通辽", "兴安", "呼伦贝尔", "赤峰"}
	neimengDlgdUrl    = "https://www.impc.com.cn/node_7810.html"

	// 国家电网
	guojiaHost          = "https://95598.cn"
	guojiaCitySelectUrl = "https://95598.cn/osgweb/index"
	guojiaDlgdUrl       = "https://95598.cn/osgweb/ipElectrovalenceStandard"

	// 冀北电网
	jibeiCities = []string{"唐山", "秦皇岛", "廊坊", "承德", "张家口"}
	// 河北南网
	hebeinanCities = []string{"石家庄", "保定", "邢台", "邯郸", "沧州", "衡水"}
	hebeinanTitle  = `河北南网代理购电转供电主体的终端用户[\(（]工商业[\)）]电价表`
	// 榆林地区
	yulinCities = []string{"榆林", "神木"}

	// 西藏电网：工商业电价固定
	xizang = "西藏"

	// 两省电价和最高气温挂钩
	CapitalWeather = map[string]any{
		"广东": nil,
		"四川": nil,
	}
)

// GetDefaultCity 返回区域非省会默认城市，如"秦皇岛"使用"唐山"
func GetDefaultCity(cty string) string {
	if slices.Contains(yulinCities, cty) {
		return yulinCities[0]
	}

	if slices.Contains(jibeiCities, cty) {
		return jibeiCities[0]
	}

	if slices.Contains(neimengdongCities, cty) {
		return neimengdongCities[0]
	}

	return ""
}

// DlgdConfig 代理购电任务配置
type DlgdConfig struct {
	Area      string       `json:"area"`      // 区域名称
	Month     string       `json:"month"`     // 执行月份，格式"2025年3月"
	Dp        chromedpx.DP `json:"dp"`        // 网页爬虫配置
	TitlePat  string       `json:"titlePat"`  // 电价表标题正则表达式
	Threshold int          `json:"threshold"` // 图片底纹阈值(230~245)
	Ocr       AliOcr       `json:"ocr"`       // 阿里云OCR配置
}

type MiniDlgdConfig struct {
	Province string `json:"province"` // 省
	City     string `json:"city"`     // 城
	Area     string `json:"area"`     // 区
	Month    string `json:"month"`    // 月，格式"2006年1月"
}

func NewDlgdConfig(mini MiniDlgdConfig, ocr AliOcr) DlgdConfig {
	var cfg DlgdConfig
	cfg.Ocr = ocr
	cfg.Area = mini.Province
	cfg.Dp.DownloadDir = fmt.Sprintf("%s/%s", tempDir, time.Now().Format("20060102150405"))

	cty := Shorten(mini.City)
	sepicalTitle := ""
	for k, v := range specialCities {
		if strings.Contains(k, cty) {
			sepicalTitle = v
			cfg.Area = k
			break
		}
	}

	if slices.Contains(hebeinanCities, cty) {
		cfg.TitlePat = hebeinanTitle
	} else {
		prefix := expx.If(cty == shenzhen, cty, mini.Province)
		cfg.TitlePat = fmt.Sprintf(dlgdTableTitleFormat, prefix, sepicalTitle)
	}

	// "2025-01" → "2025年01月|2025年1月"
	monthPat := mini.Month
	subs := regexp.MustCompile(`[1-9]\d*`).FindAllString(mini.Month, -1)
	if len(subs) == 2 {
		monthPat = fmt.Sprintf("%s年%02s月|%s年%s月", subs[0], subs[1], subs[0], subs[1])
		cfg.Month = fmt.Sprintf("%s年%s月", subs[0], subs[1])
	}

	// 南方电网
	if p, ok := NanfangfProvinces[mini.Province]; ok {
		code := p
		if p, ok = NanfangfProvinces[mini.City]; ok {
			code = p
		}

		url := fmt.Sprintf(nanfangfDlgdUrlFormat, code)
		cfg.Dp.Urls = []chromedpx.DPUrl{
			{
				Url:    url,
				Clicks: []string{"信息公开", "电价及收费标准", "+代理购电-" + monthPat},
			},
		}
		if mini.Province == yunnan {
			cfg.Dp.Outer = chromedpx.DPOuter{
				Selector: `img[src^="https://95598.csg.cn/ucs/sr/info/"]`,
				Pattern:  `src="([^"]*)"`,
			}
		} else {
			cfg.Dp.Outer = chromedpx.DPOuter{
				Selector: `a[href*="documentType=pdf"]`,
				Pattern:  `href="([^"]*)"`,
			}
		}

		return cfg
	}

	// 内蒙电网
	if slices.Contains(neimengxiCities, cty) {
		cfg.Dp.Urls = []chromedpx.DPUrl{
			{
				Url:    neimengDlgdUrl,
				Clicks: []string{"主动公开信息", "供电企业电价和收费标准", "+代理购电-" + monthPat},
			},
		}
		cfg.Dp.Outer = chromedpx.DPOuter{
			Selector: `a[href$=".pdf"]`,
			Pattern:  `href="([^"]*)"`,
		}

		return cfg
	}

	// 国家电网
	cfg.Dp.Urls = []chromedpx.DPUrl{
		{
			Url:    guojiaCitySelectUrl,
			Clicks: []string{"*确认", "city_select", mini.Province}, // 如站点升级公告弹出框
		},
		{
			Url:    guojiaDlgdUrl,
			Clicks: []string{"*确认", mini.City, mini.Area, "代理购电", "+代理购电-" + monthPat},
		},
	}

	if mini.Province == xizang {
		// 西藏电网
		cfg.Dp.Urls[1].Clicks = []string{"*确认", mini.City, mini.Area, "工商业用电", "工商业用电价格表"}
	}

	cfg.Dp.Outer = chromedpx.DPOuter{
		Selector: `img[src^="/omg-static/"]`,
		Pattern:  `src="([^"]*)"`,
		Host:     guojiaHost,
	}

	return cfg
}
