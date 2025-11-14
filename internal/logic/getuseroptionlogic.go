package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/copierx"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserOptionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserOptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOptionLogic {
	return &GetUserOptionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用电档案
func (l *GetUserOptionLogic) GetUserOption(in *cron.GetUserOptionReq) (*cron.UserOptionBody, error) {

	option, err := l.svcCtx.OptionModel.FindOneByAccountNearlyArea(l.ctx, in.Account, in.Area)
	if err != nil {
		return nil, err
	}

	var op cron.UserOptionBody
	copierx.MustCopy(&op, option)

	return &op, nil
}
