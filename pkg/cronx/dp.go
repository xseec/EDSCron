package cronx

import (
	"fmt"
	"slices"
	"time"

	"seeccloud.com/edscron/pkg/chromedpx"
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

	nanfangfDlgdUrlFormat = "https://95598.csg.cn/#/%s/businessPage"

	// 蒙西电网：下辖八市，暨内蒙电网
	NeimengxiCities = []string{"呼和浩特市", "包头市", "乌海市", "乌兰察布市", "鄂尔多斯市", "巴彦淖尔市", "锡林郭勒盟", "阿拉善盟"}
	// 蒙东电网：下辖四市，隶属国家电网
	NeimengdongCities = []string{"通辽", "兴安", "呼伦贝尔", "赤峰"}
	neimengDlgdUrl    = "https://www.impc.com.cn/node_7810.html"

	// 国家电网
	guojiaHost          = "https://95598.cn"
	guojiaCitySelectUrl = "https://95598.cn/osgweb/index"
	guojiaDlgdUrl       = "https://95598.cn/osgweb/ipElectrovalenceStandard"

	// 西藏电网：工商业电价固定
	xizang = "西藏"
)

type Region struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Area     string `json:"area"`
}

func NewDlgdDp(province, city, area string, month time.Time) chromedpx.DP {
	dp := chromedpx.DP{}
	monthPat := fmt.Sprintf("%s|%s", month.Format("2006年01月"), month.Format("2006年1月"))

	// 南方电网
	if p, ok := NanfangfProvinces[province]; ok {
		code := p
		if p, ok = NanfangfProvinces[city]; ok {
			code = p
		}

		url := fmt.Sprintf(nanfangfDlgdUrlFormat, code)
		dp.Urls = []chromedpx.DPUrl{
			{
				Url:    url,
				Clicks: []string{"信息公开", "电价及收费标准", "+代理购电-" + monthPat, "*documentType=pdf"},
			},
		}
		dp.Outer = chromedpx.DPOuter{
			Selector: `img[src^="https://95598.csg.cn/ucs/sr/info/"]`,
			Pattern:  `src="([^"]*)"`,
		}

		return dp
	}

	// 内蒙电网
	if slices.Contains(NeimengxiCities, city) {
		dp.Urls = []chromedpx.DPUrl{
			{
				Url:    neimengDlgdUrl,
				Clicks: []string{"主动公开信息", "供电企业电价和收费标准", "+代理购电-" + monthPat},
			},
		}
		dp.Outer = chromedpx.DPOuter{
			Selector: `a[href$=".pdf"]`,
			Pattern:  `href="([^"]*)"`,
		}

		return dp
	}

	// 国家电网
	dp.Urls = []chromedpx.DPUrl{
		{
			Url:    guojiaCitySelectUrl,
			Clicks: []string{"city_select", province},
		},
		{
			Url:    guojiaDlgdUrl,
			Clicks: []string{"*确认", city, area, "代理购电", "+代理购电-" + monthPat},
		},
	}

	if province == xizang {
		// 西藏电网
		dp.Urls[1].Clicks = []string{"*确认", city, area, "工商业用电", "工商业用电价格表"}
	}
	dp.Outer = chromedpx.DPOuter{
		Selector: `img[src^="/omg-static/"]`,
		Pattern:  `src="([^"]*)"`,
		Host:     guojiaHost,
	}

	return dp
}
