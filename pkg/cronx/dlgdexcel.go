package cronx

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
	"seeccloud.com/edscron/pkg/x/slicex"
)

// 常量定义
const (
	yuanUnit         = "元/千瓦时"                 // 电费单位：元/千瓦时
	fenUnit          = "分/千瓦时"                 // 电费单位：分/千瓦时（1元=100分）
	defSheet         = "sheet1"                // 默认Excel工作表名称
	voltageName      = "电压等级"                  // 电压等级列标题
	voltageSZName    = "用电类别"                  // 深圳特殊列标题
	voltagePattern   = "千伏(及以上)?$"             // 电压等级匹配正则
	voltageSZPattern = "(千伏安·月|千瓦·月|接入用电\\))$" // 深圳特殊电压匹配正则
	dlgdTag          = "dlgd"                  // 结构体字段标签前缀
	unitTag          = "unit"                  // 单位转换标签
)

// 单位转换映射表
var units = map[string]float64{
	yuanUnit: 1,   // 元/千瓦时不需转换
	fenUnit:  100, // 分/千瓦时需要除以100转换为元
}

// format 格式化DlgdRow数据
// 参数unit: 单位转换系数（分转元需要除以100）
func (row *DlgdRow) format(unit float64) {
	rType := reflect.TypeOf(DlgdRow{})
	rValue := reflect.ValueOf(row).Elem()

	for i := 0; i < rType.NumField(); i++ {
		// 处理字符串字段：去除末尾分隔符
		if rValue.Field(i).Kind() == reflect.String {
			rValue.Field(i).SetString(strings.Trim(rValue.Field(i).String(), fieldValueSep))
			continue
		}

		// 对有unit标签的数值字段进行单位转换
		if _, ok := rType.Field(i).Tag.Lookup(unitTag); ok {
			rValue.Field(i).SetFloat(rValue.Field(i).Float() / unit)
		}
	}
}

// unexcelize 从Excel文件解析电价数据
//
// 参数:
//   - name: 文件名
//   - initRow: 初始化行结构体
//   - dlgd: 存储解析结果的指针
//
// 返回:
//   - error: 错误信息
func unexcelize(name string, initRow DlgdRow, dlgd *Dlgd) error {
	// 打开Excel文件
	f, err := excelize.OpenFile(name)
	if err != nil {
		return err
	}
	defer f.Close()

	// 获取所有行数据
	rows, err := f.GetRows(defSheet)
	if err != nil {
		return err
	}

	// 计算最大行列数
	rowNum := len(rows)
	colNum := 0
	if maxs, ok := slicex.MaxFunc(rows, func(row []string) int { return len(row) }); ok {
		colNum = len(maxs)
	}

	// 提取关键信息：电压等级行、列和单位
	volRow, volCol, unit := extInfo(f, rowNum, colNum)
	volReg := regexp.MustCompile(voltagePattern)
	volSZReg := regexp.MustCompile(voltageSZPattern)

	// 获取标题行
	titles := slicex.NewFunc(colNum, func(i int) string { return mustCell(f, defSheet, volRow, i+1) })

	// 遍历数据行
	for row := volRow + 1; row <= rowNum; row++ {
		cell := mustCell(f, defSheet, row, volCol)
		// 匹配电压等级行
		if volReg.MatchString(cell) || volSZReg.MatchString(cell) {
			// 获取当前行所有单元格值
			values := slicex.NewFunc(colNum, func(i int) string {
				return mustCell(f, defSheet, row, i+1)
			})

			// 解析数据到结构体
			dr := initRow
			(&dr).evaluate(titles, values)
			(&dr).splitSZCategory()   // 处理深圳特殊情况
			(&dr).format(units[unit]) // 单位转换
			dlgd.Rows = append(dlgd.Rows, dr)
			dlgd.Comment = ""
			continue
		}
		// 收集备注信息
		dlgd.Comment += fmt.Sprintln(mustCell(f, defSheet, row, 1))
	}
	dlgd.Comment = strings.Trim(dlgd.Comment, "\n")

	return nil
}

// evaluate 将Excel行数据映射到结构体字段
//
// 参数:
//   - titles: 标题行
//   - values: 数据行
func (row *DlgdRow) evaluate(titles, values []string) {
	dlType := reflect.TypeOf(row).Elem()
	dlValue := reflect.ValueOf(row).Elem()

	// 遍历所有列
	for index, title := range titles {
		// 遍历结构体所有字段
		for i := 0; i < dlType.NumField(); i++ {
			// 获取字段的dlgd标签
			tag := dlType.Field(i).Tag.Get(dlgdTag)
			if len(tag) == 0 {
				continue
			}

			// 匹配标题和标签
			reg := regexp.MustCompile(tag)
			if reg.MatchString(title) {
				switch dlValue.Field(i).Kind() {
				case reflect.Float64:
					// 数值类型转换
					if value, ok := strconv.ParseFloat(values[index], 64); ok == nil {
						dlValue.Field(i).SetFloat(value)
					}
				case reflect.String:
					// 字符串类型处理（深圳特殊情况可能跨多列）
					if !strings.Contains(dlValue.Field(i).String(), values[index]) {
						s := dlValue.Field(i).String() + fieldValueSep + values[index]
						dlValue.Field(i).SetString(strings.Trim(s, fieldValueSep))
					}
				}
			}
		}
	}
}

// splitSZCategory 处理深圳特殊格式的分类数据
// 示例输入："一、工商业用电(101至3000kVA),10千伏高供低计(380V/220V计量),250kW·h及以下／千伏安·月"
// 输出:
//
//	Category: "一、工商业用电(101至3000kVA)"
//	Voltage: "10千伏高供低计(380V/220V计量)"
//	Stage: "250kW·h及以下／千伏安·月"
func (row *DlgdRow) splitSZCategory() {
	infos := strings.Split(row.Category, fieldValueSep)
	if len(infos) == 3 && len(row.Voltage) == 0 && len(row.Stage) == 0 {
		row.Category = infos[0]
		row.Voltage = infos[1]
		row.Stage = infos[2]
	}
}

// extInfo 提取Excel中的关键信息
//
// 返回:
//   - row: 电压等级所在行号
//   - col: 电压等级所在列号
//   - unit: 电价单位
func extInfo(f *excelize.File, rowNum int, colNum int) (row int, col int, unit string) {
	unit = yuanUnit // 默认单位
	for i := 1; i <= rowNum; i++ {
		for j := 1; j <= colNum; j++ {
			cell := mustCell(f, defSheet, i, j)
			// 定位电压等级列
			if cell == voltageName || cell == voltageSZName {
				row = i
				col = j
			}

			// 检测电价单位
			if strings.Contains(cell, fenUnit) {
				unit = fenUnit
			}
		}
	}
	return
}

// mustCell 安全获取单元格值
//
// 参数:
//   - f: Excel文件对象
//   - sheet: 工作表名称
//   - row: 行号(1-based)
//   - col: 列号(1-based)
//
// 返回:
//   - 单元格值，无效坐标返回空字符串
func mustCell(f *excelize.File, sheet string, row int, col int) string {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return ""
	}

	value, err := f.GetCellValue(sheet, cell)
	if err != nil {
		return ""
	}

	return regexp.MustCompile(`\s+`).ReplaceAllString(value, "")
	// return strings.TrimSpace(value)
}

// mustRange 安全获取单元格范围值
//
// 参数:
//   - f: Excel文件对象
//   - sheet: 工作表名称
//   - startRow: 起始行号(1-based)
//   - startCol: 起始列号(1-based)
//   - endRow: [含]结束行号(1-based)
//   - endCol: [含]结束列号(1-based)
//   - ignoreMerged: 是否忽略合并单元格
//
// 返回:
//   - 单元格范围值，无效坐标返回空字符串
func mustRange(f *excelize.File, sheet string, startRow, startCol, endRow, endCol int, ignoreMerged bool) string {
	if startRow > endRow || startCol > endCol {
		return ""
	}
	if _, err := f.GetSheetIndex(sheet); err != nil {
		return ""
	}

	var sb strings.Builder
	for i := startRow; i <= endRow; i++ {
		for j := startCol; j <= endCol; j++ {
			cell := mustCell(f, sheet, i, j)
			if !ignoreMerged || sb.Len() == 0 || !strings.HasSuffix(sb.String(), cell) {
				sb.WriteString(cell)
			}
		}
	}
	return sb.String()
}

// getMergedOrCellEndAxis 获取合并单元格或单元格的结束坐标
//
// 参数:
//   - cells: 合并单元格列表
//   - row: 行号(1-based)
//   - col: 列号(1-based)
//
// 返回:
//   - 结束行号、结束列号
func getMergedOrCellEndAxis(cells []excelize.MergeCell, row, col int) (int, int) {
	for _, cell := range cells {
		startCol, startRow, _ := excelize.CellNameToCoordinates(cell.GetStartAxis())
		endCol, endRow, _ := excelize.CellNameToCoordinates(cell.GetEndAxis())
		if row >= startRow && row <= endRow && col >= startCol && col <= endCol {
			return endRow, endCol
		}
	}

	return row, col
}
