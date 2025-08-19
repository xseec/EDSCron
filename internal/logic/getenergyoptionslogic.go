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

type GetEnergyOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEnergyOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnergyOptionsLogic {
	return &GetEnergyOptionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取能源可选项
func (l *GetEnergyOptionsLogic) GetEnergyOptions(in *cron.EnergyOptionsReq) (*cron.EnergyOptionsRsp, error) {
	if err := expx.HasZeroError(in, "Address"); err != nil {
		return nil, err
	}

	province, city := cronx.ExtractAddress(in.Address, true)
	categories, err := l.svcCtx.DlgdModel.FindCategoriesByAreas(l.ctx, city, province)
	if err != nil {
		return nil, err
	}

	if categories == nil {
		return nil, fmt.Errorf("所在区域用电类别为空, 请联系系统管理员. Address: %s", in.Address)
	}

	return &cron.EnergyOptionsRsp{
		Categories:   *categories,
		PowerFactors: cronx.PowerFactors[:],
	}, nil
}
