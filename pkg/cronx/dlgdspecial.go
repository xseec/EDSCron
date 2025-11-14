package cronx

import "regexp"

var (
	// [2025-11] 江苏"单一制"被误拆行
	jiangsuArea        = "江苏"
	jiangsuCategoryReg = regexp.MustCompile(`(工商业用电,)(单|一|制|单一|一制)(,\w+)`)
)

func specialiseDlgd(dlgds []DlgdRow) []DlgdRow {
	values := make([]DlgdRow, 0, len(dlgds))
	for _, dlgd := range dlgds {
		if dlgd.Area != jiangsuArea {
			return dlgds
		}

		dlgd.Category = jiangsuCategoryReg.ReplaceAllString(dlgd.Category, `${1}单一制${3}`)
		values = append(values, dlgd)
	}
	return values
}
