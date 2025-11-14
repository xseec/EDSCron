package logic

import (
	"testing"

	"seeccloud.com/edscron/internal/testsetup"
)

func TestGetHolidayOffs(t *testing.T) {
	setup := testsetup.SetupTest(t)
	tests := []struct {
		date string
		want int64
	}{
		{
			date: "2026-10-03",
			want: 7,
		},
		{
			date: "2026-02-20",
			want: 9,
		},
	}

	for _, tt := range tests {
		size, err := setup.SvcCtx.HolidayModel.GetHolidayOffSizeByAreaDate(setup.Ctx, "china", tt.date)
		if err != nil {
			t.Errorf("FindAllHolidayOffByAreaDate() error = %v", err)
			return
		}

		if size != tt.want {
			t.Errorf("FindAllHolidayOffByAreaDate() = %v, want %v", size, tt.want)
			return
		}
	}
}
