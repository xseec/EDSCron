package logic

import (
	"context"
	"fmt"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/model"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/vars"

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

	var data model.Cron
	copierx.MustCopy(&data, in)
	result, err := l.svcCtx.CronModel.BatchInsert(l.ctx, []model.Cron{data})
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
