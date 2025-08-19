package cronx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/jinzhu/now"
	"seeccloud.com/edscron/pkg/x/slicex"
)

// 常量定义 - 气象数据API端点
const (
	provinceUrl = "http://www.nmc.cn/rest/province/all"         // 省级行政区划代码接口
	cityUrl     = "http://www.nmc.cn/rest/province/%s"          // 市级行政区划代码接口(需省级代码)
	weatherUrl  = "http://www.nmc.cn/rest/weather?stationid=%s" // 城市天气预报接口
)

// 数据结构定义

// Province 省级行政区划信息
type Province struct {
	Code string `json:"code"` // 行政区划代码，如AFJ
	Name string `json:"name"` // 省份名称，如"福建省"
	Url  string `json:"url"`  // 预报页面URL
}

// City 市级行政区划信息
type City struct {
	Code     string `json:"code"`     // 城市代码，如gDCDS
	Province string `json:"province"` // 所属省份名称
	City     string `json:"city"`     // 城市名称
	Url      string `json:"url"`      // 预报页面URL
}

// Predict 天气预报数据结构
type Predict struct {
	City    string        `json:"city"`    // 城市全称(省+市)
	Publish string        `json:"publish"` // 发布时间
	Dates   []DateWeather `json:"dates"`   // 每日天气预报
	Temps   []Temperature `json:"temps"`   // 温度趋势数据
}

// DateWeather 单日天气数据
type DateWeather struct {
	Date  string      `json:"date"`  // 日期(YYYY-MM-DD)
	Day   WeatherWind `json:"day"`   // 白天天气
	Night WeatherWind `json:"night"` // 夜间天气
}

// WeatherWind 天气和风力组合数据
type WeatherWind struct {
	Weather WeatherDetail `json:"weather"` // 天气详情
	Wind    Wind          `json:"wind"`    // 风力信息
}

// WeatherDetail 天气详情
type WeatherDetail struct {
	Info        string `json:"info"`        // 天气现象(晴/雨等)
	Temperature string `json:"temperature"` // 温度(字符串形式)
}

// Wind 风力信息
type Wind struct {
	Direct string `json:"direct"` // 风向
	Power  string `json:"power"`  // 风力等级
}

// Temperature 温度数据
type Temperature struct {
	Time string  `json:"time"`     // 日期(YYYY/MM/DD)
	Max  float64 `json:"max_temp"` // 最高气温
	Min  float64 `json:"min_temp"` // 最低气温
}

// WeatherConfig 天气查询配置
type WeatherConfig struct {
	Province string `json:"province"` // 省级名称
	City     string `json:"city"`     // 市级名称
}

// Weather 简化版天气数据(用于输出)
type Weather struct {
	Date         string  `json:"date"`          // 日期
	City         string  `json:"city"`          // 城市名称
	DayWeather   string  `json:"day_weather"`   // 白天天气现象
	DayTemp      float64 `json:"day_temp"`      // 白天温度(最高温)
	NightWeather string  `json:"night_weather"` // 夜间天气现象
	NightTemp    float64 `json:"night_temp"`    // 夜间温度(最低温)
}

// Run 执行天气查询任务
//
// 参数:
//   - m: 邮件配置(可选)
//
// 返回值:
//   - any: 查询结果
//   - error: 错误信息
func (w WeatherConfig) Run(m *MailConfig) (any, error) {
	var p Predict
	if err := getPredict(w.Province, w.City, &p); err != nil {
		return nil, fmt.Errorf("获取天气预报失败: %v", err)
	}

	var results []Weather
	for _, wea := range p.Dates {
		dayTemp, err1 := strconv.ParseFloat(wea.Day.Weather.Temperature, 64)
		nightTemp, err2 := strconv.ParseFloat(wea.Night.Weather.Temperature, 64)

		// 过滤无效数据(9999表示数据不可用)
		if err1 != nil || err2 != nil || dayTemp >= 100 || nightTemp <= -100 {
			continue
		}

		results = append(results, Weather{
			Date:         wea.Date,
			City:         w.City,
			DayWeather:   wea.Day.Weather.Info,
			DayTemp:      dayTemp,
			NightWeather: wea.Night.Weather.Info,
			NightTemp:    nightTemp,
		})
	}

	return emptyValueErr(m, w, &results)
}

// getPredict 获取天气预报主流程
func getPredict(province, city string, p *Predict) error {
	var prvCode, cityCode string

	// 1. 获取省级代码
	if err := getProvince(province, &prvCode); err != nil {
		return fmt.Errorf("获取省级代码失败: %v", err)
	}

	// 2. 获取城市代码
	if err := getCity(city, prvCode, &cityCode); err != nil {
		return fmt.Errorf("获取城市代码失败: %v", err)
	}

	// 3. 获取天气数据
	if err := getWeather(cityCode, p); err != nil {
		return fmt.Errorf("获取天气数据失败: %v", err)
	}

	return nil
}

// getMaxTempOf 获取指定日期的最高气温
func getMaxTempOf(province, city, date string, temp *float64) error {
	var pre Predict
	if err := getPredict(province, city, &pre); err != nil {
		return err
	}

	d, err := now.Parse(date)
	if err != nil {
		return fmt.Errorf("日期解析失败: %v", err)
	}
	dt := now.With(d).BeginningOfDay()

	// 从温度数据中查找匹配日期的记录
	if t, ok := slicex.FirstFunc(pre.Temps, func(t Temperature) bool {
		time, err := now.Parse(t.Time)
		return err == nil && now.With(time).BeginningOfDay() == dt
	}); ok {
		*temp = t.Max
	}

	return nil
}

// getProvince 获取省级行政区划代码
func getProvince(text string, code *string) error {
	resp, err := http.Get(provinceUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var provinces []Province
	if err = json.Unmarshal(buf, &provinces); err != nil {
		return err
	}

	// 查找匹配的省份
	if p, ok := slicex.FirstFunc(provinces, func(s Province) bool {
		return strings.Contains(s.Name, text)
	}); ok {
		*code = p.Code
		return nil
	}

	return fmt.Errorf("未找到匹配的省份: %s", text)
}

// getCity 获取市级行政区划代码
func getCity(text string, prvCode string, code *string) error {
	url := fmt.Sprintf(cityUrl, prvCode)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var cities []City
	if err = json.Unmarshal(buf, &cities); err != nil {
		return err
	}

	// 查找匹配的城市
	if c, ok := slicex.FirstFunc(cities, func(s City) bool {
		return strings.Contains(s.City, text)
	}); ok {
		*code = c.Code
		return nil
	}

	return fmt.Errorf("未找到匹配的城市: %s", text)
}

// getWeather 获取城市天气预报
func getWeather(text string, pre *Predict) error {
	url := fmt.Sprintf(weatherUrl, text)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 使用simplejson处理复杂的JSON结构
	res, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return err
	}

	// 解析预测数据
	pres := res.Get("data").Get("predict")
	pre.City = pres.Get("station").Get("province").MustString() + pres.Get("station").Get("city").MustString()
	pre.Publish = pres.Get("publish_time").MustString()

	// 解析每日天气详情
	if buf, err := pres.Get("detail").MarshalJSON(); err == nil {
		json.Unmarshal(buf, &pre.Dates)
	}

	// 解析温度趋势数据
	if buf, err := res.Get("data").Get("tempchart").MarshalJSON(); err == nil {
		json.Unmarshal(buf, &pre.Temps)
	}

	return nil
}
