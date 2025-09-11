# EDS定时任务服务（cron.rpc）

## 🔖简介

> EDS部分功能依赖于外部机构定期/不定期发布的数据，人为录入存在时效性、准确性和重复性困难，开发定时任务服务(cron.rpc)解决此类问题。

引用资源：

|类型标识|名称|数据源|
|-|-|-|
|dlgd|代理购电|[国家电网](https://95598.cn/osgweb/index)/[南方电网](https://95598.csg.cn/)|
|twdl|台湾电力|[台湾電力公司-電價表/電價日曆表](https://99z.top/https://www.taipower.com.tw/2289/2290/46940/46945/normalPost)|
|holiday|节假日|[各国假期日历](https://holidays-calendar.net/)|
|weather|天气预报|[中央气象台](http://www.nmc.cn/)|
|carbon|大陆碳排因子|[国家碳排因子库](https://data.ncsc.org.cn/factoryes/index)|
|tw-carbon|台湾碳排因子|[經濟部能源署-温室气体](https://www.moeaea.gov.tw/ecw/populace/content/SubMenu.aspx?menu_id=114)|
|pdf24|转Excel/Word|[免费、单元格底纹保留👍](https://tools.pdf24.org/zh/pdf-to-excel)|
|99z.top|代理访问|[免费、稳定](https://99z.top/)|
|aliyun|文件格式转换|[收费、文本识别准确率高](https://docmind.console.aliyun.com/doc-overview)|

## 🐼前提条件

政府类网站具有较强的反爬虫机制，用[chromedp](https://github.com/chromedp/chromedp)模拟人为操作，通过点击、跳转和选择等动作提取网页关键元素。

**运行依赖**：

- 本地需安装 Chrome/Chromium 或兼容浏览器（如 Edge）
- 构建docker镜像时需安装chromedp所需的依赖（Chromium + 必要库）

## 🎸接口列表

### 用电档案

- 基于用电档案获得电价和账单信息。当用户存在多个户号时，每个工程ID可以有多个用电档案，用<span style="background-color:#fff8e1;padding:2px 4px">area(工程区域、支路名称)</span>区分。  

（示例）厦门士林电机有两个户号：士林厂和成宇厂，每个户号下均含两个变压器，其<span style="background-color:#fff8e1;padding:2px 4px">capacity(合同容量)</span>为下辖变压器容量之和。

```json
{
    "account": "edsdemo",                            // 工程ID
    "area": "成宇厂",                                 // 工程区域、支路名称
    "category": "福建>工商业,两部制>1-10（20）千伏",    // 用电类别
    "powerFactor": 0.9,                              // [大陆]功率因数标准
    "capacity": 1260,                                // [大陆]合同容量
    "demand": 0,                                     // [大陆]合同需量
    "installedCap": 0,                               // [台湾]装置契约
    "regularCap": 0,                                 // [台湾]常规契约
    "nonSummerCap": 0,                               // [台湾]非夏季契约
    "semiPeakCap": 0,                                // [台湾]半尖峰契约
    "satSemiPeakCap": 0,                             // [台湾]周六半尖峰契约
    "offPeakCap": 0,                                 // [台湾]离峰契约
    "id": 2
}

```

### GetAvailableOptions（获取用电档案可选项）

- 基于用户地址获得用电档案可选项，地址应包含省份、城市、区县等信息
  
  - ✅ 福建省厦门市集美区孙坂南路92号

  - ✅ 台北市中山北路六段88號16樓（台湾地区可以从市级开始）

  - ❌ 厦门市集美区孙坂南路92号（大陆地区必须含省级信息）
  
```json
// Request
{
    "address": "福建省厦门市集美区孙坂南路92号"
}

// Response
{
    "categories": [
        "福建>工商业,两部制>1-10（20）千伏",,
        "福建>工商业,两部制>35千伏"
        "福建>工商业,两部制>110千伏",
        "福建>工商业,两部制>220千伏及以上"
    ],
    "powerFactors": [0.8, 0.85, 0.9]
}

```

### GetUserOption（获取用电档案）

```json
// Request
{
    "account":"edsdemo",
    "area":"成宇厂"
}
```

### 其他

#### AddUserOption（新增用电档案）| UpdateUserOption（更新用电档案）| DeleteUserOption（删除用电档案）

### GetCarbon（获取碳排因子）

- 碳排因子延迟发布，如2025年发布2024年数据；若指定年份无数据，返回最近年份数据。

```json
// Request
{
    "address": "福建省厦门市集美区孙坂南路92号",
    "year": 2025                                    // 可选，缺省时返回最新值
}

// Response
{
    "value": 0.4092                                // 实为2022年数据
}
```

### GetPrice（获取电价）

```json
// Request1
{
    "category":"福建>工商业,两部制>1-10（20）千伏",
    "time":"2025-08-01 00:00:00"
}

// Response1
{
    "name": "valley",        // 所属时段，可选值：deep, valley, flat, peak, sharp
    "price": 0.34861174,     // 时段电价
    "desc": "谷段",          // 时段描述, 可选值：深谷，谷段, 平段, 峰段, 尖段
    "color": "#27AE60",      // 时段色块
}

// Request2
{
    "category":"低壓電力電價>時間電價>三段式",
    "time":"2025-08-01 00:00:00"
}

// Response2
{
    "name": "weekdayOffPeak",  // 所属时段，可选值：weekdayPeak, weekdaySemiPeak, weekdayOffPeak, satSemiPeak, satOffPeak, sunOffPeak
    "price": 2.23,
    "desc": "平日离峰",         // 时段描述, 可选值：平日尖峰, 平日半尖峰, 平日离峰, 周六半尖峰, 周六离峰, 假日离峰
    "color": "#27AE60"
}

```

### GetMonthlyBill（获取月度账单）

```json
// Request
{
    "account": "edsdemo",
    "area": "成宇厂",
    "month": "2025-08",             // 账单周期，格式: 2006-01
    "ep30ms": [10.0, 20.0, 20.0],   // 正向有功分段值，每30分钟一个点，从每月1日0时起
    "eqTotal": 100.0                // 总无功（Q1+Q4象限）
}

// Response
{
    "fee": 10000.0,                 // 本期电费
    "basicFee": 1000.0,             // 基础费用
    "usageFee": 8000.0,             // 电量电费
    "pfFee": -1000.0,               // 功率因数调整电费：奖励(<0), 罚款(>0)
    "stageFee": 2000.0,             // 阶梯费用
    "usage": 20000.0,               // 本期电量
    "details": [                    // 分时电量电费列表
        {
            "name": "valley",
            "price": 0.34861174,
            "desc": "谷段",
            "color": "#27AE60",
            "usage": 30000.0
        }
    ]
}

```
