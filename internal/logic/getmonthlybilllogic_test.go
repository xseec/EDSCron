package logic

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/testsetup"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/slicex"
)

func TestGetTwdlBill(t *testing.T) {
	tests := []struct {
		name     string
		month    string
		category string
		area     string
		basic    float64
		want     float64
	}{
		{
			name:     "新丰厂-25年10月，低压三段",
			month:    "2025-10",
			category: "低壓電力電價>時間電價>三段式",
			area:     "测试-新丰厂",
			basic:    500*173.20 + 50*34.6 + 262.5,
		},
		{
			name:     "新丰厂-25年09月，低压三段",
			month:    "2025-09",
			category: "低壓電力電價>時間電價>三段式",
			area:     "测试-新丰厂",
			basic:    500*236.20 + 50*34.6 + 262.5,
		},
		{
			name:     "新丰厂-25年01月，低压三段，离峰日x5",
			month:    "2025-01",
			category: "低壓電力電價>時間電價>三段式",
			area:     "测试-新丰厂",
			basic:    500*173.20 + 50*34.6 + 262.5,
		},
		{
			name:     "新丰厂-24年10月，低压三段，夸周期",
			month:    "2024-10",
			category: "低壓電力電價>時間電價>三段式",
			area:     "测试-新丰厂",
			basic:    500*173.20 + 50*34.6 + 262.5,
		},
		{
			name:     "新丰厂-25年10月，低压二段，需量契约",
			month:    "2025-10",
			category: "低壓電力電價>時間電價>二段式",
			area:     "测试-新丰厂-二段",
			basic:    500*173.20 + 100*34.60 + 50*34.60 + 262.5,
		},
		{
			name:     "新丰厂-25年10月，低压二段，装置契约",
			month:    "2025-10",
			category: "低壓電力電價>時間電價>二段式",
			area:     "测试-新丰厂-二段-装置",
			basic:    1000*137.50 + 105.00,
		},
		{
			name:     "新丰厂-25年10月，低压非时间电价，装置契约",
			month:    "2025-10",
			category: "低壓電力電價>非時間電價",
			area:     "测试-新丰厂-非时间",
			basic:    1000 * 137.50,
		},
		{
			name:     "新丰厂-25年10月，低压非时间电价，需量契约",
			month:    "2025-10",
			category: "低壓電力電價>非時間電價",
			area:     "测试-新丰厂-非时间-装置",
			basic:    500*173.20 + 50*173.20,
		},
		{
			name:     "表灯-25年10月，非时间，阶梯电价",
			month:    "2025-10",
			category: "表燈(住商)電價>非時間電價>營業用",
			area:     "测试-表灯-非时间",
		},
		{
			name:     "表灯-25年10月，简易二段式，阶梯电价",
			month:    "2025-10",
			category: "表燈電價>時間電價>簡易型時間電價>二段式",
			area:     "测试-表灯-简易二段式",
		},
	}

	setup := testsetup.SetupTest(t)
	l := NewGetMonthlyBillLogic(setup.Ctx, setup.SvcCtx)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			eps, fee := randomTwdl(test.month + ">" + test.category)
			if strings.Contains(test.name, "表灯") {
				usage := 0.0
				slicex.EachFunc(eps, func(ep float64) { usage += ep })
				switch test.name {
				case "表灯-25年10月，非时间，阶梯电价":
					fee = 300*2.18 + 370*3 + 800*3.61 + 1500*5.56 + (usage-3000)*5.83
				case "表灯-25年10月，简易二段式，阶梯电价":
					fee += (usage - 2000) * 1.02
				}

			}

			test.want = fee + test.basic
			req := cron.BillReq{
				Account: "edsdemo",
				Area:    test.area,
				Month:   test.month,
				Ep30Ms:  eps,
			}

			got, err := l.GetMonthlyBill(&req)

			if err != nil {
				t.Errorf("GetBill() error = %v", err)
				return
			}

			if !assert.InEpsilon(t, got.Fee, test.want, 0.005) {
				t.Errorf("GetBill() = %v, want %v", got.Fee, test.want)
			}
		})
	}

}

func TestGetDlgdBill(t *testing.T) {
	tests := []struct {
		name     string
		month    string
		category string
		area     string
		sharp    float64
		peak     float64
		flat     float64
		valley   float64
		eq1      float64
		eq2      float64
		want     float64
	}{
		{
			name:     "士林厂-6月",
			month:    "2025-06",
			category: "福建>工商业,两部制>1-10（20）千伏",
			area:     "士林厂",
			peak:     107037,
			flat:     91435,
			valley:   35798,
			eq1:      42720,
			eq2:      11670,
			want:     194951.14,
		},
		{
			name:     "士林厂-7月",
			month:    "2025-07",
			category: "福建>工商业,两部制>1-10（20）千伏",
			area:     "士林厂",
			sharp:    31568,
			peak:     83712,
			flat:     98445,
			valley:   40109,
			eq1:      56880,
			eq2:      5700,
			want:     200485.3,
		},
		{
			name:     "士林厂-8月",
			month:    "2025-08",
			category: "福建>工商业,两部制>1-10（20）千伏",
			area:     "士林厂",
			sharp:    32719,
			peak:     86671,
			flat:     101769,
			valley:   42436,
			eq1:      52110,
			eq2:      9570,
			want:     199368.96,
		},
		{
			name:     "成宇厂-2月",
			month:    "2025-02",
			category: "福建>工商业,两部制>1-10（20）千伏",
			area:     "成宇厂",
			sharp:    0,
			peak:     82430,
			flat:     79197,
			valley:   59391,
			eq1:      27717,
			eq2:      12683,
			want:     173633.55,
		},
		{
			name:     "成宇厂-7月",
			month:    "2025-07",
			category: "福建>工商业,两部制>1-10（20）千伏",
			area:     "成宇厂",
			sharp:    27962,
			peak:     83169,
			flat:     105610,
			valley:   79344,
			eq1:      34361,
			eq2:      6921,
			want:     217949.05,
		},
		{
			name:     "成宇厂-8月",
			month:    "2025-08",
			category: "福建>工商业,两部制>1-10（20）千伏",
			area:     "成宇厂",
			sharp:    27531,
			peak:     81516,
			flat:     104580,
			valley:   77487,
			eq1:      36472,
			eq2:      8051,
			want:     207711.15,
		},
	}

	setup := testsetup.SetupTest(t)
	l := NewGetMonthlyBillLogic(setup.Ctx, setup.SvcCtx)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eps := randomEps(l, test.month, test.category, test.sharp, test.peak, test.flat, test.valley, 0)
			req := cron.BillReq{
				Account: "edsdemo",
				Area:    test.area,
				Month:   test.month,
				Ep30Ms:  eps,
				EqTotal: test.eq1 + test.eq2,
			}

			got, err := l.GetMonthlyBill(&req)

			if err != nil {
				t.Errorf("GetBill() error = %v", err)
				return
			}

			if !assert.InEpsilon(t, got.Fee, test.want, 0.005) {
				t.Errorf("GetBill() = %v, want %v", got.Fee, test.want)
			}
		})
	}
}

func randomTwdl(sheet string) (eps []float64, fee float64) {
	f, err := excelize.OpenFile("../../temp/新丰厂-模拟电量电价.xlsx")
	if err != nil {
		return
	}
	defer f.Close()

	rowss, err := f.GetRows(sheet)
	if err != nil {
		return
	}

	// excel-sheet
	// 生效日期 | 生效月份 | 生效星期 | 0:00 0:30 ······ 23:30 (电价x48)
	// ······
	// 生效日期 | 生效月份 | 生效星期 | 0:00 0:30 ······ 23:30 (电价x48)
	// -------------------- 标题行 ----------------------------------
	// ------- | 离峰日期 |   01号  | 0:00 0:30 ······ 23:30 (电量x48)
	// ······
	// ------- | 离峰日期 |   31号  | 0:00 0:30 ······ 23:30 (电量x48)
	colLen := 51
	hourStartCol := 3
	priceStartTimeCol := 0
	priceMonthsCol := 1
	priceWeekdaysCol := 2
	usageDateCol := 2
	offPeakText := "离峰日"

	if len(rowss) == 0 || len(rowss[0]) != colLen {
		return
	}

	priceLen := 0
	for i := range rowss {
		rows := rowss[i]
		if len(rows) == 0 || len(rows[0]) != 0 {
			priceLen = i
			continue
		}

		priceRow := getPriceRow(rowss, rows, priceLen, priceStartTimeCol, priceMonthsCol, priceWeekdaysCol, usageDateCol, offPeakText)
		if priceRow == -1 {
			return
		}

		prices := rowss[priceRow]

		for j := hourStartCol; j < len(rows); j++ {
			ep, _ := strconv.ParseFloat(rows[j], 64)
			price, _ := strconv.ParseFloat(prices[j], 64)
			eps = append(eps, ep)
			fee += ep * price
		}

		// fmt.Printf("Test-randomTwdl StartTime: %s, Date: %s, Fee: %d\n", rowss[priceRow][priceStartTimeCol], rows[usageDateCol], int(fee))
	}

	return
}

func getPriceRow(rowss [][]string, rows []string, priceLen, priceStartTimeCol, priceMonthsCol, priceWeekdaysCol, usageDateCol int, offPeakText string) int {
	result := -1
	date, err := time.ParseInLocation("2006-01-02", rows[usageDateCol], time.Local)
	if err != nil {
		return result
	}

	isOffPeakDay := rows[1] == offPeakText
	for i := range priceLen {
		startTime, _ := time.ParseInLocation("2006-01-02", rowss[i][priceStartTimeCol], time.Local)

		if !date.Before(startTime) {
			// 夏月、非夏月
			if strings.Contains(rowss[i][priceMonthsCol], fmt.Sprintf("%d", int(date.Month()))) {
				// 离峰日优先级高于星期
				if isOffPeakDay {
					// Sunday("0")
					if strings.Contains(rowss[i][priceWeekdaysCol], "0") {
						result = i
					}
					continue
				}

				// 星期
				if strings.Contains(rowss[i][priceWeekdaysCol], fmt.Sprintf("%d", int(date.Weekday()))) {
					result = i
				}
			}
		}
	}

	return result
}

// randomEps 基于分时电量总值随机生成分时电量
func randomEps(l *GetMonthlyBillLogic, month, category string, sharp, peak, flat, valley, deep float64) []float64 {

	splits := strings.Split(category, cronx.CategorySep)
	if len(splits) != 3 {
		return nil
	}

	t, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return nil
	}

	area, category, voltage := splits[0], splits[1], splits[2]
	one, err := l.svcCtx.DlgdModel.FindFirstByAreaStartTimeCategoryVoltage(l.ctx, area, t.Format(vars.DatetimeFormat), category, voltage)
	if err != nil {
		return nil
	}

	deepIdxs := cronx.GetHalfHourIndexs(one.DeepHour)
	peakIdxs := cronx.GetHalfHourIndexs(one.PeakHour)
	flatIdxs := cronx.GetHalfHourIndexs(one.FlatHour)
	valleyIdxs := cronx.GetHalfHourIndexs(one.ValleyHour)
	sharpIdxs := cronx.GetHalfHourIndexs(one.SharpHour)
	if len(deepIdxs)+len(peakIdxs)+len(flatIdxs)+len(valleyIdxs)+len(sharpIdxs) != 48 {
		return nil
	}

	dayOfMonth := t.AddDate(0, 1, -1).Day()
	total := dayOfMonth * 48
	perDeep := deep / float64(len(deepIdxs)*dayOfMonth)
	perPeak := peak / float64(len(peakIdxs)*dayOfMonth)
	perFlat := flat / float64(len(flatIdxs)*dayOfMonth)
	perValley := valley / float64(len(valleyIdxs)*dayOfMonth)
	perSharp := sharp / float64(len(sharpIdxs)*dayOfMonth)

	eps := []float64{}

	for i := range total {
		scale := rand.Float64()*0.1 + 0.95
		ep := 0.0
		if slicex.Contains(deepIdxs, i%48) {
			if deep <= 1.5*perDeep {
				ep = deep
			} else {
				ep = math.Round(scale * perDeep)
			}

			deep -= ep
		}

		if slicex.Contains(peakIdxs, i%48) {
			if peak <= 1.5*perPeak {
				ep = peak
			} else {
				ep = math.Round(scale * perPeak)

			}

			peak -= ep
		}

		if slicex.Contains(flatIdxs, i%48) {
			if flat <= 1.5*perFlat {
				ep = flat
			} else {
				ep = math.Round(scale * perFlat)

			}

			flat -= ep
		}

		if slicex.Contains(valleyIdxs, i%48) {
			if valley <= 1.5*perValley {
				ep = valley
			} else {
				ep = math.Round(scale * perValley)
			}

			valley -= ep
		}

		if slicex.Contains(sharpIdxs, i%48) {
			if sharp <= 1.5*perSharp {
				ep = sharp
			} else {
				ep = math.Round(scale * perSharp)
			}

			sharp -= ep
		}

		eps = append(eps, ep)
	}

	return eps

}
