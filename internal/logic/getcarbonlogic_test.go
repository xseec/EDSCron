package logic

import (
	"testing"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/testsetup"
)

func TestGetCarbonLogic(t *testing.T) {
	tests := []struct {
		name    string
		req     *cron.CarbonReq
		want    *cron.CarbonRsp
		wantErr bool
	}{
		{
			name: "无地址，无年份",
			req: &cron.CarbonReq{
				Address: "",
			},
			wantErr: true,
		},
		{
			name: "厦门, 无年份",
			req: &cron.CarbonReq{
				Address: "福建省厦门市",
			},
			want: &cron.CarbonRsp{
				Value: 0.4092,
			},
		},
		{
			name: "厦门, 2025年",
			req: &cron.CarbonReq{
				Address: "福建省厦门市",
				Year:    2025,
			},
			want: &cron.CarbonRsp{
				Value: 0.4092,
			},
		},
		{
			name: "厦门, 2022年",
			req: &cron.CarbonReq{
				Address: "福建省厦门市",
				Year:    2022,
			},
			want: &cron.CarbonRsp{
				Value: 0.4092,
			},
		},
		{
			name: "厦门, 2021年",
			req: &cron.CarbonReq{
				Address: "福建省厦门市",
				Year:    2021,
			},
			want: &cron.CarbonRsp{
				Value: 0.4711,
			},
		},
		{
			name: "台湾, 2025年",
			req: &cron.CarbonReq{
				Address: "台湾",
				Year:    2025,
			},
			want: &cron.CarbonRsp{
				Value: 0.474,
			},
		},
		{
			name: "台湾, 2024年",
			req: &cron.CarbonReq{
				Address: "台湾",
				Year:    2024,
			},
			want: &cron.CarbonRsp{
				Value: 0.474,
			},
		},
		{
			name: "台湾, 2022年",
			req: &cron.CarbonReq{
				Address: "台湾",
				Year:    2022,
			},
			want: &cron.CarbonRsp{
				Value: 0.509,
			},
		},
	}

	setup := testsetup.SetupTest(t)
	l := NewGetCarbonLogic(setup.Ctx, setup.SvcCtx)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := l.GetCarbon(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCarbon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}
			if result.Value != tt.want.Value {
				t.Errorf("GetCarbon() result = %v, want %v", result, tt.want)
			}

		})
	}
}
