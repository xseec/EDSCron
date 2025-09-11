package cronx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStageFee(t *testing.T) {
	tests := []struct {
		name  string
		stage string
		total float64
		want  float64
	}{
		{
			name:  "表灯-住宅用",
			stage: "120度以下部分:1.68,121~330度部分:2.45,331~500度部分:3.70,501~700度部分:5.04,701~1000度部分:6.24,1001度以上部分:8.46",
			total: 400,
			want:  975 - 400*1.68,
		},
		{
			name:  "表灯-住宅以外非营业用",
			stage: "120度以下部分:1.68,121~330度部分:2.45,331~500度部分:3.70,501~700度部分:5.04,701~1000度部分:6.24,1001度以上部分:8.46",
			total: 800,
			want:  2977 - 800*1.68,
		},
		{
			name:  "表灯-营业用",
			stage: "330度以下部分:2.61,331~700度部分:3.66,701~1500度部分:4.46,1501~3000度部分:7.08,3001度以上部分:7.43",
			total: 1800,
			want:  7908 - 1800*2.61,
		},
		{
			name:  "表灯-简易型时间电价",
			stage: "超過2000度之部分:加1.02",
			total: 3000,
			want:  (3000 - 2000) * 1.02,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := GetStageFee(test.stage, test.total)
			if !assert.InEpsilon(t, test.want, got, 0.01) {
				t.Errorf("GetStageFee(%v) = %v, want %v", test.stage, got, test.want)
			}
		})
	}
}

func TestGetStagePrice(t *testing.T) {
	tests := []struct {
		total int
		want  float64
	}{
		{
			total: 120,
			want:  1.68,
		},
		{
			total: 200,
			want:  2.45,
		},
		{
			total: 400,
			want:  3.7,
		},
		{
			total: 600,
			want:  5.04,
		},
		{
			total: 800,
			want:  6.24,
		},
		{
			total: 1001,
			want:  8.46,
		},
	}

	for _, test := range tests {
		got, _ := GetStagePrice("120度以下部分:1.68,121~330度部分:2.45,331~500度部分:3.70,501~700度部分:5.04,701~1000度部分:6.24,1001度以上部分:8.46", ",", ":", test.total)
		if got != test.want {
			t.Errorf("GetStagePrice(%v) = %v, want %v", test.total, got, test.want)
		}
	}
}
