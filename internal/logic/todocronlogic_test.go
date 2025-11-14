package logic

import (
	"context"
	"testing"

	"seeccloud.com/edscron/internal/testsetup"
	"seeccloud.com/edscron/pkg/copierx"
	"seeccloud.com/edscron/pkg/cronx"
)

func TestDlgdHour(t *testing.T) {
	setup := testsetup.SetupTest(t)

	hours, err := setup.SvcCtx.DlgdHourModel.QueryAll(context.Background(), "福建", "闽发改规[2023]8号")
	if err != nil {
		t.Errorf("查询时段划分结果失败: %v", err)
	}

	cfg := cronx.DlgdConfig{
		Area:  "福建",
		Month: "2025年9月",
	}

	hors := []cronx.DlgdHour{}
	copierx.MustCopy(&hors, hours)
	rows := []cronx.DlgdRow{
		{},
	}
	cfg.AutoFill(&rows, &hors)

	if rows[0].SharpHour != "11:00-12:00,17:00-18:00" {
		t.Errorf("时段划分结果错误: %v", rows[0].SharpHour)
	}
}
