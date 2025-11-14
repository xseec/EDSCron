package logic

import (
	"context"
	"fmt"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/vars"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddUserOptionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddUserOptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserOptionLogic {
	return &AddUserOptionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建用电档案
func (l *AddUserOptionLogic) AddUserOption(in *cron.UserOptionBody) (*cron.ResultRsp, error) {
	if err := expx.HasZeroError(in, "Account", "Category"); err != nil {
		return nil, err
	}

	option := &model.UserOption{}
	copierx.MustCopy(&option, in)

	rst, err := l.svcCtx.OptionModel.Insert(l.ctx, option)
	if err != nil {
		return nil, err
	}

	id, err := rst.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &cron.ResultRsp{
		Message: fmt.Sprintf("%s, id: %v", vars.SuccessMessage, id),
	}, nil
}
