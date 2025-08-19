package logic

import (
	"context"
	"fmt"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type AddCarbonLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCarbonLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCarbonLogic {
	return &AddCarbonLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建碳排因子
func (l *AddCarbonLogic) AddCarbon(in *cron.AddCarbonReq) (*cron.ResultRsp, error) {
	err := expx.HasZeroError(in, "Area", "Year", "Value")
	if err != nil {
		return nil, err
	}

	carbon, _ := l.svcCtx.CarbonModel.FindOneByAreaYear(l.ctx, in.Area, in.Year)
	var c model.Carbon
	copier.Copy(&c, in)
	// 创建或更新
	if carbon == nil {
		rst, _ := l.svcCtx.CarbonModel.Insert(l.ctx, &c)
		if id, err := rst.LastInsertId(); err == nil {
			return &cron.ResultRsp{
				Message: fmt.Sprintf("%s, id: %d", vars.SuccessMessage, id),
			}, nil
		}
	} else {
		c.Id = carbon.Id
		err = l.svcCtx.CarbonModel.Update(l.ctx, &c)
	}

	if err != nil {
		return nil, err
	}

	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
