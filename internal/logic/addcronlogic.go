package logic

import (
	"context"
	"fmt"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/vars"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type AddCronLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCronLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCronLogic {
	return &AddCronLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建任务
func (l *AddCronLogic) AddCron(in *cron.CronBody) (*cron.ResultRsp, error) {
	if err := svc.Format(in); err != nil {
		return nil, err
	}

	var data model.Cron
	copier.Copy(&data, in)
	result, err := l.svcCtx.CronModel.InsertOrIgnore(l.ctx, &data)
	if err != nil {
		return nil, err
	}

	var id any
	id, err = result.LastInsertId()
	if err != nil {
		id = "未知"
	}

	l.svcCtx.StartCron()
	return &cron.ResultRsp{
		Message: fmt.Sprintf("%s, id: %v", vars.SuccessMessage, id),
	}, nil
}
