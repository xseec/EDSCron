package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"
	"seeccloud.com/edscron/pkg/x/slicex"
	"seeccloud.com/edscron/pkg/x/stringx"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPriceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPriceLogic {
	return &GetPriceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取电价
func (l *GetPriceLogic) GetPrice(in *cron.PriceReq) (*cron.PriceRsp, error) {
	if err := expx.HasZeroError(in, "Category", "Time"); err != nil {
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

	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()
	all, err := l.svcCtx.DlgdModel.FindAllByAreaStartTimeCategoryVoltage(l.ctx, infos[0], start, infos[1], infos[2])
	if err != nil {
		return nil, err
	}

	if len(*all) == 0 {
		return nil, fmt.Errorf("未找到%s>%d年%d月>%s>%s电价表", infos[0], t.Year(), t.Month(), infos[1], infos[2])
	}

	comment := ""
	one := (*all)[0]
	if len(*all) > 1 {
		one = slicex.FirstOrDefFunc(*all, one, func(o model.Dlgd) bool {
			return stringx.ContainsAny(o.Stage, "以下", "<", "<=", "≤")
		})
		comment = fmt.Sprintf("category(%s)存在阶梯电价，返回值采用(%s)档，仅供参考", in.Category, one.Stage)
	}

	rsp := l.svcCtx.GetPrice(t, &one)
	rsp.Comment = comment
	return rsp, nil
}
