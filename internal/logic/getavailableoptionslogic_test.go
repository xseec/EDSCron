package logic

import (
	"testing"

	"seeccloud.com/edscron/internal/testsetup"
)

func TestFindParent(t *testing.T) {
	tests := []struct {
		area    string
		want    string
		wantErr bool
	}{
		{
			area:    "中国",
			wantErr: true,
		},
		{
			area: "北京市",
			want: "北京",
		},
		{
			area: "福建",
			want: "中国",
		},
		{
			area: "福建省",
			want: "中国",
		},
		{
			area: "厦门市",
			want: "福建省",
		},
		{
			area: "厦门",
			want: "福建省",
		},
		{
			area: "集美区",
			want: "厦门市",
		},
		{
			area: "集美",
			want: "厦门市",
		},
	}

	setup := testsetup.SetupTest(t)

	for _, tt := range tests {
		t.Run(tt.area, func(t *testing.T) {
			got, err := setup.SvcCtx.AreaModel.FindParent(setup.Ctx, tt.area)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindParent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && got.Name != tt.want {
				t.Errorf("FindParent() got = %v, want %v", got, tt.want)
			}
		})
	}
}
