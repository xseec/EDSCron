package cronx

// [废弃]获取不同国家或地区的节假日
// 台湾地区有专属电价日历表，获取其法定节假日无意义
// 因此，仅限大陆地区的节假日用HolidayGovConfig更佳

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"regexp"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/PuerkitoBio/goquery"
// 	"seeccloud.com/edscron/pkg/x/slicex"
// )

// // HolidayConfig 节假日配置结构体
// type HolidayConfig struct {
// 	Area      string   `json:"area"`      // 国家或地区代码，如：china/taiwan
// 	Alias     string   `json:"alias"`     // 国家或地区名称，如：中国/台湾
// 	Year      int      `json:"year"`      // 年份，如：2025
// 	Selectors []string `json:"selectors"` // HTML选择器列表，用于解析节假日详情
// }

// func DefaultHolidayTask(province string) string {
// 	cfg := HolidayConfig{
// 		Year:      2006,
// 		Selectors: []string{".hol-item", "li:has(div.setumei)", "li:has(div.hol-notes-mobi-zh-en)", "li"},
// 	}

// 	if province == TaiwanAreaName {
// 		cfg.Area = string(TaiwanArea)
// 		cfg.Alias = TaiwanAreaName
// 	} else {
// 		cfg.Area = string(ChinaArea)
// 		cfg.Alias = ChinaAreaName
// 	}

// 	task, _ := json.Marshal(cfg)
// 	return string(task)
// }

// // Run 执行节假日抓取任务
// //
// // 参数:
// //   - m: 邮件配置指针，成功后可发送通知
// //
// // 返回值:
// //   - *[]Holiday: 解析出的节假日列表
// //   - error: 错误信息
// func (h HolidayConfig) Run(m *MailConfig) (*[]Holiday, error) {
// 	// 构造节假日日历URL
// 	url := fmt.Sprintf("https://holidays-calendar.net/%d/calendar_zh_cn/%s_zh_cn.html", h.Year, h.Area)

// 	// 发起HTTP请求
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	// 读取响应内容
// 	buf, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var days []Holiday
// 	// 解析日历表格中的节假日信息
// 	regexHolidayCalendar(h.Area, h.Alias, h.Year, string(buf), &days)
// 	// 解析节假日详情信息
// 	regexHolidayDetail(h.Year, string(buf), &days, h.Selectors...)

// 	// 发送邮件通知
// 	if m != nil {
// 		elems := slicex.MapFunc(days, func(h Holiday) string {
// 			return fmt.Sprintf("<li>%s %s %s</li>", h.Date, h.Category, h.Detail)
// 		})
// 		go m.Send(&HolidayNoticeTemplate{
// 			Area:    h.Alias,
// 			Details: strings.Join(elems, ""),
// 		})

// 	}

// 	return emptyValueErr(m, h, &days)
// }

// // regexHolidayCalendar 从日历表格中解析节假日信息
// //
// // 参数:
// //   - area: 地区代码
// //   - alias: 地区名称
// //   - year: 年份
// //   - html: HTML内容
// //   - days: 节假日列表指针
// func regexHolidayCalendar(area string, alias string, year int, html string, days *[]Holiday) {
// 	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
// 	var month int // 当前月份计数器

// 	// 遍历所有tbody元素（每月一个表格）
// 	doc.Find("tbody").Each(func(_ int, s *goquery.Selection) {
// 		// 检查是否为月份标题行
// 		tr0 := s.Find("tr:nth-child(1) > td")
// 		if col, exist := tr0.Attr("colspan"); !exist || col != "7" {
// 			return
// 		}

// 		month++ // 月份递增

// 		// 遍历最多6周（一个月的最大周数）
// 		for i := 0; i < 6; i++ {
// 			sel := fmt.Sprintf("tr:nth-child(%d) > td", i+3)
// 			// 遍历每周的日期单元格
// 			s.Find(sel).Each(func(i int, s *goquery.Selection) {
// 				className, _ := s.Attr("class")
// 				day, err := strconv.Atoi(s.Text())
// 				// 过滤无效日期（如2017-3-39）
// 				if err != nil || day < 1 || day > 31 {
// 					return
// 				}

// 				// 构造日期对象
// 				date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
// 				dateStr := date.Format(dateFormat)

// 				// 判断节假日类型
// 				if strings.Contains(className, "hol") {
// 					// 标准节假日
// 					*days = append(*days, Holiday{
// 						Area:     area,
// 						Alias:    alias,
// 						Date:     dateStr,
// 						Category: string(HolidayOff),
// 					})
// 				} else if (date.Weekday() == time.Saturday && !strings.Contains(className, "sat")) ||
// 					(date.Weekday() == time.Sunday && !strings.Contains(className, "sun")) {
// 					// 调休工作日（周末上班）
// 					*days = append(*days, Holiday{
// 						Area:     area,
// 						Alias:    alias,
// 						Date:     dateStr,
// 						Category: string(HolidayOn),
// 					})
// 				}
// 			})
// 		}
// 	})
// }

// // regexHolidayDetail 解析节假日详情信息
// //
// // 参数:
// //   - year: 年份
// //   - html: HTML内容
// //   - days: 节假日列表指针
// //   - selectors: HTML选择器列表
// func regexHolidayDetail(year int, html string, days *[]Holiday, selectors ...string) {
// 	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
// 	var details []string

// 	// 尝试多种选择器获取节假日详情
// 	selectors = append(selectors, ".hol-item", "li:has(div.setumei)", "li:has(div.hol-notes-mobi-zh-en)", "li")
// 	for _, sel := range selectors {
// 		values := doc.Find(sel).Map(func(i int, s *goquery.Selection) string { return s.Text() })
// 		if len(values) > 0 {
// 			details = values
// 			break
// 		}
// 	}

// 	// 处理每个节假日详情
// 	for _, detail := range details {
// 		// 提取日期信息（如"1月1日"）
// 		value := fmt.Sprintf("%d年%s", year, regexp.MustCompile(`\d{1,2}月\d{1,2}日`).FindString(detail))
// 		date, err := time.ParseInLocation("2006年1月2日", value, time.Local)
// 		if err != nil {
// 			continue
// 		}

// 		// 清理补假信息（如"13日、14日补假"）
// 		detail = regexp.MustCompile(`(\d{1,2}日?[、~-])?\d{1,2}日补假`).ReplaceAllString(detail, "")
// 		// 提取节假日名称（至少2个中文字符）
// 		holName := regexp.MustCompile("(\\d{4}年)?[\u4e00-\u9fa5、]{2,}").FindString(detail)

// 		// 查找匹配的日期索引
// 		index := slicex.FirstIndexFunc(*days, func(p Holiday) bool {
// 			return p.Date == date.Format(dateFormat)
// 		})
// 		if index < 0 {
// 			continue
// 		}

// 		// 更新节假日详情
// 		(*days)[index].Detail += holName
// 		// 扩展连续节假日（前后日期）
// 		expandHoliday(days, holName, index, 1)  // 向后扩展
// 		expandHoliday(days, holName, index, -1) // 向前扩展
// 	}
// }

// // expandHoliday 扩展连续节假日
// //
// // 参数:
// //   - days: 节假日列表指针
// //   - holiday: 节假日名称
// //   - from: 起始索引
// //   - delta: 扩展方向（1向后，-1向前）
// func expandHoliday(days *[]Holiday, holiday string, from, delta int) {
// 	i := from + delta
// 	if i < 0 || i >= len(*days) {
// 		return
// 	}

// 	// 计算相邻日期
// 	d, _ := time.ParseInLocation(dateFormat, (*days)[from].Date, time.Local)
// 	date := d.AddDate(0, 0, delta).Format(dateFormat)

// 	// 检查是否为连续节假日
// 	if (*days)[i].Date == date && (*days)[i].Category == string(HolidayOff) {
// 		(*days)[i].Detail += holiday
// 		// 递归扩展
// 		expandHoliday(days, holiday, i, delta)
// 	}
// }
