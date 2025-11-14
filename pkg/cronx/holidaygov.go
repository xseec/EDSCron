package cronx

// 数据源：https://www.gov.cn/

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

	"seeccloud.com/edscron/pkg/chromedpx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"
	"seeccloud.com/edscron/pkg/x/slicex"
)

type HolidayCategory string

const (
	HolidayOff     HolidayCategory = "假日"
	HolidayOn      HolidayCategory = "调休工作日"
	HolidayPeakOff HolidayCategory = "离峰日"
	HolidayNull    HolidayCategory = ""
)

type HolidayGovConfig struct {
	DP   chromedpx.DP `json:"dp"`
	Year int          `json:"year"`
}

type Holiday struct {
	Area     string `json:"area"`     // 国家或地区代码
	Alias    string `json:"alias"`    // 国家或地区名称
	Date     string `json:"date"`     // 日期，格式：YYYY-MM-DD
	Category string `json:"category"` // 日期类型：假日/调休工作日
	Detail   string `json:"detail"`   // 节假日名称详情
}

func DefaultHolidayGovTask() string {
	url := "https://sousuo.www.gov.cn/sousuo/search.shtml?code=17da70961a7&searchWord=2006年部分节假日"
	title := "国务院办公厅关于2006年部分节假日安排的通知"
	cfg := HolidayGovConfig{
		DP: chromedpx.DP{
			Urls: []chromedpx.DPUrl{
				{
					Url:    url,
					Clicks: []string{title},
				},
			},
			Outer: chromedpx.DPOuter{
				Selector: "div#UCAP-CONTENT",
				OnlyText: true,
			},
		},
		Year: 2006,
	}

	task, _ := json.Marshal(cfg)
	return string(task)
}

func (h HolidayGovConfig) Run(m *MailConfig) (*[]Holiday, error) {
	html := ""
	if err := h.DP.Run(context.Background(), &html); err != nil {
		return nil, err
	}

	days := matchHoliday(html, h.Year)

	// 发送邮件通知
	if m != nil {
		elems := slicex.MapFunc(days, func(h Holiday) string {
			return fmt.Sprintf("<li>%s %s %s</li>", h.Date, h.Category, h.Detail)
		})
		details := "<ul>\n" + strings.Join(elems, "") + "</ul>"
		go m.Send(&HolidayNoticeTemplate{
			Area:    ChinaAreaName,
			Year:    h.Year,
			Details: template.HTML(details),
		})

	}

	return emptyValueErr(m, h, &days)
}

func matchHoliday(html string, year int) []Holiday {
	hols := []Holiday{}
	// 举例：一、元旦：1月1日至3日放假调休，共3天。1月4日（星期日）上班。
	html = regexp.MustCompile(`（[^）]+）`).ReplaceAllString(html, "")
	itemReg := regexp.MustCompile(`(?m)[一二三四五六七八九十]、.*?$(?:\n|$)`)
	holReg := regexp.MustCompile(`、([^：]+)：`)
	dateReg := regexp.MustCompile(`(?:(\d{1,2})月)?(\d{1,2})日`)
	items := itemReg.FindAllString(html, -1)
	for _, item := range items {
		// (1) 假日名称
		subs := holReg.FindStringSubmatch(item)
		if len(subs) != 2 {
			continue
		}
		holName := subs[1]

		// (2) 放假范围：连续日期
		subItems := strings.Split(item, "。")
		dates := dateReg.FindAllStringSubmatch(subItems[0], -1)
		if len(dates) == 0 {
			continue
		}

		startDate := parseDate(year, dates[0][1], dates[0][2])
		endDate := startDate
		if len(dates) == 2 {
			monthStr := expx.If(dates[1][1] == "", dates[0][1], dates[1][1])
			endDate = parseDate(year, monthStr, dates[1][2])
		}

		current := startDate
		for !current.After(endDate) {
			hols = append(hols, Holiday{
				Area:     string(ChinaArea),
				Alias:    ChinaAreaName,
				Date:     current.Format(vars.DateFormat),
				Category: string(HolidayOff),
				Detail:   holName,
			})
			current = current.AddDate(0, 0, 1)
		}

		// (3) 调休范围：离散日期
		if len(subItems) < 2 {
			continue
		}

		dates = dateReg.FindAllStringSubmatch(subItems[1], -1)
		for _, subs := range dates {
			date := parseDate(year, subs[1], subs[2])
			hols = append(hols, Holiday{
				Area:     string(ChinaArea),
				Alias:    ChinaAreaName,
				Date:     date.Format(vars.DateFormat),
				Category: string(HolidayOn),
			})
		}
	}

	return hols
}

func parseDate(year int, month string, day string) time.Time {
	// 节假日中涉及12月的皆为上一年
	if month == "12" {
		year--
	}

	monthInt, _ := strconv.Atoi(month)
	dayInt, _ := strconv.Atoi(day)
	return time.Date(year, time.Month(monthInt), dayInt, 0, 0, 0, 0, time.Local)
}
