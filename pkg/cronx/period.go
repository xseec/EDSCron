package cronx

type Period struct {
	Name  string  `json:"name"`
	Desc  string  `json:"desc"`
	Color string  `json:"color"`
	Price float64 `json:"price"` // 电价
	Usage float64 `json:"usage"` // 电量
}

var (
	PeriodWeekdayPeak     = Period{Name: "weekdayPeak", Desc: "平日尖峰", Color: "#C0392B"}
	PeriodWeekdaySemiPeak = Period{Name: "weekdaySemiPeak", Desc: "平日半尖峰", Color: "#F39C12"}
	PeriodSatSemiPeak     = Period{Name: "satSemiPeak", Desc: "周六半尖峰", Color: "#3498DB"}
	// 台电平日、周六和假日离峰电价相同，但不排除未来可能形成差异，先用相同颜色
	PeriodWeekdayOffPeak = Period{Name: "weekdayOffPeak", Desc: "平日离峰", Color: "#27AE60"}
	PeriodSatOffPeak     = Period{Name: "satOffPeak", Desc: "周六离峰", Color: "#27AE60"}
	PeriodSunOffPeak     = Period{Name: "sunOffPeak", Desc: "假日离峰", Color: "#27AE60"}

	PeriodSharp  = Period{Name: "sharp", Desc: "尖段", Color: "#C0392B"}
	PeriodPeak   = Period{Name: "peak", Desc: "峰段", Color: "#F39C12"}
	PeriodFlat   = Period{Name: "flat", Desc: "平段", Color: "#3498DB"}
	PeriodValley = Period{Name: "valley", Desc: "谷段", Color: "#27AE60"}
	PeriodDeep   = Period{Name: "deep", Desc: "深谷", Color: "#006442"}

	dlgdPeriodDescs = []string{
		PeriodSharp.Desc,
		PeriodPeak.Desc,
		PeriodFlat.Desc,
		PeriodValley.Desc,
		PeriodDeep.Desc,
	}

	PeriodStandard = Period{Name: "standard", Desc: "基准", Color: "#3498DB"}
)
