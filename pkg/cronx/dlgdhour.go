package cronx

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"seeccloud.com/edscron/pkg/x/slicex"
	"seeccloud.com/edscron/pkg/x/stringx"
)

var (
	// 沪发改价管[2022]50号
	periodDocNoReg = regexp.MustCompile(`\S(?:发改|政办)[^\[]{0,10}\[\d{4}\]\d+号`)

	periodNamePat = `(?:尖峰|尖段|高峰|峰段|峰时段|平时|平段|平时段|低谷|谷段|谷时段|深谷)(?:时段)?`

	// 0-8点，17:00至22:00，0:00-6:00，0-6时
	baseTime       = `\d{1,2}(?::\d{2})?[点时]?[\-至]\d{1,2}(?::\d{2}|[点时])`
	periodValuePat = fmt.Sprintf(`(?:%s)(?:[,及]%s)*`, baseTime, baseTime)
	baseTimeReg    = regexp.MustCompile(`(\d{1,2})[:点时]?(\d{2})?[\-至](\d{1,2})[:点时]?(\d{2})?`)

	// 峰时段6小时:06:00-08:00,18:00-22:00
	periodReg      = regexp.MustCompile(fmt.Sprintf(`(%s)(?:为|\d+小时|\(|:|每日)*(%s)`, periodNamePat, periodValuePat))
	periodRevReg   = regexp.MustCompile(fmt.Sprintf(`(%s)(?:为)?(%s)`, periodValuePat, periodNamePat))
	periodOtherReg = regexp.MustCompile(`(其[他它余](?:时段)?)(?:为)?(平段|平时)`)

	conditionPat = fmt.Sprintf(`%s|%s|%s|%s|%s`, tempPat, weekendPat, monthPat, holidayPat, categoryPat)
	conditionReg = regexp.MustCompile(conditionPat)

	// 尖峰时段:7月,8月20:00-22:00,其他月份18:00-20:00
	// 高峰时段16:00-24:00(7,8月为16:00-20:00;1,12月为16:00-18:00,22:00-24:00)
	periodBetweenReg    = regexp.MustCompile(fmt.Sprintf(`(%s)\p{Han}{0,3}((?:(?:[:,;\(]*)(?:%s|,)*\p{Han}{0,3}(?:%s))+)`, periodNamePat, conditionPat, periodValuePat))
	periodBetweenSubReg = regexp.MustCompile(fmt.Sprintf(`(%s)*\p{Han}{0,3}(%s)`, conditionPat, periodValuePat))

	// 尖峰时段:20:00-24:00(7,8月),18:00-22:00(1,12月)
	periodAfterReg    = regexp.MustCompile(fmt.Sprintf(`(%s)\p{Han}{0,3}((?:(?:[:,]*)(?:%s)\((?:%s)\))+)`, periodNamePat, periodValuePat, monthPat))
	periodAfterSubReg = regexp.MustCompile(fmt.Sprintf(`(%s)\((%s)\)`, periodValuePat, monthPat))

	// 7月21:00—23:00,1,11,12月19:00—21:00为尖峰时段
	multiValuesReg = regexp.MustCompile(fmt.Sprintf(`((?:%s)[为是:])?(\(?(?:%s)\)?(?:%s)[,;]?){2,}([为是:](?:%s))?`, periodNamePat, monthPat, periodValuePat, periodNamePat))

	periodNameReg  = regexp.MustCompile(periodNamePat)
	periodValueReg = regexp.MustCompile(periodValuePat)

	months    = [12]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	baseMonth = `\d[\d,\-月和至以及]*月`
	monthPat  = fmt.Sprintf(`%s|其他月份`, baseMonth)
	monthReg  = regexp.MustCompile(".*月(份)?$")

	// "x月的休息日", 注意"的"已经预处理
	weekendPat = fmt.Sprintf(`%s(?:,%s)*休息日`, baseMonth, baseMonth) //`\d[\d,\-月和至及]*月(?:,\d[\d,\-月和至及]*月)*休息日`
	weekendReg = regexp.MustCompile(weekendPat)

	holiday    = `(?:元旦|春节|清明节|劳动节|端午节|中秋节|国庆节)(?:,\s*(?:元旦|春节|清明节|劳动节|端午节|中秋节|国庆节))*`
	holidayPat = fmt.Sprintf(`%s|3天及以上节假日`, holiday)
	holidayReg = regexp.MustCompile(holidayPat)

	categoryPat = `容量\d+千伏安及以上|大工业用电|单一制|两部制`
	cateReg     = regexp.MustCompile(categoryPat)

	tempPat = `其他月份.{0,15}最高气温[^\d]*\d+℃`
	tempReg = regexp.MustCompile(tempPat)
)

type DlgdPeriod struct {
	Area  string     `json:"area"`  // 区域
	DocNo string     `json:"docNo"` // 正常文号
	Hours []DlgdHour `json:"hours"`
}

type DlgdHour struct {
	Name          string   `json:"name"`          // 时段名称
	Value         string   `json:"value"`         // 时段值
	Temp          string   `json:"temp"`          // 温度条件
	Months        []int64  `json:"months"`        // 月份条件
	WeekendMonths []int64  `json:"weekendMonths"` // 月份-休息日条件
	Holidays      []string `json:"holidays"`      // 节假日条件
	Categories    []string `json:"categories"`    // 用电类别条件

	conditions []string `json:"-"`
}

func (d *DlgdPeriod) normalize() {
	d.DocNo = strings.TrimPrefix(d.DocNo, ",")
	for i, p := range d.Hours {

		// 1. 标准化时段名称
		switch {
		case strings.HasPrefix(p.Name, "尖"):
			p.Name = "尖段"
		case strings.HasPrefix(p.Name, "高"), strings.HasPrefix(p.Name, "峰"):
			p.Name = "峰段"
		case strings.HasPrefix(p.Name, "平"):
			p.Name = "平段"
		case strings.HasPrefix(p.Name, "低"), strings.HasPrefix(p.Name, "谷"):
			p.Name = "谷段"
		case strings.HasPrefix(p.Name, "深"):
			p.Name = "深谷"
		}

		// 2. 标准化时段值
		p.Value = baseTimeReg.ReplaceAllStringFunc(p.Value, func(s string) string {
			subs := baseTimeReg.FindStringSubmatch(s)
			if len(subs) < 5 {
				return s
			}
			return fmt.Sprintf("%02s:%02s-%02s:%02s", subs[1], subs[2], subs[3], subs[4])
		})

		// 只有平时段（最低优先级）值可能是：其他，全时段覆盖即可
		if strings.HasPrefix(p.Value, "其") {
			p.Value = "00:00-24:00"
		}

		// 3. 标准化条件
		for j, cond := range p.conditions {
			if monthReg.MatchString(cond) {
				p.Months = stringx.MustInts(cond, 1, 12)
				if cond == "其他月份" {
					p.Months = months[:]
					slicex.EachFunc(d.Hours[:i], func(pp DlgdHour) {
						if pp.Name == p.Name &&
							len(pp.conditions) == len(p.conditions) &&
							slices.Equal(pp.conditions[:j], p.conditions[:j]) {
							p.Months = slicex.Diff(p.Months, pp.Months)
						}
					})
				}
			} else if holidayReg.MatchString(cond) {
				p.Holidays = append(p.Holidays, strings.Split(cond, ",")...)
			} else if weekendReg.MatchString(cond) {
				p.WeekendMonths = stringx.MustInts(cond, 1, 12)
			} else if cateReg.MatchString(cond) {
				p.Categories = append(p.Categories, cond)
			} else if tempReg.MatchString(cond) {
				p.Temp = cond
			}
		}
		p.Holidays = slicex.RemoveDuplicates(p.Holidays)

		d.Hours[i] = p
	}
}

func NewPeriod(text string) DlgdPeriod {

	var period DlgdPeriod

	// 1. 排除干扰元素
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`[：∶﹕]`).ReplaceAllString(text, ":")
	text = regexp.MustCompile(`（`).ReplaceAllString(text, "(")
	text = regexp.MustCompile(`）`).ReplaceAllString(text, ")")
	text = regexp.MustCompile(`[﹝〔]`).ReplaceAllString(text, "[")
	text = regexp.MustCompile(`[〕﹞]`).ReplaceAllString(text, "]")
	text = regexp.MustCompile(`[，、]`).ReplaceAllString(text, ",")
	text = regexp.MustCompile(`；`).ReplaceAllString(text, ";")
	text = regexp.MustCompile(`[-—–−－]+`).ReplaceAllString(text, "-")
	text = regexp.MustCompile(`\(含\)|\(含尖峰\)|\(含深谷\)`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`\(?共?\d+小时\)?`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`\(?次[日年]\)?`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`[春夏秋冬]季`).ReplaceAllString(text, "")
	// "由高峰时段调整为尖峰时段"→"为尖峰时段"
	text = regexp.MustCompile(fmt.Sprintf(`由%s调整`, periodNamePat)).ReplaceAllString(text, "")
	// "7月份的21:00—23:00" → "7月21:00—23:00"
	text = regexp.MustCompile(`(\d+)月[份的]+`).ReplaceAllString(text, "${1}月")
	// "7月-9月"、"7至9月" → "7-9月"
	text = regexp.MustCompile(`(\d+)月?[-至到](\d+)月`).ReplaceAllString(text, "${1}-${2}月")

	// 用index因条件在字符串中的位置影响时段匹配
	// subss := regexp.MustCompile(`(\d{4})-(\d{2})`).FindAllStringSubmatchIndex(part, -1)
	// 原始字符串: 今天是2023-10，明天是2023-11，昨天是2023-09
	// 索引结果: [
	//  [3 9 3 7 8 9]        // 第一个匹配项："2023-10", 子组1: "2023"[3:7],   子组2: "10"[8:9]
	//  [12 18 12 16 17 18]  // 第二个匹配项："2023-11", 子组1: "2023"[12:16], 子组2: "11"[17:18]
	//  [21 27 21 25 26 27]  // 第三个匹配项："2023-09", 子组1: "2023"[21:25], 子组2: "09"[26:27]
	// ]

	// 2. 按电价表备注点遍历：
	// 备注: 1.xxx 2.xxx 3.xxx
	coments := regexp.MustCompile(`\d\.[^0-9]`).Split(text, -1)
	for _, comment := range coments {
		periods := make([]DlgdHour, 0)

		// 2.1 按句拆分遍历："。"、"(1)"、"(2)"等
		parts := regexp.MustCompile(`。|\(\d+\)`).Split(comment, -1)
		for _, part := range parts {

			// 特殊一：专属条件
			periods = append(periods, matchBetween(&part)...)
			periods = append(periods, matchAfter(&part)...)
			periods = append(periods, matchMulti(&part)...)

			condIndexs := conditionReg.FindAllStringIndex(part, -1)

			// 特殊二：一段语句(part)只含一个时段值，且时段名和时段值间隔字符不确定，通常用于"尖峰"或"深谷"，
			name := periodNameReg.FindStringIndex(part)
			values := periodValueReg.FindAllStringIndex(part, -1)
			if len(name) > 0 && len(values) == 1 {
				periods = append(periods, DlgdHour{
					Name:       part[name[0]:name[1]],
					Value:      part[values[0][0]:values[0][1]],
					conditions: matchPreContitions(part, condIndexs, len(part), len(part)),
				})
				continue
			}

			subss := periodReg.FindAllStringSubmatchIndex(part, -1)
			for _, subs := range subss {
				if len(subs) != 6 {
					continue
				}

				periods = append(periods, DlgdHour{
					Name:       part[subs[2]:subs[3]],
					Value:      part[subs[4]:subs[5]],
					conditions: matchPreContitions(part, condIndexs, subs[2], subs[4]),
				})
			}

			subss = periodRevReg.FindAllStringSubmatchIndex(part, -1)
			for _, subs := range subss {
				if len(subs) != 6 {
					continue
				}

				periods = append(periods, DlgdHour{
					Name:       part[subs[4]:subs[5]],
					Value:      part[subs[2]:subs[3]],
					conditions: matchPreContitions(part, condIndexs, subs[4], subs[2]),
				})
			}

			subss = periodOtherReg.FindAllStringSubmatchIndex(part, -1)
			for _, subs := range subss {
				if len(subs) != 6 {
					continue
				}

				periods = append(periods, DlgdHour{
					Name:       part[subs[4]:subs[5]],
					Value:      part[subs[2]:subs[3]],
					conditions: matchPreContitions(part, condIndexs, subs[4], subs[2]),
				})
			}

		}

		if len(periods) > 0 {
			period.Hours = append(period.Hours, periods...)
			docNo := periodDocNoReg.FindString(comment)
			if !strings.Contains(period.DocNo, docNo) {
				period.DocNo += "," + docNo
			}
		}
	}

	if period.DocNo == "" {
		period.DocNo = periodDocNoReg.FindString(text)
	}

	(&period).normalize()
	return period
}

func matchBetween(part *string) []DlgdHour {
	periods := []DlgdHour{}
	btws := periodBetweenReg.FindAllStringSubmatch(*part, -1)
	for _, btw := range btws {
		if len(btw) != 3 {
			continue
		}

		subBtws := periodBetweenSubReg.FindAllStringSubmatch(btw[2], -1)

		// 存在中间条件才有效
		if !slicex.Any(subBtws, func(subs []string) bool { return len(subs) == 3 && subs[1] != "" }) {
			continue
		}

		for _, subs := range subBtws {
			if len(subs) != 3 {
				continue
			}

			periods = append(periods, DlgdHour{
				Name:       btw[1],
				Value:      subs[2],
				conditions: []string{subs[1]},
			})
		}

		*part = strings.ReplaceAll(*part, btw[0], strings.Repeat("#", len(btw[0])))
	}
	return periods
}

func matchAfter(part *string) []DlgdHour {
	periods := []DlgdHour{}
	afts := periodAfterReg.FindAllStringSubmatch(*part, -1)
	for _, aft := range afts {
		if len(aft) != 3 {
			continue
		}

		subAfts := periodAfterSubReg.FindAllStringSubmatch(aft[2], -1)
		for _, subs := range subAfts {
			if len(subs) != 3 {
				continue
			}

			periods = append(periods, DlgdHour{
				Name:       aft[1],
				Value:      subs[1],
				conditions: []string{subs[2]},
			})
		}

		*part = strings.ReplaceAll(*part, aft[0], strings.Repeat("#", len(aft[0])))
	}
	return periods
}

func matchMulti(part *string) []DlgdHour {
	periods := []DlgdHour{}
	mults := multiValuesReg.FindAllString(*part, -1)
	for _, mult := range mults {
		name := periodNameReg.FindString(mult)
		conds := conditionReg.FindAllString(mult, -1)
		values := periodValueReg.FindAllString(mult, -1)
		if len(conds) != len(values) {
			continue
		}

		for i := range conds {
			periods = append(periods, DlgdHour{
				Name:       name,
				Value:      values[i],
				conditions: []string{conds[i]},
			})
		}

		*part = strings.ReplaceAll(*part, mult, strings.Repeat("#", len(mult)))
	}

	return periods
}

// matchPreContitions 匹配前置条件
// "其他月份……6-8月,12月-次年2月：高峰时段14:00-22:00;3-5月,9-11月：高峰时段15:00-22:00。"
// cons: ["其他月份","6-8月,12月-次年2月","3-5月,9-11月"]
// 时段1: "6-8月,12月-次年2月" → "高峰时段14:00-22:00"
// 时段2: "3-5月,9-11月"      → "高峰时段15:00-22:00"
func matchPreContitions(s string, condIndexs [][]int, nameIndex, valueIndex int) []string {
	cons := []string{}
	for _, ids := range condIndexs {
		if len(ids) == 2 && ids[0] < nameIndex && ids[0] < valueIndex {
			cons = append(cons, s[ids[0]:ids[1]])
		}
	}

	cons = slicex.RemoveDuplicates(cons)

	// 多个月份条件时取离得最近的
	hasMonth := false
	for i := len(cons) - 1; i >= 0; i-- {
		if monthReg.MatchString(cons[i]) {
			if hasMonth {
				cons = append(cons[:i], cons[i+1:]...)
			} else {
				hasMonth = true
			}
		}
	}

	return cons
}
