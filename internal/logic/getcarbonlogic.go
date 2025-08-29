package logic

import (
	"context"
	"fmt"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCarbonLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCarbonLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCarbonLogic {
	return &GetCarbonLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取碳排因子
func (l *GetCarbonLogic) GetCarbon(in *cron.CarbonReq) (*cron.CarbonRsp, error) {
	if err := expx.HasZeroError(in, "Address"); err != nil {
		return nil, err
	}

	province, _ := cronx.ExtractAddress(in.Address, true)
	if len(province) == 0 {
		return nil, fmt.Errorf("依提供地址无法筛查省级信息, Address: %s", in.Address)
	}

	var c *model.Carbon
	var err error
	if in.Year > 0 {
		c, err = l.svcCtx.CarbonModel.FindOneByAreaYear(l.ctx, province, in.Year)
		if err == nil {
			return &cron.CarbonRsp{
				Value: c.Value,
			}, nil
		}
	}

	// 依区域&年份查询，无数据，则依区域查询
	c, err = l.svcCtx.CarbonModel.FindOneByArea(l.ctx, province)
	if err != nil {
		// 依区域查询，无数据，则查更上一级
		c, err = l.svcCtx.CarbonModel.FindOneByArea(l.ctx, "中国")
	}

	if err != nil {
		return nil, err
	}

	// 避免频繁查询，仅缓存当天
	l.svcCtx.CarbonModel.SaveCacheOnlyToday(l.ctx, province, in.Year, c.Id)
	return &cron.CarbonRsp{
		Value: c.Value,
	}, nil
}
