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
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"
	"seeccloud.com/edscron/pkg/x/slicex"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBillLogic {
	return &GetBillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取账单
func (l *GetBillLogic) GetBill(in *cron.BillReq) (*cron.BillRsp, error) {
	if err := expx.HasZeroError(in, "Category", "PowerFactor", "Time", "Ep30Ms", "EqTotal"); err != nil {
		return nil, err
	}

	t, err := time.Parse(vars.DatetimeFormat, in.Time)
	if err != nil {
		return nil, err
	}

	infos := strings.Split(in.Category, model.CategorySep)
	if len(infos) < 3 {
		return nil, fmt.Errorf("req.Category格式错误, 正确格式: %s", model.CategoryFormatTip)
	}

	// 考虑存在阶梯电价
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()
	all, err := l.svcCtx.DlgdModel.FindAllByAreaStartTimeCategoryVoltage(l.ctx, infos[0], start, infos[1], infos[2])
	if err != nil {
		return nil, err
	}

	if len(*all) == 0 {
		return nil, fmt.Errorf("未找到%s>%d年%d月>%s>%s电价表", infos[0], t.Year(), t.Month(), infos[1], infos[2])
	}

	ranges := make(map[cronx.Range]int64)
	slicex.EachFunc(*all, func(d model.Dlgd) {
		r, _ := cronx.ExtractRange(d.Stage)
		ranges[r] = d.Id
	})

	var dlgd model.Dlgd
	var totalEp float64
	var usages = make(map[string]float64)
	var bills = make(map[string]float64)

	// 电量电费：尖、峰、平、谷各时段累加
	for i, ep := range in.Ep30Ms {
		tm := t.Add(time.Minute * 30 * time.Duration(i))
		totalEp += ep
		dlgd = getDlgdByEp(all, ranges, totalEp)
		price := l.svcCtx.GetPrice(tm, &dlgd)
		bills[price.Period] += price.Value * ep
		usages[price.Period] += ep
	}

	// 峰谷分时电量电费
	var totalBill float64
	details := make([]*cron.BillDetail, 0)
	for k, v := range usages {
		detail := cron.BillDetail{
			Period: k,
			Usage:  math.Round(v),
			Bill:   math.Round(bills[k]*100) / 100,
		}
		details = append(details, &detail)
		totalBill += detail.Bill
	}

	// 基础电费：需量*需量电价、容量*容量电价
	var basic float64
	if in.Demand > 0 {
		basic = in.Demand * dlgd.Demand
	} else if in.Capacity > 0 {
		basic = in.Capacity * dlgd.Capacity
	}
	basic = math.Round(basic*100) / 100

	// 力调电费：功率因素实际值→功率因素标准→调整系数，参与调整电费金额=基本电费+电量电费-政府基金及附加
	var pf = cronx.AdjustPowerFactorFee(in.PowerFactor, totalEp, in.EqTotal) * (totalBill + basic - totalEp*dlgd.Fund)
	pf = math.Round(pf*100) / 100

	return &cron.BillRsp{
		Total:       math.Round((basic+totalBill+pf)*10) / 10,
		Basic:       basic,
		Usage:       totalBill,
		PowerFactor: pf,
		Details:     details,
	}, nil
}

func getDlgdByEp(all *[]model.Dlgd, ranges map[cronx.Range]int64, ep float64) model.Dlgd {
	if len(*all) == 1 {
		return (*all)[0]
	}

	var id int64
	for k, v := range ranges {
		if k.Contains(ep) {
			id = v
			break
		}
	}

	dlgd, _ := slicex.FirstFunc(*all, func(d model.Dlgd) bool { return d.Id == id })

	return dlgd
}
