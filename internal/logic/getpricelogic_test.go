package logic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/testsetup"
)

func TestGetPriceLogic(t *testing.T) {
	tests := []struct {
		name    string
		req     *cron.PriceReq
		want    float64
		wantErr bool
	}{
		{
			name:    "电价类别或时间信息不完整",
			req:     &cron.PriceReq{},
			wantErr: true,
		},
		{
			name: "时间格式错误",
			req: &cron.PriceReq{
				Time: "2023-01-01 12:00",
			},
			wantErr: true,
		},
		{
			name: "电价类别格式错误",
			req: &cron.PriceReq{
				Category: "北京",
				Time:     "2023-01-01 12:00:00",
			},
			wantErr: true,
		},
		{
			name: "福建，谷段",
			req: &cron.PriceReq{
				Category: "福建>工商业,两部制>1-10（20）千伏",
				Time:     "2025-08-01 00:00:00",
			},
			want: 0.348612,
		},
		{
			name: "福建，平段",
			req: &cron.PriceReq{
				Category: "福建>工商业,两部制>1-10（20）千伏",
				Time:     "2025-08-01 08:00:00",
			},
			want: 0.584169,
		},
		{
			name: "福建，峰段",
			req: &cron.PriceReq{
				Category: "福建>工商业,两部制>1-10（20）千伏",
				Time:     "2025-08-01 10:00:00",
			},
			want: 0.801031,
		},
		{
			name: "福建，尖段",
			req: &cron.PriceReq{
				Category: "福建>工商业,两部制>1-10（20）千伏",
				Time:     "2025-08-01 11:00:00",
			},
			want: 0.883289,
		},
		{
			name: "广州，高温，尖段",
			req: &cron.PriceReq{
				Category: "广州/珠海/佛山/中山/东莞>工商业用电,两部制>1-10（20）千伏",
				Time:     "2025-06-08 11:00:00",
			},
			want: 1.43077,
		},
		{
			name: "广州，常温，峰段",
			req: &cron.PriceReq{
				Category: "广州/珠海/佛山/中山/东莞>工商业用电,两部制>1-10（20）千伏",
				Time:     "2025-06-07 11:00:00",
			},
			want: 1.15017,
		},
		{
			name: "杭州，劳动节，深谷",
			req: &cron.PriceReq{
				Category: "浙江>两部制,一般工商业用电>1~10(20)千伏",
				Time:     "2025-05-05 11:00:00",
			},
			want: 0.1264,
		},
		{
			name: "杭州，非节日，谷时",
			req: &cron.PriceReq{
				Category: "浙江>两部制,一般工商业用电>1~10(20)千伏",
				Time:     "2025-05-06 11:00:00",
			},
			want: 0.2844,
		},
		{
			name: "上海，周末，深谷",
			req: &cron.PriceReq{
				Category: "上海>大工业用电,两部制>10千伏",
				Time:     "2025-05-18 00:00:00",
			},
			want: 0.2062,
		},
		{
			name: "上海，周间，谷时",
			req: &cron.PriceReq{
				Category: "上海>大工业用电,两部制>10千伏",
				Time:     "2025-05-19 00:00:00",
			},
			want: 0.4019,
		},
		{
			name: "江苏，2月无值就近取3月",
			req: &cron.PriceReq{
				Category: "江苏>工商业用电,两部制>1-10（20）千伏",
				Time:     "2025-02-15 00:00:00",
			},
			want: 0.2714,
		},
		{
			name: "台湾，表灯，夏月",
			req: &cron.PriceReq{
				Category: "表燈(住商)電價>非時間電價>營業用",
				Time:     "2025-06-01 00:00:00",
			},
			want: 2.61,
		},
		{
			name: "台湾，表灯，非夏月",
			req: &cron.PriceReq{
				Category: "表燈(住商)電價>非時間電價>營業用",
				Time:     "2025-10-01 00:00:00",
			},
			want: 2.18,
		},
		{
			name: "台湾，低压，非时间，夏月",
			req: &cron.PriceReq{
				Category: "低壓電力電價>非時間電價",
				Time:     "2025-06-01 00:00:00",
			},
			want: 4.08,
		},
		{
			name: "台湾，低压，非时间，非夏月",
			req: &cron.PriceReq{
				Category: "低壓電力電價>非時間電價",
				Time:     "2025-10-01 00:00:00",
			},
			want: 3.87,
		},
		{
			name: "台湾，低压，二段式，夏月，平日尖峰",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>二段式",
				Time:     "2025-06-02 09:00:00",
			},
			want: 5.54,
		},
		{
			name: "台湾，低压，二段式，夏月，周六半尖峰",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>二段式",
				Time:     "2025-06-07 09:00:00",
			},
			want: 2.76,
		},
		{
			name: "台湾，低压，二段式，夏月，周日离峰",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>二段式",
				Time:     "2025-06-08 09:00:00",
			},
			want: 2.27,
		},
		{
			name: "台湾，低压，三段式，夏月，平日尖峰",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>三段式",
				Time:     "2025-06-02 16:00:00",
			},
			want: 8.12,
		},
		{
			name: "台湾，低压，三段式，夏月，平日半尖峰",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>三段式",
				Time:     "2025-06-02 09:00:00",
			},
			want: 5.02,
		},
		{
			name: "台湾，低压，三段式，夏月，周六半尖峰",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>三段式",
				Time:     "2025-06-07 16:00:00",
			},
			want: 2.50,
		},
		{
			name: "台湾，低压，三段式，夏月，周日离峰",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>三段式",
				Time:     "2025-06-08 16:00:00",
			},
			want: 2.23,
		},
		{
			name: "台湾，低压，三段式，非夏月，平日半尖峰",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>三段式",
				Time:     "2025-10-01 06:00:00",
			},
			want: 4.86,
		},
		{
			name: "台湾，低压，三段式，非夏月，离峰日",
			req: &cron.PriceReq{
				Category: "低壓電力電價>時間電價>三段式",
				Time:     "2025-10-10 06:00:00",
			},
			want: 2.12,
		},
	}

	setup := testsetup.SetupTest(t)
	l := NewGetPriceLogic(setup.Ctx, setup.SvcCtx)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := l.GetPrice(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if !assert.InEpsilon(t, tt.want, got.Value, 0.001) {
				t.Errorf("GetPrice() = %v, want %v", got.Value, tt.want)
			}
		})
	}
}
