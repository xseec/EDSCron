package logic

import (
	"context"
	"fmt"
	"time"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/cronx"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWeathersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWeathersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWeathersLogic {
	return &GetWeathersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取天气预报列表
func (l *GetWeathersLogic) GetWeathers(in *cron.WeathersReq) (*cron.WeathersRsp, error) {
	if err := expx.HasZeroError(in, "Address"); err != nil {
		return nil, err
	}

	if in.Date == "" {
		in.Date = time.Now().Format("2006-01-02")
	}

	if in.Size <= 0 {
		in.Size = 1
	}

	_, city := cronx.ExtractAddress(in.Address, true)
	if len(city) == 0 {
		return nil, fmt.Errorf("依提供地址无法筛查市级信息, Address: %s", in.Address)
	}

	more, err := l.svcCtx.WeatherModel.FindMoreByDateCity(l.ctx, in.Date, city, in.Size)
	if err != nil {
		return nil, err
	}

	var weas []*cron.Weather
	for _, w := range *more {
		var wea cron.Weather
		copier.Copy(&wea, w)
		weas = append(weas, &wea)
	}

	return &cron.WeathersRsp{
		Weathers: weas,
	}, nil
}
