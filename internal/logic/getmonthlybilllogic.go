package logic

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"
	"seeccloud.com/edscron/pkg/x/slicex"
	"seeccloud.com/edscron/pkg/x/timex"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMonthlyBillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMonthlyBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMonthlyBillLogic {
	return &GetMonthlyBillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取账单
func (l *GetMonthlyBillLogic) GetMonthlyBill(in *cron.BillReq) (*cron.BillRsp, error) {
	if err := expx.HasZeroError(in, "Account", "Time", "Ep30Ms"); err != nil {
		return nil, err
	}

	monthStart := timex.MustMonth(in.Month)

	option, err := l.svcCtx.OptionModel.FindOneByAccountNearlyArea(l.ctx, in.Account, in.Area)
	if err != nil {
		return nil, err
	}

	option, err = l.svcCtx.OptionModel.FindOneByAccountNearlyArea(l.ctx, in.Account, in.Area)
	if err != nil {
		return nil, err
	}

	ep30Ms := in.Ep30Ms
	ep30MsMax := monthStart.AddDate(0, 1, -1).Day() * 48
	if len(ep30Ms) > ep30MsMax {
		ep30Ms = ep30Ms[:ep30MsMax]
	}

	if slicex.Contains(cronx.TwdlCategories, option.Category) {
		return GetTwdlBill(l, *option, monthStart, ep30Ms)
	}

	return GetDlgdBill(l, *option, monthStart, ep30Ms, in.EqTotal)
}

func GetTwdlBill(l *GetMonthlyBillLogic, option model.UserOption, monthStart time.Time, ep30Ms []float64) (*cron.BillRsp, error) {
	periods := make(map[string]*cronx.Period, 0)
	var first *model.Twdl
	var usageFee, totalUsage float64
	for i := 0; i < len(ep30Ms); i += 48 {
		dayStart := monthStart.AddDate(0, 0, i/48)

		// 一个月中可能含多条电价条目，基本电费依据第一条
		one, err := l.svcCtx.TwdlModel.FindOneByDayStartTimeCategory(l.ctx, dayStart.Format(vars.DatetimeFormat), option.Category)
		if err != nil {
			return nil, err
		}

		if first == nil {
			first = one
		}

		holiday, _ := l.svcCtx.HolidayModel.FindOneByAreaDateCache(l.ctx, string(cronx.TaiwanArea), dayStart.Format(vars.DateFormat))

		for j := i; j < i+48 && j < len(ep30Ms); j++ {
			t := monthStart.Add(time.Minute * time.Duration(30*j))
			period := one.GetPrice(t.Format(vars.DatetimeFormat), holiday != nil && holiday.Category == string(cronx.HolidayPeakOff))
			totalUsage += ep30Ms[j]
			usageFee += ep30Ms[j] * period.Price
			if _, ok := periods[period.Name]; !ok {
				periods[period.Name] = &period
			} else {
				periods[period.Name].Usage += ep30Ms[j]
				// 一月同时段可能存在多种电价：夸夏、非夏月或电价换新，时段电价依据第一笔记录
				// periods[period.Name].Price = period.Price
			}
		}

		// fmt.Printf("Test-GetTwdlBill StartTime: %s, Date: %s, Fee: %d\n", one.StartTime.Format(vars.DateFormat), dayStart.Format(vars.DateFormat), int(usageFee))
	}

	// 阶梯电价
	stageFee := cronx.GetStageFee(first.Stage, totalUsage)

	// 基本电费 = 按户计收 + 契约电费
	var basicFee float64

	// 按户计收：需量契约按户计收优先级更高，暂未考虑单相按户计收
	if option.RegularCap > 0 && first.RegularCustomer > 0 {
		basicFee += first.RegularCustomer
	} else {
		basicFee += first.InstalledCustomer
	}

	// 契约电费
	if option.RegularCap > 0 {
		basicFee += first.RegularCap * option.RegularCap
		basicFee += first.NonSummerCap * option.NonSummerCap
		basicFee += first.OffPeakCap * option.OffPeakCap
		basicFee += first.SatSemiPeakCap * option.SatSemiPeakCap
		basicFee += first.SemiPeakCap * option.SemiPeakCap
	} else {
		basicFee += first.InstalledCap * option.InstalledCap
	}

	basicFee = math.Round(basicFee*100) / 100
	usageFee = math.Round(usageFee*100) / 100
	stageFee = math.Round(stageFee*100) / 100
	details := []*cron.BillDetail{}
	copierx.MustCopy(&details, periods)

	return &cron.BillRsp{
		Fee:      math.Round((basicFee+usageFee+stageFee)*10) / 10,
		BasicFee: basicFee,
		UsageFee: usageFee,
		StageFee: stageFee,
		Usage:    totalUsage,
		Details:  details,
	}, nil
}

func GetDlgdBill(l *GetMonthlyBillLogic, option model.UserOption, monthStart time.Time, ep30Ms []float64, eqTotal float64) (*cron.BillRsp, error) {
	infos := strings.Split(option.Category, cronx.CategorySep)
	if len(infos) < 3 {
		return nil, fmt.Errorf("req.Category格式错误, 正确格式: %s", model.CategoryFormatTip)
	}

	// 账单不考虑使用临近电价
	start := time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, time.Local).Format(vars.DatetimeFormat)
	one, err := l.svcCtx.DlgdModel.FindFirstByAreaStartTimeCategoryVoltage(l.ctx, infos[0], start, infos[1], infos[2])
	if err != nil {
		return nil, err
	}

	periods := make(map[string]*cronx.Period, 0)
	for i, ep := range ep30Ms {
		tm := monthStart.Add(time.Minute * 30 * time.Duration(i))
		period, err := l.svcCtx.GetDlgdPrice(tm, one)
		if err != nil {
			return nil, err
		}

		if v, ok := periods[period.Name]; ok {
			v.Usage += ep
			periods[period.Name] = v
		} else {
			period.Usage = ep
			periods[period.Name] = period
		}
	}

	// 峰谷分时电量电费
	var totalEp, usageFee float64
	details := make([]*cron.BillDetail, 0)
	for _, v := range periods {
		totalEp += v.Usage
		usageFee += v.Usage * v.Price
		var detail cron.BillDetail
		copierx.MustCopy(&detail, v)
		details = append(details, &detail)
	}

	// 基础电费：需量*需量电价、容量*容量电价
	var basic float64
	if option.Demand > 0 {
		basic = option.Demand * one.Demand
	} else if option.Capacity > 0 {
		basic = option.Capacity * one.Capacity
	}
	basic = math.Round(basic*100) / 100

	// 力调电费：功率因素实际值→功率因素标准→调整系数，参与调整电费金额=基本电费+电量电费-政府基金及附加
	var pf = cronx.AdjustPowerFactorFee(option.PowerFactor, totalEp, eqTotal) * (usageFee + basic - totalEp*one.Fund)
	pf = math.Round(pf*100) / 100

	// 代购购电阶梯电费仅存在于深圳，且计算较复杂暂不考虑
	return &cron.BillRsp{
		Fee:      math.Round((basic+usageFee+pf)*10) / 10,
		BasicFee: basic,
		UsageFee: usageFee,
		PfFee:    pf,
		StageFee: 0,
		Usage:    totalEp,
		Details:  details,
	}, nil
}
