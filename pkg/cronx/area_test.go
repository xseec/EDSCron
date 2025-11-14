package cronx

import (
	"testing"
)

func TestExtractAddress(t *testing.T) {
	tests := []struct {
		address  string
		province string
		city     string
	}{
		{
			address:  "广东省中山市士林电机",
			province: "广东",
			city:     "中山",
		},
		{
			address:  "广东省深圳市士林电机",
			province: "广东",
			city:     "深圳",
		},
		{
			address:  "江苏省苏州市士林电机",
			province: "江苏",
			city:     "苏州",
		},
		{
			address:  "上海市士林电机",
			province: "上海",
			city:     "上海",
		},
		{
			address:  "新竹县新丰乡中仑村中仑234号",
			province: "台湾",
			city:     "新竹",
		},
	}

	for _, test := range tests {
		province, city := ExtractAddress(test.address, true)
		if province != test.province || city != test.city {
			t.Fatalf("ExtractAddress failed, expect: %s, %s, got: %s, %s", test.province, test.city, province, city)
		}
	}
}
