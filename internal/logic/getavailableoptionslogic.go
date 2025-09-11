package logic

import (
	"context"
	"fmt"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAvailableOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAvailableOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailableOptionsLogic {
	return &GetAvailableOptionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取能源可选项
func (l *GetAvailableOptionsLogic) GetAvailableOptions(in *cron.AvailableOptionsReq) (*cron.AvailableOptionsRsp, error) {
	err := expx.HasZeroError(in, "Address")
	if err != nil {
		return nil, err
	}

	area := cronx.EitherChinaOrTaiwan(in.Address)
	var categories *[]string
	var factors []float64
	if area == cronx.TaiwanArea {
		categories, err = l.svcCtx.TwdlModel.FindCategories(l.ctx)
	} else {
		province, city := cronx.ExtractAddress(in.Address, true)
		categories, err = l.svcCtx.DlgdModel.FindCategoriesByAreas(l.ctx, city, province)
		factors = cronx.PowerFactors[:]
	}

	if err != nil {
		return nil, err
	}

	if categories == nil {
		return nil, fmt.Errorf("所在区域用电类别为空, 请联系系统管理员。 Address: %s", in.Address)
	}

	return &cron.AvailableOptionsRsp{
		Categories:   *categories,
		PowerFactors: factors,
	}, nil
}
