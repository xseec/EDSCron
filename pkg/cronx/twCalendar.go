package cronx

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
	"seeccloud.com/edscron/pkg/x/timex"
)

func ExcelizeCalendar(excelPath string) ([]string, error) {
	file, err := excelize.OpenFile(excelPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var dates []string
	sheets := file.GetSheetList()
	yearReg := regexp.MustCompile(`(\d{3,4})\s*年`)
	monthReg := regexp.MustCompile(`[一二三四五六七八九十\d]{1,2}\s*月`)

	for _, sheet := range sheets {
		rows, err := file.GetRows(sheet)
		if err != nil {
			return nil, err
		}

		var year, month int
		for rowIdx := 0; rowIdx < len(rows); rowIdx++ {
			row := rows[rowIdx]
			rowText := strings.Join(row, "")

			// 捕获年份行
			if subs := yearReg.FindStringSubmatch(rowText); len(subs) == 2 {
				year, _ = strconv.Atoi(subs[1])
				month = 0
				continue
			}

			// 跳过月标题行
			if monthReg.MatchString(rowText) {
				continue
			}

			endRow := rowIdx + 1
			for colIdx := 0; colIdx < len(row); colIdx++ {
				// 捕获周标题行列
				if mustRange(file, sheet, rowIdx+1, colIdx+1, rowIdx+1, colIdx+7, true) != "日一二三四五六" {
					continue
				}

				month++
				dates = append(dates, filterMonthlyDates(file, sheet, rowIdx+1, colIdx+1, year, month, &endRow)...)

				colIdx += 6
			}

			// 跳过日期所在行
			rowIdx = endRow - 1
		}

	}

	return dates, nil
}

func filterMonthlyDates(file *excelize.File, sheet string, startRow, startCol, year, month int, endRow *int) []string {

	dates := make([]string, 0)

	// 离峰日和周日单元格背景色相同，偏移2行首列定是周日，保证能取到颜色
	bgColor := mustBackground(file, sheet, startRow+2, startCol)
	// 一个月最多跨度不超过6周（1号为周六）
	lastDay := 0
Month:
	for i := startRow + 1; i <= startRow+6; i++ {
		for j := startCol; j <= startCol+6; j++ {
			day, err := strconv.Atoi(mustCell(file, sheet, i, j))
			// 月首、尾空单元格
			if err != nil {
				if lastDay == 0 {
					continue
				}

				break Month
			}

			lastDay = day
			*endRow = int(math.Max(float64(*endRow), float64(i)))

			// 跳过周日
			if j == startCol {
				continue
			}

			if mustBackground(file, sheet, i, j) == bgColor {
				dates = append(dates, fmt.Sprintf("%d-%02d-%02d", timex.Year(year), month, day))
			}
		}
	}

	return dates
}

func mustBackground(file *excelize.File, sheet string, row, col int) string {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return ""
	}

	styleId, err := file.GetCellStyle(sheet, cell)
	if err != nil {
		return ""
	}

	style, err := file.GetStyle(styleId)
	if err != nil {
		return ""
	}

	colors := style.Fill.Color
	if style.Fill.Type != "pattern" || len(colors) == 0 {
		return ""
	}

	return colors[0]
}
