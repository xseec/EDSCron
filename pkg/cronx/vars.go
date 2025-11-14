package cronx

// 输出格式枚举类型
type OutputFormat string

const (
	CategorySep = ">" // 电价类别层级分隔符, 如"五、高壓及特高壓電力電價>(一)二段式時間電價"

	fieldItemSep     = ";" // 字段项分隔符
	fieldSubSep      = "," // 字段值分隔符
	fieldKeyValueSep = ":" // 字段值分隔符

	formatExcel OutputFormat = "excel" // Excel格式输出
	formatWord  OutputFormat = "word"  // Word格式输出
	formatImage OutputFormat = "image" // 图片格式输出\

	TaiwanAreaName = "台湾"
	ChinaAreaName  = "中国"

	dlgdOne   = "单一制"
	dlgdTwo   = "两部制"
	dlgdLarge = "大工业用电"

	holidayDatePat  = `holiday:(\d+|(?:\p{Han}+(?:,|$))+)`
	highTempDatePat = `hi-temp:\d+(,\d+)?`
	weekendDatePat  = `weekend`
)
