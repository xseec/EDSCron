package cronx

// 输出格式枚举类型
type OutputFormat string

const ()

const (
	CategorySep = ">" // 电价类别层级分隔符, 如"五、高壓及特高壓電力電價>(一)二段式時間電價"

	dateFormat = "2006-01-02" // 日期格式

	fieldValueSep = "," // 字段值分隔符

	formatExcel OutputFormat = "excel" // Excel格式输出
	formatWord  OutputFormat = "word"  // Word格式输出
	formatImage OutputFormat = "image" // 图片格式输出\

	taiwanAreaName = "台湾"
)
