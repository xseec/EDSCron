package logic

import (
	"testing"

	"seeccloud.com/edscron/internal/testsetup"
)

func TestGetHiTempSize(t *testing.T) {
	setup := testsetup.SetupTest(t)
	tests := []struct {
		city   string
		date   string
		hiTemp float64
		want   int64
	}{
		{
			city:   "厦门",
			date:   "2025-11-15",
			hiTemp: 30,
			want:   0,
		},
		{
			city:   "厦门",
			date:   "2025-11-15",
			hiTemp: 25,
			want:   2,
		},
		{
			city:   "厦门",
			date:   "2025-11-15",
			hiTemp: 20,
			want:   11,
		},
	}

	for _, tt := range tests {
		got, err := setup.SvcCtx.WeatherModel.FindHiTempSize(setup.Ctx, tt.date, tt.city, tt.hiTemp)
		if err != nil {
			t.Errorf("FindHiTempSize(%v, %v, %v) = %v, want %v", tt.city, tt.date, tt.hiTemp, got, tt.want)
		}

		if got != tt.want {
			t.Errorf("GetHiTempSize(%v, %v, %v) = %v, want %v", tt.city, tt.date, tt.hiTemp, got, tt.want)
		}
	}
}
