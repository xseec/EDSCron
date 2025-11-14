package logic

import (
	"encoding/json"
	"testing"

	"seeccloud.com/edscron/cron"
	"seeccloud.com/edscron/internal/testsetup"
	"seeccloud.com/edscron/model"
)

func TestQuickStart(t *testing.T) {

	setup := testsetup.SetupTest(t)

	quickStartLogic := NewQuickStartLogic(setup.Ctx, setup.SvcCtx)
	quickStartLogic.QuickStart(&cron.QuickStartReq{
		Address: "福建省厦门市集美区孙坂南路92、94、96号",
	})
	quickStartLogic.QuickStart(&cron.QuickStartReq{
		Address: "台北市中山北路六段88號16樓",
	})
	quickStartLogic.QuickStart(&cron.QuickStartReq{
		Address: "广东省中山市士林电机",
	})
	quickStartLogic.QuickStart(&cron.QuickStartReq{
		Address: "广东省深圳市士林电机",
	})
	quickStartLogic.QuickStart(&cron.QuickStartReq{
		Address: "江苏省苏州市士林电机",
	})
	quickStartLogic.QuickStart(&cron.QuickStartReq{
		Address: "上海市士林电机",
	})

	crons, err := setup.SvcCtx.CronModel.FindAll(setup.Ctx)
	t.Log(*crons)
	if err != nil {
		t.Fatal(err)
	}

	type QuickStartTest struct {
		xiamenWeather    int
		fuzhouWeather    int
		taibeiWeather    int
		zhongshanWeather int
		shenzhenWeather  int
		guangzhouWeather int
		suzhouWeather    int
		nanjingWeather   int
		shanghaiWeather  int
		xiamenDlgd       int
		quanzhouDlgd     int
		fuzhouDlgd       int
		zhongshanDlgd    int
		shenzhenDlgd     int
		suzhouDlgd       int
		nanjingDlgd      int
		shanghaiDlgd     int
		twdl             int
		carbon           int
		twCarbon         int
		reDlgd           int
		holiday          int
	}

	test := QuickStartTest{
		xiamenWeather:    1,
		fuzhouWeather:    0,
		taibeiWeather:    1,
		zhongshanWeather: 1,
		shenzhenWeather:  1,
		guangzhouWeather: 1,
		suzhouWeather:    1,
		nanjingWeather:   0,
		shanghaiWeather:  1,
		xiamenDlgd:       0,
		quanzhouDlgd:     0,
		fuzhouDlgd:       1,
		zhongshanDlgd:    1,
		shenzhenDlgd:     1,
		suzhouDlgd:       0,
		nanjingDlgd:      1,
		shanghaiDlgd:     1,
		twdl:             1,
		carbon:           1,
		twCarbon:         1,
		reDlgd:           1,
		holiday:          1,
	}

	result := QuickStartTest{}
	for _, v := range *crons {
		switch v.Category {
		case string(model.CategoryWeather):
			var address model.Address
			if err := json.Unmarshal([]byte(v.Task), &address); err != nil {
				t.Fatal(err)
			}

			switch address.City {
			case "厦门":
				result.xiamenWeather++
			case "台北":
				result.taibeiWeather++
			case "福州":
				result.fuzhouWeather++
			case "中山":
				result.zhongshanWeather++
			case "深圳":
				result.shenzhenWeather++
			case "苏州":
				result.suzhouWeather++
			case "南京":
				result.nanjingWeather++
			case "上海":
				result.shanghaiWeather++
			case "广州":
				result.guangzhouWeather++
			}
		case string(model.CategoryDlgd):
			var address model.Address
			if err := json.Unmarshal([]byte(v.Task), &address); err != nil {
				t.Fatal(err)
			}

			switch address.City {
			case "厦门市":
				result.xiamenDlgd++
			case "泉州市":
				result.quanzhouDlgd++
			case "福州市":
				result.fuzhouDlgd++
			case "中山":
				result.zhongshanDlgd++
			case "深圳":
				result.shenzhenDlgd++
			case "苏州市":
				result.suzhouDlgd++
			case "南京市":
				result.nanjingDlgd++
			default:
				if address.Province == "上海" {
					result.shanghaiDlgd++
				}
			}
		case string(model.CategoryTwdl):
			result.twdl++
		case string(model.CategoryCarbon):
			result.carbon++
		case string(model.CategoryTwCarbon):
			result.twCarbon++
		case string(model.CategoryReDlgd):
			result.reDlgd++
		case string(model.CategoryHoliday):
			result.holiday++
		}
	}

	if result != test {
		t.Fatalf("result %v != test %v", result, test)
	}
}
