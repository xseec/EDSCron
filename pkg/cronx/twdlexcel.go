package cronx

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"seeccloud.com/edscron/pkg/x/slicex"
	"seeccloud.com/edscron/pkg/x/timex"
)

// 本模型不再考虑“包灯”和“包用”类别
// -----------------------------------------------------------------
// 電價類別：
// 一、包燈電價
// 	  (一)電燈(夜間供電電價，日夜供電者加倍計收)
// 	  (二)小型器具(日夜供電電價，僅於日間或夜間供電者減半計收)
// 	  (三)交通指揮燈(日夜供電電價)
// 二、包用電力電價
// 三、表燈(住商)電價
// 	  (一)非時間電價
// 	  		1.住宅用
// 	  		2.住宅以外非營業用
// 	  		3.營業用
// 	  (二)時間電價
// 	  		1.簡易型時間電價
// 	  			(1)二段式
// 	  			(2)三段式
// 	  		2.標準型時間電價
// 	  			(1)二段式
// 	  			(2)三段式
// 四、低壓電力電價
// 	  (一)非時間電價
// 	  (二)時間電價
// 	  		1.二段式
// 	  		2.三段式
// 	  		3.電動車充換電設施電價
// 五、高壓及特高壓電力電價
// 	  (一)二段式時間電價
// 	  (二)三段式時間電價（不考虑“尖峰时间可变动”情况）
// 	  (三)批次生產時間電價
// 	  (四)高壓電動車充換電設施電價：按低壓電動車充換電設施電價 95%計算。
// -----------------------------------------------------------------

type CategoryItem struct {
	Text  string
	Level int
}

// TwdlRow 台湾电力电价明细，繁体转简体可能出错，twdl避免繁体标识,\\p{Han}{n}用\\p{Han}{n-1,n+1}代替
type TwdlRow struct {
	StartTime time.Time `json:"start_time"`                     // 起始时间
	Category  string    `json:"category"`                       // 用电类别
	Date      string    `json:"date"`                           // 日期范围(含首尾)，格式如"0101-0531,1001-1231"、"0601-0930"
	Stage     string    `json:"stage" twdl:"((?:超過)?\\d+.*部分)"` // 阶梯电量阈值，加最外围括号目的是提取tile置于Stage的值中
	// 以下字段定义流动电费
	Standard        float64 `json:"standard" twdl:"^流\\p{Han}{2,4}每度$"`     // 非时间基准电价，tag: 流動電費每度
	WeekdayPeak     float64 `json:"weekday_peak" twdl:"五尖峰.*"`              // 周一至周五尖峰电价，tag: 週一至週五尖峰時間
	WeekdaySemiPeak float64 `json:"weekday_semi_peak" twdl:"[一|五].*半尖峰.*"`  // 周一至周五半尖峰电价，tag: 週一至週五半尖峰時間
	WeekdayOffPeak  float64 `json:"weekday_off_peak" twdl:"[一|五].*[^尖]峰.*"` // 周一至周五离峰电价，tag: 週一至週五離峰時間
	SatSemiPeak     float64 `json:"sat_semi_peak" twdl:".*六.*半尖峰.*"`        // 周六半尖峰电价，tag: 週六半尖峰時間
	SatOffPeak      float64 `json:"sat_off_peak" twdl:".*六.*[^尖]峰.*"`       // 周六离峰电价，tag: 週六離峰時間
	SunOffPeak      float64 `json:"sun_off_peak" twdl:".*日.*[^尖]峰.*"`       // 周日及离峰日电价，tag: 週日及離峰日離峰時間全日、
	// 以下字段定义时段
	WeekdayPeakHour     string `json:"weekday_peak_hour" twdl:"五尖峰.*" hour:"[0-9:：~]{11,}"`                 // 周一至周五尖峰时段
	WeekdaySemiPeakHour string `json:"weekday_semi_peak_hour" twdl:"[一|五].*半尖峰.*" hour:"[0-9:：~]{11,}|全日"`  // 周一至周五半尖峰时段
	WeekdayOffPeakHour  string `json:"weekday_off_peak_hour" twdl:"[一|五].*[^尖]峰.*" hour:"[0-9:：~]{11,}|全日"` // 周一至周五离峰时段
	SatSemiPeakHour     string `json:"sat_semi_peak_hour" twdl:".*六.*半尖峰.*" hour:"[0-9:：~]{11,}|全日"`        // 周六半尖峰时段
	SatOffPeakHour      string `json:"sat_off_peak_hour" twdl:".*六.*[^尖]峰.*" hour:"[0-9:：~]{11,}|全日"`       // 周六离峰时段，tag: 週六離峰時間00：00~24：00
	SunOffPeakHour      string `json:"sun_off_peak_hour" twdl:".*日.*[^尖]峰.*" hour:"[0-9:：~]{11,}|全日"`       // 周日及离峰日时段，tag: 週日及離峰日離峰時間全日
	// 以下字段定义基本电费
	InstalledCustomer   float64 `json:"installed_customer" twdl:"(?:基本\\p{Han}{1,3}|置契\\p{Han}{1,2})按.*收[每三]"`   // 按户计收，tag需排除下列两种情况：单相或需量
	InstalledCustomer1P float64 `json:"installed_customer_1p" twdl:"按.*收[^每三]相"`                                 // 按户计收单相，tag: 按戶計收單相
	RegularCustomer     float64 `json:"regular_customer" twdl:"需量契\\p{Han}{1,2}按.*收"`                            // 需量契约按户计收，tag: 需量契約按戶計收
	InstalledCap        float64 `json:"installed_cap" twdl:"(?:基本\\p{Han}{1,3}|置契\\p{Han}{1,2})\\p{Han}{1,2}置契"` // 装置契约，tag: 裝置契約
	RegularCap          float64 `json:"regular_cap" twdl:"常契"`                                                   // 经常契约，tag: 經常契約
	NonSummerCap        float64 `json:"non_summer_cap" twdl:"非夏月契"`                                              // 非夏月契约, tag: 非夏月契約
	SemiPeakCap         float64 `json:"semi_peak_cap" twdl:"[^六]半尖峰契"`                                           // 半尖峰契约，tag: 半尖峰契約
	SatSemiPeakCap      float64 `json:"sat_semi_peak_cap" twdl:"六半尖峰契"`                                          // 周六半尖峰契约，tag: 週六半尖峰契約
	OffPeakCap          float64 `json:"off_peak_cap" twdl:"[^尖]峰契"`                                              // 离峰契约，tag: 離峰契約
}

const (
	twTag     = "twdl"
	twHourTag = "hour"
	maxPage   = 12
)

var (
	TwdlCategories = []string{
		"表燈(住商)電價>非時間電價>住宅用",
		"表燈(住商)電價>非時間電價>住宅以外非營業用",
		"表燈(住商)電價>非時間電價>營業用",
		"表燈(住商)電價>時間電價>簡易型時間電價>二段式",
		"表燈(住商)電價>時間電價>簡易型時間電價>三段式",
		"表燈(住商)電價>時間電價>標準型時間電價>二段式",
		"表燈(住商)電價>時間電價>標準型時間電價>三段式",
		"低壓電力電價>非時間電價",
		"低壓電力電價>時間電價>二段式",
		"低壓電力電價>時間電價>三段式",
		"低壓電力電價>時間電價>電動車充換電設施電價",
		"高壓及特高壓電力電價>二段式時間電價>高壓供電",
		"高壓及特高壓電力電價>二段式時間電價>特高壓供電",
		"高壓及特高壓電力電價>三段式時間電價>高壓供電",
		"高壓及特高壓電力電價>三段式時間電價>特高壓供電",
		"高壓及特高壓電力電價>批次生產時間電價>高壓供電",
		"高壓及特高壓電力電價>批次生產時間電價>特高壓供電",
	}

	// 匹配以下文本时已提前处理所有空字符

	// 电价表生效日期，常见于目录页
	twStartDateReg = regexp.MustCompile(`(\d+)年(\d+)月(\d+)日起`)
	// 电价表目录页，页码
	twPageNumReg = regexp.MustCompile(`(?:壹|調整後|電價表).*?(\d+)(?:貳|凍漲|電價表).*?(\d+)`)
	// 截取PDF文档过长时，至此结束
	twPart2Reg = regexp.MustCompile(`高壓電動車|貳、|凍漲行業`)

	// 电价类别提取层级和文本
	// 经多次尝试尚无法优雅从第一层级“三、表燈(住商)電價(一)非時間電價”中提取“表燈(住商)電價”
	// 实际提取“表燈”，需在后续方法内特殊处理
	categoryReg = regexp.MustCompile(`(?:[一二三四五六七八九十]+、([^(]+))|` + // 第一层级
		`(?:\([一二三四五六七八九十]+\)([^1-9：]+))|` + // 第二层级
		`(?:[1-9]+\.+([^(]+))|` + // 第三层级
		`(?:\(([1-9])\)(二段式|三段式))`)

	// 电价类别文本中含单位干扰
	unitReg        = regexp.MustCompile(`(單位|单位)[:：]元`)
	highVoltageReg = regexp.MustCompile(`^(?:高|特高)\p{Han}{3,5}`)

	// 电价类别单元格锚点：同行相邻两列以“夏月”和“非夏月”开头
	summerReg    = regexp.MustCompile(`^夏月`)
	nonSummerReg = regexp.MustCompile(`^非`)

	dateReg = regexp.MustCompile(`(\d+)月(\d+)日`)

	// 流动电价中“夏月”、“非夏月”时段已被OCR拆成多行，需合并“[非]夏月”前缀相同的行，且“夏月”可能被误识别为“夏）月”
	// 举例：
	// 1. 流動電費(尖峰時間固定)週六離峰時間非夏）月00：00~06：00每度
	// 2. 流動電費(尖峰時間固定)週六離峰時間非夏）月11：00~14：00每度
	// 合并：
	// 流動電費(尖峰時間固定)週六離峰時間非夏）月00：00~06：00每度11：00~14：00每度
	twdlTitleReg = regexp.MustCompile(`(.*夏.{0,5}月)(.*)`)

	// 空数值可能为:"", " ", "-", "——"，考虑所有类横杆情况
	twdlValueReg = regexp.MustCompile(`^(\s*加\s*)?\d+(\.\d+)?$|^(\s*[\-‐‑‒–—―−＿﹘－ーｰ⁃﹉﹊─━➖⸺⸻᐀᠆]\s*)+$|^\s*$`)

	// 不考虑“尖峰时间可变动”情况，遍历行列遇到“指定30天”时截止
	twSharpDynamicReg = regexp.MustCompile(`指定\d+天`)
)

// excelizeTwdl 解析电价表
func excelizeTwdl(path string, startDate string, twdls *[]TwdlRow) error {

	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}

	defer f.Close()

	sheets := f.GetSheetList()

	// 工作表中“夏月”&“非夏月”出现的行startRow、列startCol
	// 同行中“夏月"、“非夏月”列数colSpan
	// 此电价类别结束行endRow、列endCol
	var startRow, endRow, startCol, colSpan int
	// 夏月范围：["6月1日至9月30日","6","1","9","30"]
	var summerRange SummerRange
	// 电价类别下电价条目
	titles := make([]string, 0)
	// 电价类别
	categories := make([]CategoryItem, 0)

SheetLoop:
	for _, sheet := range sheets {

		// 工作表下行列合集([][]string)
		rows, err := f.GetRows(sheet)
		if err != nil {
			return err
		}

		// 遍历单元格，获取“电价类别”锚点
		for i, row := range rows {
			rowText := regexp.MustCompile(`\s+`).ReplaceAllString(strings.Join(row, ""), "")

			// （1）必须先找到本期电价生效时间
			if len(startDate) == 0 {
				if match := twStartDateReg.FindStringSubmatch(rowText); len(match) == 4 {
					startDate = fmt.Sprintf("%s年%s月%s日", match[1], match[2], match[3])
				}
				continue
			}

			// （2）已找到“电价类别”锚点
			if startRow != 0 {

				// 获取“夏月”日期范围
				if len(summerRange) == 0 {
					value := mustRange(f, sheet, startRow, startCol, i+1, startCol, true)
					if match := summerRangeReg.FindString(value); len(match) > 0 {
						summerRange = SummerRange(match)
						continue
					}
				}

				cell := mustCell(f, sheet, i+1, startCol)

				// 合并单元格原因，“夏月”日期范围可能会重复（纵跨3行）
				if dateReg.MatchString(cell) {
					continue
				}

				// 遍历到有效值范围结束
				title := mustRange(f, sheet, i+1, 1, i+1, startCol-1, true)
				isDynamic := twSharpDynamicReg.MatchString(title)
				hasValue := twdlValueReg.MatchString(cell)
				if !isDynamic && hasValue {
					endRow = i + 1
					titles = adjustTwdlTitle(titles, title)
				}

				// 因OCR存在误识别导致：电价Title和电价Value不在同行，独立处理Tile和Value列表，保证两者数目相等就能匹配得上

				// 遍历到最后一行或下一行无有效数据，生成此电价类别
				if isDynamic || !hasValue || i == len(rows)-1 {
					values := adjustTwdlValue(f, sheet, startRow+1, startCol, endRow, startCol+colSpan-1)
					rows := newTwdlRows(f, sheet, startDate, startRow, startCol, colSpan, summerRange, categories, titles, values)
					*twdls = append(*twdls, rows...)

					startRow = 0
					endRow = 0
					startCol = 0
					colSpan = 0
					summerRange = ""
					titles = make([]string, 0)
				}
			}

			// （3）是否结束，需在（2）“生成电价类别”之后，防止在目录页内容时就跳出
			if len(*twdls) > 0 && twPart2Reg.MatchString(rowText) {
				break SheetLoop
			}

			// （4）未找到“电价类别”锚点，遍历列
			for j := range len(row) - 1 {
				cell1 := mustCell(f, sheet, i+1, j+1)
				cell2 := mustCell(f, sheet, i+1, j+2)
				// 寻找“夏月”、“非夏月”在同行相邻列
				if summerReg.MatchString(cell1) && nonSummerReg.MatchString(cell2) {
					startRow = i + 1
					startCol = j + 1
					colSpan = len(row) - j
					if summerRangeReg.MatchString(cell1) {
						summerRange = SummerRange(cell1)
					}

					break
				}
			}

			categories = AddTwdlCategoryIfMatched(categories, mustCell(f, sheet, i+1, 1))
		}
	}

	if len(startDate) == 0 {
		return fmt.Errorf("未找到电价生效时间")
	}

	return nil
}

// newTwdlRows 生成电价数据
func newTwdlRows(f *excelize.File, sheet, startDate string, row, col, colSpan int, summerRange SummerRange, categories []CategoryItem, titles []string, values [][]string) []TwdlRow {
	results := make([]TwdlRow, 0)
	summer, nonSummer := summerRange.MustRange()
	st := timex.MustDate(startDate)
	for i := range colSpan {
		twdl := TwdlRow{
			StartTime: st,
			Category:  joinCategory(f, sheet, categories, row, col+i),
		}

		if i%2 == 0 {
			twdl.Date = summer
		} else {
			twdl.Date = nonSummer
		}

		colValues := slicex.MapFunc(values, func(v []string) string { return v[i] })
		twdl.evaluate(titles, colValues)
		results = append(results, twdl)
	}

	return results
}

// evaluate 给电价记录赋值
func (row *TwdlRow) evaluate(titles, values []string) {
	dlType := reflect.TypeOf(row).Elem()
	dlValue := reflect.ValueOf(row).Elem()

	// 遍历所有列
	for index, title := range titles {
		// 遍历结构体所有字段
		for i := 0; i < dlType.NumField(); i++ {
			// 获取字段的dlgd标签
			tag := dlType.Field(i).Tag.Get(twTag)
			if len(tag) == 0 {
				continue
			}

			// 匹配标题和标签
			reg := regexp.MustCompile(tag)
			if reg.MatchString(title) {
				switch dlValue.Field(i).Kind() {
				case reflect.Float64:
					if value, ok := strconv.ParseFloat(values[index], 64); ok == nil {
						dlValue.Field(i).SetFloat(value)
					}
				case reflect.String:
					// 时间字段直接从标题中提取，需保证value有真实的数值，否则“夏月”将被“非夏月”覆盖
					if hour := dlType.Field(i).Tag.Get(twHourTag); len(hour) > 0 {
						if _, ok := strconv.ParseFloat(values[index], 64); ok == nil {
							ss := regexp.MustCompile(hour).FindAllString(title, -1)
							dlValue.Field(i).SetString(formatTwHour(strings.Join(ss, "")))
						}

						continue
					}

					s := values[index]
					if subs := reg.FindStringSubmatch(title); len(subs) >= 2 {
						s = subs[1] + fieldKeyValueSep + s
					}

					s = dlValue.Field(i).String() + fieldSubSep + s
					s = strings.Trim(s, fieldSubSep)
					dlValue.Field(i).SetString(s)
				}
			}
		}
	}
}

// 格式化电价类别，加入“高压供电”和“特高压供电”等细类
func joinCategory(f *excelize.File, sheet string, categories []CategoryItem, row, col int) string {
	if len(categories) == 0 {
		return ""
	}

	cats := slicex.MapFunc(categories, func(c CategoryItem) string { return c.Text })
	category := strings.Join(cats, CategorySep)
	cell := mustCell(f, sheet, row-1, col)
	if m := highVoltageReg.FindString(cell); len(m) > 0 {
		category += CategorySep + m
	}

	return category
}

// 调整电价类别层级, row 行号(1-based)
func AddTwdlCategoryIfMatched(cats []CategoryItem, value string) []CategoryItem {
	value = unitReg.ReplaceAllString(value, "")
	matches := categoryReg.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		var text string
		var level int

		// 确定匹配到的具体模式和层级
		for i := 1; i < len(match); i++ {
			if match[i] != "" {
				switch {
				case i == 1: // 第一层级
					text = match[1]
					// 以上正则表达式尚无法从"表燈(住商)電價(一)非時間電價"提取"表燈(住商)電價"
					if text == "表燈" {
						text = "表燈(住商)電價"
					}

					level = 0
				case i == 2: // 第二层级
					text = match[2]
					level = 1
				case i == 3: // 第三层级
					text = match[3]
					level = 2
				case i >= 4: // 第四层级
					// 第四层级有两个捕获组：match[4]是数字，match[5]是"二段式"或"三段式"
					text = match[5]
					level = 3
				}
				break
			}
		}

		if text != "" {
			item := CategoryItem{Text: text, Level: level}
			endIndex := len(cats)
			for i, cat := range cats {
				if cat.Level >= level {
					endIndex = i
					break
				}
			}

			cats = append(cats[:endIndex], item)
		}
	}

	return cats
}

// 调整有效的电价条目
//
// 参数：
//   - titles: ["流動電費xxxx夏月00：00~09：00每度"]
//   - title: "流動電費xxxx夏月12：00~15：00每度"
//
// 返回值：["流動電費xxxx夏月00：00~09：00每度12：00~15：00每度"]
func adjustTwdlTitle(titles []string, title string) []string {
	subs := twdlTitleReg.FindStringSubmatch(title)
	if len(subs) != 3 {
		return append(titles, title)
	}

	var builder strings.Builder
	for i, p := range titles {
		temps := twdlTitleReg.FindStringSubmatch(p)
		if len(temps) == 3 && subs[1] == temps[1] && !strings.Contains(subs[2], temps[2]) {
			builder.WriteString(subs[1])
			builder.WriteString(temps[2])
			builder.WriteString(subs[2])
			titles[i] = builder.String()
			return titles
		}
	}
	return append(titles, title)
}

// 调整有效的电价值
func adjustTwdlValue(f *excelize.File, sheet string, startRow, startCol, endRow, endCol int) [][]string {
	cells, err := f.GetMergeCells(sheet)
	if err != nil {
		return nil
	}

	// 从上到下(startRow→endRow)查询纵向合并单元格，若存在纵向合并单元格，相邻列数值合并单元格
	//     夏月            非夏月
	// ---------------------------------
	//              |                  |
	//     1.00     |-------------------
	//              |       ——         |
	// ---------------------------------
	//          (合并如下)
	// ---------------------------------
	//     1.00     |       ——         |
	// ---------------------------------
	results := make([][]string, 0)
	for i := startRow; i <= endRow; i++ {
		mergedRow := i
		// 同行中是否存在纵向合并单元格
		for j := startCol; j <= endCol; j++ {
			if endR, _ := getMergedOrCellEndAxis(cells, i, j); endR > i {
				mergedRow = endR
				break
			}
		}

		// 同列中是否存在横向合并单元格
		values := make([]string, 0)
		for j := startCol; j <= endCol; j++ {
			values = append(values, mustRange(f, sheet, i, j, mergedRow, j, true))
		}

		// “夏月”和相邻的“非夏月”至少一个值有效
		valid := false
		for k := 0; k < len(values)-1; k += 2 {
			if twdlValueReg.MatchString(values[k]) || twdlValueReg.MatchString(values[k+1]) {
				valid = true
				break
			}
		}

		if valid {
			results = append(results, values)
		}

		i = mergedRow
	}

	return results
}

// formatTwdls 格式化电价数据
func formatTwdls(rows []TwdlRow, useHtml bool) string {
	var builder strings.Builder

	for _, row := range rows {
		color := ""
		rowSeparator := "\n"
		if useHtml {
			rowSeparator = "<hr/>"
			if slices.Contains(TwdlCategories, row.Category) {
				color = "gray"
			} else {
				color = "red"
			}
		}

		dlType := reflect.TypeOf(row)
		dlValue := reflect.ValueOf(row)
		for i := 0; i < dlType.NumField(); i++ {
			field := dlType.Field(i)
			value := dlValue.Field(i)

			if !value.IsZero() {
				if useHtml {
					builder.WriteString(fmt.Sprintf("<li style='color:%s;'><em>%s</em> : %v</li>\n",
						color, field.Name, value.Interface()))
				} else {
					builder.WriteString(fmt.Sprintf("%s: %v\n", field.Name, value.Interface()))
				}
			}
		}

		builder.WriteString(rowSeparator)
	}

	if useHtml {
		return "<ul>\n" + builder.String() + "</ul>"
	}
	return builder.String()
}
