package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/vars"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserOptionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserOptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserOptionLogic {
	return &DeleteUserOptionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除用电档案
func (l *DeleteUserOptionLogic) DeleteUserOption(in *cron.DelReq) (*cron.ResultRsp, error) {
	err := l.svcCtx.OptionModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	return &cron.ResultRsp{
		Message: vars.SuccessMessage,
	}, nil
}
