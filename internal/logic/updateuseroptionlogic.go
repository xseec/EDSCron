package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/vars"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserOptionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserOptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserOptionLogic {
	return &UpdateUserOptionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用电档案
func (l *UpdateUserOptionLogic) UpdateUserOption(in *cron.UserOptionBody) (*cron.ResultRsp, error) {
	option := &model.UserOption{}
	copier.Copy(&option, in)
	err := l.svcCtx.OptionModel.Update(l.ctx, option)
	if err != nil {
		return nil, err
	}

	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
