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

type QuickStartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQuickStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuickStartLogic {
	return &QuickStartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 快速开始
func (l *QuickStartLogic) QuickStart(in *cron.QuickStartReq) (*cron.ResultRsp, error) {
	if err := expx.HasZeroError(in, "Address"); err != nil {
		return nil, err
	}

	crons := make([]model.Cron, 0)
	prv, cty := cronx.ExtractAddress(in.Address, true)
	address := model.Address{
		Province: prv,
		City:     cty,
	}

	crons = append(crons,
		model.NewCron(model.CategoryWeather, address),
		model.NewCron(model.CategoryHoliday, address),
	)

	if _, ok := cronx.CapitalWeather[prv]; ok {
		capital, err := l.svcCtx.AreaModel.GetProvincialCapital(l.ctx, prv)
		if err != nil {
			return nil, err
		}

		crons = append(crons,
			model.NewCron(model.CategoryWeather, model.Address{
				Province: prv,
				City:     capital,
			}),
		)
	}

	if prv == cronx.TaiwanAreaName {
		crons = append(crons,
			model.NewCron(model.CategoryTwCarbon, address),
			model.NewCron(model.CategoryTwdl, address),
		)
	} else {
		crons = append(crons,
			model.NewCron(model.CategoryCarbon, address),
			model.NewCron(model.CategoryReDlgd, address),
		)

		// 最后添加至crons
		addr, err := l.svcCtx.AreaModel.Get95598Address(l.ctx, in.Address)
		if err != nil {
			return nil, err
		}
		crons = append(crons, model.NewCron(model.CategoryDlgd, *addr))
	}

	// 批量添加任务
	_, err := l.svcCtx.CronModel.BatchInsert(l.ctx, crons)
	if err != nil {
		return nil, err
	}

	return &cron.ResultRsp{
		Message: fmt.Sprintf("快速开始, 添加%d个任务", len(crons)),
	}, nil
}
