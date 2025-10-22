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

func TestGetDefaultRegion(t *testing.T) {
	setup := testsetup.SetupTest(t)
	addrs := []string{
		"北京市西城区西直门外大街南方电网北京办事处",
		"上海市浦东新区陆家嘴环路南方电网上海联络处",
		"天津市河西区友谊路南方电网天津服务中心",
		"重庆市渝中区解放碑邹容路南方电网重庆分公司",
		"广东省广州市天河区黄埔大道中南方电网大厦",
		"广东省深圳市福田区深南中路南方电网深圳供电局",
		"云南省昆明市盘龙区北京路南方电网云南公司",
		"广西壮族自治区南宁市青秀区民族大道南方电网广西公司",
		"贵州省贵阳市南明区滨河路南方电网贵州公司",
		"海南省海口市美兰区海府路南方电网海南公司",
		"河北省石家庄市桥西区自强路南方电网河北电力服务处",
		"山西省太原市杏花岭区府西街南方电网山西业务部",
		"辽宁省沈阳市和平区南京街南方电网辽宁联络点",
		"吉林省长春市朝阳区人民大街南方电网吉林办事处",
		"黑龙江省哈尔滨市南岗区中山路南方电网黑龙江服务中心",
		"江苏省南京市鼓楼区中央路南方电网江苏联络处",
		"浙江省杭州市西湖区天目山路南方电网浙江办事处",
		"安徽省合肥市庐阳区长江中路南方电网安徽业务部",
		"福建省福州市鼓楼区五四路南方电网福建服务中心",
		"江西省南昌市东湖区阳明路南方电网江西联络处",
		"山东省济南市历下区泉城路南方电网山东办事处",
		"河南省郑州市金水区花园路南方电网河南业务部",
		"湖北省武汉市江汉区解放大道南方电网湖北服务中心",
		"湖南省长沙市芙蓉区五一大道南方电网湖南联络处",
		"四川省成都市锦江区东大街南方电网四川办事处",
		"陕西省西安市碑林区南大街南方电网陕西业务部",
		"甘肃省兰州市城关区张掖路南方电网甘肃服务中心",
		"青海省西宁市城中区西大街南方电网青海联络处",
		"内蒙古自治区呼和浩特市新城区新华大街南方电网内蒙古办事处",
		"宁夏回族自治区银川市兴庆区解放西街南方电网宁夏业务部",
		"新疆维吾尔自治区乌鲁木齐市天山区中山路南方电网新疆服务中心",
		"西藏自治区拉萨市城关区北京中路南方电网西藏联络处",
	}

	for _, addr := range addrs {
		region, err := setup.SvcCtx.AreaModel.Get95598Region(setup.Ctx, addr)
		if err != nil {
			t.Errorf("GetDefaultRegion() error = %v", err)
			continue
		}
		t.Logf("GetDefaultRegion() region = %v", region)
	}
}
