package logic

import (
	"context"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/svc"
	"seeccloud.com/edscron/pkg/x/expx"

	"github.com/zeromicro/go-zero/core/logx"
)

type TodoCronLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTodoCronLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TodoCronLogic {
	return &TodoCronLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行任务
func (l *TodoCronLogic) TodoCron(in *cron.TodoCronReq) (*cron.ResultRsp, error) {
	if err := expx.HasZeroError(in, "Id"); err != nil {
		return nil, err
	}

	c, err := l.svcCtx.CronModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	// 只执行，不等待结果
	go l.svcCtx.Todo(c, in.Time)
	return &cron.ResultRsp{
		Message: "完成",
	}, nil
}
