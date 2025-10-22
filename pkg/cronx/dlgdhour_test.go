package cronx

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewPeriod(t *testing.T) {
	inputs := map[string]string{
		"安徽":    "2.根据皖发改价格〔2025〕302号文件规定，分时电价浮动比例：1、7-9、12月高峰上浮84.3%，其他月份高峰上浮74%，低谷下浮61.8%。时段划分：每年7-9月，每日高峰时段16:00-24:00；平段00:00-2:00，9:00-11:00，13:00-16:00；低谷时段2:00-9:00，11:00-13:00。1月、12月，每日高峰时段15:00-23:00；平段6:00-12:00，14:00-15:00；低谷时段12:00-14:00，23:00-次日6:00。其他月份，每日高峰时段6:00-8:00，16:00-22:00；平时段8:00-11:00，14:00-16:00，22:00-23:00；低谷时段11:00-14:00，23:00-次日6:00。深谷电价执行范围为用电容量315千伏安及以上执行工商业两部制电价和峰谷分时电价的工业用电，执行时段为每年3天及以上节假日（具体时间以国家公布为准）期间11:00-15:00，深谷电价在低谷电价基础上下浮20%。政策于2025年7月1日起执行。",
		"北京":    "2.分时电度电价在代理购电价格基础上根据京发改规〔2023〕11号文件峰谷比价关系和时段划分规定形成。具体时段划分为：高峰时段10:00-13:00，17:00-22:00；平时段7:00-10:00，13:00-17:00，22:00-23:00；低谷时段23:00-次日7:00，其中：夏季（7、8月）11:00-13:00，16:00-17:00；冬季（1、12月）18:00-21:00为尖峰时段。分时电价浮动比例为：不满1千伏单一制用电峰平谷电价比例为1.71:1:0.36；1千伏及以上单一制用电峰平谷电价比例为1.8:1:0.3；两部制用电峰平谷电价比例为1.6:1:0.4。尖峰电价在高峰电价基础上上浮20%。",
		"福建":    "2.根据《福建省发展和改革委员会关于完善分时电价政策的通知》（闽发改规〔2023〕8号）（以下简称《通知》）规定，电力现货市场未运行、市场交易合同未申报用电曲线以及市场电价峰谷比例低于《通知》要求的，结算时购电价格按《通知》规定的峰谷时段及浮动幅度执行，国家政策如有调整从其规定。全省峰时段为10:00—12:00，15:00—20:00，21:00—22:00；谷时段为0:00—8:00；其余为平时段。峰时段上浮幅度为58%，谷时段下浮幅度为63%。7—9月，11:00—12:00，17:00—18:00为尖峰时段，上浮幅度为80%。",
		"广东":    "6.峰谷分时电价按粤发改价格〔2021〕331号文执行，以代理购电用户电价（不含政府性基金及附加）为基础电价（平段电价），按规定的实施范围、峰谷时段、峰谷比价、尖峰电价等政策执行。其中，峰谷分时电价全省执行统一时段划分，高峰时段为10-12点、14-19点；低谷时段为0-8点；其余时段为平段；峰平谷比价为1.7:1:0.38。尖峰电价执行时间为7月、8月和9月三个整月，以及其他月份中广州日最高气温达到35℃及以上的高温天，执行时段为每天11-12时、15-17时共三个小时，尖峰电价在峰段电价基础上上浮25%。具体电价水平以分/千瓦时为单位四舍五入到小数点后两位。",
		"深圳":    "5.峰谷分时电价按深发改〔2021〕1005号文执行，以代理购电用户电价（含代理购电价格、输配电价等，不含政府性基金及附加）为基础电价（平段电价），按规定的实施范围、峰谷时段、峰谷比价、尖峰电价等政策执行。其中，峰谷分时电价全省执行统一时段划分，高峰时段为10-12点、14-19点；低谷时段为0-8点；其余时段为平段。尖峰电价执行时间为7月、8月和9月三个整月，以及其他月份中广州日最高气温达到35℃及以上的高温天，执行时段为每天11-12时、15-17时共三个小时，尖峰电价在峰段电价基础上上浮25%。具体电价水平以分/千瓦时为单位四舍五入到小数点后两位。",
		"贵州":    "2.省发展改革委关于完善峰谷分时电价机制有关事项的通知（黔发改价格〔2023〕481号）时段划分：高峰时段10:00(含)-13:00，17:00(含)-22:00；平时段8:00(含)-10:00，13:00(含)-17:00，22:00(含)-00:00；低谷时段00:00(含)-8:00。平段电价不浮动，高峰电价为平段电价上浮60%，低谷电价为平段电价下浮60%。",
		"海南":    "2.分时电度用电价格在电度用电价格基础上根据《海南省发展和改革委员会关于进一步完善峰谷分时电价机制有关问题的通知》（琼发改规〔2021〕18号）文件规定形成。时段划分：峰时段10：00-12：00，16：00-22：00；平时段7：00-10：00，12：00-16：00，22：00-23：00；谷时段23：00-次日7：00。输配电价按我省现行峰谷分时电价浮动比例浮动，用户承担的上网环节线损电价、系统运行费、政府性基金及附加、其他分摊（分享）费用折价不参与峰谷电价浮动。浮动比例：高峰电价为平段电价上浮70%，低谷电价为平段电价下浮60%。尖峰电价执行时间为每年的5月、6月、7月，尖峰时段为每日20：00-22：00，尖峰电价在峰段电价基础上上浮20%。尖峰电价实施范围与峰谷分时电价政策一致，包括电动汽车充换电用户。",
		"黑龙江":   "2.黑龙江省发展和改革委员会关于进一步完善峰谷分时电价政策措施有关事项的通知黑发改价格函〔2024〕291号电量上网电价执行峰谷分时电价政策。上网环节线损费用、电量输配电价、系统运行费用、政府性基金及附加不参与峰谷分时电价计算。高峰上浮50%，低谷下浮50%，暂停执行尖峰电价，今后根据电力供需状况适时启动尖峰电价。时段划分：高峰时段7:00-8:00、9:00-11:30、15:30-20:00；低谷时段12:00-14:00、23:30-5:30；其余为平时段。",
		"吉林":    "3.分时电度用电价格在电度用电价格基础上根据《吉林省发展改革委关于进一步优化分时电价政策的通知》（吉发改价格〔2025〕11号）等文件执行。时段划分：高峰时段8:00-10:00、16:00-21:00；平时段5:00-8:00、10:00-11:00、14:00-16:00、21:00-23:00；低谷时段00:00-5:00、11:00-14:00、23:00-24:00。高峰和低谷时段用电价格在平时段电价基础上分别上下浮动55%，尖峰电价不再按固定的月份与时段执行，调整为灵活启动机制。输配电价、系统运行费用折价、政府性基金及附加和上网环节线损电价均不参与峰谷浮动。",
		"江苏":    "3.根据《关于优化工商业分时电价结构促进新能源消纳降低企业用电成本支持经济社会发展的通知》（苏发改价格发〔2025〕426号）文件要求，将工商业分时电价执行范围扩大到除国家有专门规定的电气化铁路牵引用电外的执行工商业电价的电力用户。全体商业用户和100千伏安以下的工业用户，公用水厂、污水处理厂、分布式能源站、地下铁路、城铁、电车自行选择执行分时电价。夏、冬两季（每年6-8月、12月-次年2月）时段设置：高峰时段14:00-22:00，平时段6:00-11:00、13:00-14:00、22:00-24:00，低谷时段0:00-6:00、11:00-13:00。春、秋两季（每年3-5月、9-11月）时段设置：高峰时段15:00-22:00，平时段6:00-10:00、14:00-15:00、22:00-次日2:00，低谷时段2:00-6:00、10:00-14:00。选择执行分时电价的全体商业用户和100千伏安以下的工业用户，可选择执行季节性分时时段或全年执行春、秋两季分时时段。为有效衔接电力市场交易，将工商业用户分时电价计价基础，调整为以工商业用户购电价格为基础，并优化峰谷浮动比例，高峰时段上浮比例：两部制用户80%、100千伏安及以上单一制用户70%、100千伏安以下单一制用户60%；低谷时段下浮比例65%。工商业用户中除工业用户以外的电热锅炉（蓄冰制冷）用电，执行两段制电价。",
		"江西":    "2.分时电价按《江西省发展改革委关于进一步完善分时电价机制有关事项的通知》（赣发改价管〔2025〕463号）执行。峰谷时段划分：1月和12月：高峰（含尖峰）时段9:00-12:00、18:00-21:00，其中尖峰时段为18:00-20:00；低谷时段0:00-6:00；其余时段为平段。2月：高峰时段16:00-22:00；低谷时段0:00-6:00；其余时段为平段。7-9月：高峰（含尖峰）时段17:00-23:00，其中尖峰时段为7月、8月的20:30-22:30；低谷（含深谷）时段1:00-5:00、11:30-14:30，其中深谷时段为12:00-14:00；其余时段为平段。3-6月和10-11月：高峰时段16:00-22:00；低谷（含深谷）时段1:00-5:00、11:30-14:30，其中深谷时段为12:00-14:00；其余时段为平段。春节、“五一”国际劳动节、国庆节的12:00-14:00为深谷时段。全年高峰、平段、低谷浮动比例统一调整为1.6:1:0.4，深谷浮动比例较平段电价下浮70%，尖峰浮动比例较平段电价上浮80%。上网环节线损费用、输配电价、系统运行费用、政府性基金及附加不参与浮动。",
		"陕西":    "4.陕西省发展和改革委员会关于调整陕西电网分时电价政策有关事项的通知(陕发改价格〔2025〕1034号):分时电价在平段价格基础上扣除输配电价、基金附加、系统运行费和上网环节线损电价后：大工业（两部制）、一般工商业（单一制）峰谷浮动比例均为70%，尖峰电价在平段价格基础上上浮90%。时段划分：尖峰时段为夏季7、8月19:00-21:00，冬季12、1月18:00-20:00；高峰时段为16:00-23:00；低谷时段为0:00-6:00、11:00-14:00；其余时段为平时段。",
		"上海":    "2.分时电价时段划分按照《关于进一步完善我市分时电价机制有关事项的通知》（沪发改价管﹝2022﹞50号）文件规定执行：（1）一般工商业单一制未分时用户执行非分时电度价格；一般工商业单一制分时用户：高峰时段（6-22时），低谷时段（22时-次日6时）；（2）两部制分时用户：夏季7-9月：高峰时段（8-15时、18-21时），平时段（6-8时、15-18时、21-22时），低谷时段（22时-次日6时），其中7、8月尖峰时段（12-14时）；其他月份：高峰时段（8-11时、18-21时），平时段（6-8时、11-18时、21-22时），低谷时段（22时-次日6时），其中1、12月尖峰时段（19-21时）；两部制未分时用户按照非分时电度电价标准执行，容（需）量用电价格按照国家规定标准执行。3.分时电价浮动比例按照《关于进一步完善我市分时电价机制有关事项的通知》（沪发改价管﹝2022﹞50号）文件规定执行：（1）一般工商业及其他两部制、大工业两部制：夏季（7-9月）和冬季（1、12月）高峰上浮80%，低谷下浮60%，尖峰电价在高峰电价基础上上浮25%，其他月份高峰上浮60%，低谷下浮50%；（2）一般工商业及其他单一制：夏季（7-9月）和冬季（1、12月）高峰上浮20%，低谷下浮45%，其他月份高峰上浮17%，低谷下浮45%。4.按照《关于增加大工业深谷电价实施时间有关事项的通知》（沪发改价管﹝2024﹞28号）文件规定执行：我市大工业用电深谷实施时间由原来春节、劳动节、国庆节扩大到元旦、春节、清明节、劳动节、端午节、中秋节、国庆节，以及2月-6月、9-11月的休息日，深谷时段为当日0:00-6:00及22:00-24:00，深谷电价在平段电价基础上下浮80%。上述国家法定节假日及休息日具体时间以国家公布为准。5.对于已直接参与市场交易（不含已在电力交易平台注册但未曾参与电力市场交易）在无正当理由情况下改由电网企业代理购电的用户，拥有燃煤发电自备电厂、由电网企业代理购电的用户，暂不能直接参与市场交易由电网企业代理购电的高耗能用户，代理购电价格按上表中的1.5倍执行，其他标准及规则同常规用户。",
		"浙江":    "2.根据《省发展改革委关于调整工商业峰谷分时电价政策有关事项的通知》（浙发改价格〔2024〕21号），春秋季（2-6月、9-11月）峰谷分时电价时段划分：高峰时段：8:00-11:00、13:00-17:00；平段时段：17:00-24:00；低谷时段：0:00-8:00、11:00-13:00；夏冬季（1月、7月、8月、12月）峰谷分时电价时段划分：尖峰时段：9:00-11:00、15:00-17:00；高峰时段：8:00-9:00、17:00-23:00；平段时段：13:00-15:00、23:00-24:00；低谷时段：0:00-8:00、11:00-13:00；深谷时段：春节、劳动节、国庆节的10:00-14:00，三个节假日具体时间以国家公布为准。",
		"青海":    "2.青海省发展和改革委员会关于优化完善我省峰谷分时电价政策的通知（青发改价格〔2024〕179号）分时电价在代理购电价格基础上，高峰上浮63%，低谷下浮65%，上（下）浮后再加计上网环节线损折价、输配电价、系统运行费用折价、政府性基金及附加。我省所有执行峰谷分时电价政策的工商业用户高峰时段7:00-9:00、17:00-23:00；低谷时段9:00-17:00；其余时段为平时段。",
		"内蒙古_1": "3.分时电度用电价格在电度用电价格基础上根据内发改价费字〔2023〕1630号文件规定形成。大风季（1-5月、9-12月）时段划分：峰时段6小时:06:00-08:00、18:00-22:00，平时段9小时:04:00-06:00、08:00-11:00、16:00-18:00、22:00-24:00，谷时段9小时:0:00-4:00、11:00-16:00。大风季峰平谷交易价格比为1.68:1:0.48，平段价格为平时段平均交易价格，峰段在平段价格的基础上上浮68%，谷段在平段价格的基础上下浮52%。小风季（6-8月）时段划分：峰时段6小时:06:00-08:00、18:00-22:00，平时段13小时:00:00-06:00、08:00-11:00、16:00-18:00、22:00-24:00，谷时段5小时:11:00-16:00。小风季峰平谷交易价格比为1.54:1:0.44，平段价格为平时段平均交易价格，峰段在平段价格的基础上上浮54%，谷段在平段价格的基础上下浮56%。每年6-8月实施尖峰电价、深谷电价，尖峰时段为每日19:00-21:00，尖峰电价在峰段价格基础上上浮20%；深谷时段为每日13:00-15:00，深谷电价在谷段价格基础上下浮20%。",
		"内蒙古_2": "2. 按照《内蒙古自治区发展和改革委员会关于完善蒙东电网工商业分时电价政策的通知（试行）》内发改价费字〔2023〕1631 号文件，峰时段 06:00 - 09:00、17:00 - 22:00，平时段 05:00 - 06:00、09:00 - 11:00、14:00 - 17:00、22:00 - 24:00，谷时段 11:00 - 14:00、00:00 - 05:00。每年 6 - 8 月实施尖峰、深谷电价，尖峰时段为每日 18:00 - 20:00，深谷时段为每日 12:00 - 14:00。高峰电价在平段价格基础上上浮 68%，低谷电价在平段价格基础上下浮 52%，尖峰电价在峰段价格基础上上浮 20%，深谷电价在谷段价格基础上下浮 20%。",
		"山东":    "2. 根据鲁发改价格〔2023〕914 号文件规定，容量补偿电价、上网环节线损费用折价、抽水蓄能容量电费折合度电水平、煤电容量电费折合度电水平以及电网企业代理购电用户 (不含国家有专门规定的电气化铁路牵引用电) 当月平均购电价格执行分时电价政策，历史偏差电费折价、各类分摊损益折合度电水平等其它电价部分不执行。其中 2025 年时段划分：1 月至 2 月、12 月，低谷时段为 02:00 至 06:00、10:00 至 15:00，其中，深谷时段为 11:00 至 14:00；高峰时段为 07:00 至 09:00、16:00 至 21:00，其中，尖峰时段为 16:00 至 19:00。3 月至 5 月，低谷时段为 10:00 至 15:00，其中，深谷时段为 11:00 至 14:00；高峰时段为 17:00 至 22:00，其中，尖峰时段为 17:00 至 20:00。6 月，低谷时段为 07:00 至 12:00；高峰时段为 16:00 至 23:00，其中，尖峰时段为 17:00 至 22:00。7 月至 8 月，低谷时段为 01:00 至 06:00；高峰时段为 16:00 至 23:00，其中，尖峰时段为 17:00 至 22:00。9 月至 11 月，低谷时段为 10:00 至 15:00，其中，深谷时段为 11:00 至 14:00；高峰时段为 16:00 至 21:00，其中，尖峰时段为 17:00 至 19:00。峰谷时段外其他时段为平时段。浮动比例：峰段上浮 70%、谷段下浮 70%、尖峰上浮 100%、深谷下浮 90%。",
		"新疆":    "2. 分时电度用电价格根据《自治区发展改革委关于进一步完善分时电价有关事宜的通知》(新发改规〔2023〕11 号) 文件规定形成。分时电价时段划分：高峰时段 8 小时 (8:00—11:00,19:00—24:00)；平段 8 小时 (11:00—13:00,17:00—19:00,0:00—4:00)；低谷时段 8 小时 (4:00—8:00,13:00—17:00)。7 月份的 21:00—23:00，1、11、12 月份的 19:00—21:00 由高峰时段调整为尖峰时段，执行尖峰电价。5、6、7、8 月份 14:00 - 16:00 由低谷时段调整为深谷时段，执行深谷电价。浮动比例：工商业用电高峰、低谷电价分别在平段电价基础上上下浮动 75%，尖峰时段电价在平段电价上上浮 100%，深谷时段电价在平段电价基础上下浮 90%。代理购电历史偏差电费折价、上网环节线损电价、输配电价、系统运行费折价、政府性基金及附加不参与分时电价浮动。",
		"云南":    "4.云南省发展和改革委员会关于优化调整分时电价政策有关事项的通知云发改价格〔2024〕924号峰谷分时电价时段划分：高峰时段 7:00 - 9:00、18:00 - 24:00，平时段 0:00 - 2:00、6:00 - 7:00、9:00 - 12:00、16:00 - 18:00，低谷时段 2:00 - 6:00、12:00 - 16:00。浮动比例：分时电价以各电力用户市场化交易价格（代理购电价格）为基准进行浮动，高峰时段电价以平时段电价为基础上浮 50%，低谷时段电价以平时段电价为基础下浮 50%。输配电价（含线损）、系统运行费用和政府性基金及附加不参与浮动。",
		"西藏":    "备注：1.西藏自治区人民政府办公厅关于进一步优化调整全区上网电价和销售电价引导降低社会用电成本的通知藏政办发〔2023〕28号丰水期（5 - 10 月）用电时段：平段：0:00 - 3:00，6:00 - 11:00，17:00 - 24:00；低谷：11:00 - 17:00；深谷：3:00 - 6:00。2. 枯水期（1 - 4 月、11 - 12 月）用电时段：平段：0:00 - 9:00，13:00 - 19:00；高峰时段：9:00 - 13:00，22:00 - 24:00；尖峰时段：19:00 - 22:00。3. 根据国家代理购电相关政策规定，结合我区实际，高耗能企业用户枯水期用电根据购电侧价格变动情况，执行浮动价格；4. 非工业用户不执行一般工商业分季节电价。",
		"天津":    "3. 分时电度电价在非分时电度电价基础上形成。时段划分、浮动比例以及执行范围等事项按照《市发展改革委关于峰谷分时电价政策有关事项的通知》（津发改价综〔2021〕395 号）文件规定执行。时段划分：高峰时段 9:00 - 12:00，16:00 - 21:00；低谷时段 23:00 - 7:00；平时段 7:00 - 9:00，12:00 - 16:00，21:00 - 23:00。高峰电价在平时段电价的基础上上浮 50%，低谷电价在平时段电价的基础上下浮 54%。9 月份无尖峰时段，对尚未调整表计时段，显示为尖峰的电量应为高峰电量，按高峰电价执行。根据《市发展改革委关于贯彻落实第三监管周期输配电价改革有关事项的通知》（津发改价格〔2023〕142 号）、《市发展改革委关于峰谷分时电价政策有关事项的通知》（津发改价综〔2021〕395 号）、《市工业和信息化局关于做好天津市 2025 年电力市场化交易工作的通知》（津工信电力〔2024〕21 号）文件规定，平时段电价中上网环节线损费用折价、系统运行费用折合度电水平、政府性基金及附加、两部制电价的基本电费、功率因数调整电费不参与浮动。",
		"四川":    "3. 四川省发展和改革委员会关于进一步调整我省分时电价机制的通知（川发改价格〔2025〕185号）符合我省分时电价政策执行范围的由电网企业代理购电的工商业用户和直接从电力市场购电的工商业用户，执行峰谷分时电价政策。分时电价时段划分：春秋季（3 - 6、10、11 月）高峰时段 10:00 - 12:00，17:00 - 22:00；平段 8:00 - 10:00，12:00 - 17:00；低谷时段 22:00 - 次日 8:00。夏季（7 - 9 月）高峰时段 11:00 - 18:00，20:00 - 23:00；平段 7:00 - 11:00，18:00 - 20:00、23:00 - 次日 1:00；低谷时段 1:00 - 7:00。冬季（1、2、12 月）高峰时段 10:00 - 12:00，16:00 - 22:00；平段 8:00 - 10:00、12:00 - 16:00、22:00 - 24:00；低谷时段 0:00 - 8:00。高峰时段电价在平段电价上上浮 60%，低谷时段电价在平段电价上下浮 60%。7 - 8 月全月、其他月份连续三日最高气温≥35℃时，对执行分时电价的大工业用户实行尖峰电价政策（其中：攀枝花市、凉山州、甘孜州、阿坝州大工业用户暂不执行），尖峰时段 13:00 - 14:00、21:00 - 23:00，尖峰时段用电价格在高峰时段电价基础上上浮 20%。",
		"湖南":    "2. 分时电度用电价格根据湖南省发改委《关于优化我省分时电价政策有关事项的通知》（湘发改价调〔2025〕385 号）文件规定形成。时段划分：尖峰时段 20:00 - 24:00（7、8 月），18:00 - 22:00（1、12 月）；高峰时段 16:00 - 24:00（7、8 月为 16:00 - 20:00；1、12 月为 16:00 - 18:00，22:00 - 24:00）；平时段 6:00 - 12:00，14:00 - 16:00；低谷时段 0:00 - 6:00，12:00 - 14:00。浮动比例：高峰电价为平段电价上浮 60%，低谷电价为平段电价下浮 60%，尖峰电价在高峰电价基础上上浮 20%。",
		"湖北":    "2. 分时电价根据《省发改委关于完善工商业分时电价机制有关事项的通知》（鄂发改价管〔2024〕77 号）文件规定形成。时段划分：尖峰时段：7 月、8 月 20:00 - 22:00，其他月份 18:00 - 20:00（共 2 小时）；高峰时段：7 月、8 月 16:00 - 20:00、22:00 - 24:00，其他月份 16:00 - 18:00、20:00 - 24:00（共 6 小时）；平时段：6:00 - 12:00、14:00 - 16:00（共 8 小时）；低谷时段：0:00 - 6:00、12:00 - 14:00（共 8 小时）。基础电价：在电度电价基础上扣除政府性基金及附加、电度输配电价和系统运行费折价，作为峰谷分时电价计算的基础电价。",
		"甘肃":    "3. 分时电度用电价格在电度用电价格基础上根据《关于优化调整工商业等用户峰谷分时电价政策有关事项的通知》（甘发改价格〔2024〕424 号）规定形成。时段划分：高峰时段 6:00 - 8:00、18:00 - 23:00；平时段 23:00 - 6:00（次日）、8:00 - 10:00、16:00 - 18:00；低谷时段 10:00 - 16:00。按照《甘肃省 2025 年省内电力中长期年度交易实施方案》规定，上网电价峰谷价差比例由市场自主形成，上网环节线损电价随上网电价参与峰谷浮动，电度输配电价、系统运行费、政府性基金及附加不参与峰谷浮动。",
		"河南":    "关于调整工商业分时电价有关事项的通知文号：豫发改价管〔2024〕283号《河南省发展和改革委员会关于进一步完善分时电价机制有关事项的通知》（豫发改价管〔2022〕867 号）已实施满 1 年。为更好保障电力系统安全稳定经济运行，在改善电力供需状况、促进新能源消纳的基础上，进一步引导用户调整用电负荷，鼓励新能源城市公交车辆等电动汽车更多在低谷时段充电，根据用电负荷变化、新能源出力特点等因素，结合现行分时电价政策执行评估情况，经省政府同意，现就调整工商业分时电价有关事项通知如下：一、优化峰谷时段设置 1 月、2 月、12 月，高峰（含尖峰）时段 16∶00 至 24∶00，其中尖峰时段为 1 月和 12 月的 17∶00 至 19∶00；低谷时段 0∶00 至 7∶00，其他时段为平段。 3—5 月和 9—11 月，高峰时段 16∶00 至 24∶00，低谷时段 0∶00 至 6∶00、11∶00 至 14∶00，其他时段为平段。 6—8 月，高峰（含尖峰）时段 16∶00 至 24∶00，其中尖峰时段为 7 月和 8 月的 20∶00 至 23∶00；低谷时段 0∶00 至 7∶00，其他时段为平段。具体时段详见附件。",
		"河北_1":  "河北省发展和改革委员会关于优化调整河北南网工商业及其他用户分时电价政策的通知(冀发改能价[2025]1115号)：时段划分春季（3、4、5 月）：深谷时段为 12—15 时，低谷时段为 3—7 时、11—12 时，平段时段为 0—3 时、7—11 时、15—16 时，高峰时段为 16—24 时；夏季（6、7、8 月）：低谷时段为 1—7 时、12—14 时，平段时段为 0—1 时、7—12 时、14—16 时，高峰时段为 16—19 时、22—24 时，尖峰时段为 19—22 时；秋季（9、10、11 月）：深谷时段为 12—14 时，低谷时段为 2—6 时、11—12 时、14—15 时，平段时段为 0—2 时、6—11 时、15—16 时，高峰时段为 16—24 时；冬季（12、1、2 月）：低谷时段为 2—6 时、11—15 时，平段时段为 0—2 时、6—7 时、9—11 时、15—17 时、23—24 时，高峰时段为 7—9 时、19—23 时，尖峰时段为 17—19 时。",
		"河北_2":  "河北省发展和改革委员会关于进一步完善冀北电网工商业及其他用户分时电价政策的通知（冀发改能价 [2023] 1711 号）二、时段划分（一）夏季（每年 6、7、8 月）低谷：0-7 时、23-24 时；平段：7-10 时、12-16 时、22-23 时；高峰：10-12 时、16-17 时、20-22 时；尖峰：17-20 时。（二）冬季（每年 11、12 月及次年 1 月）低谷：1-7 时、12-14 时；平段：0-1 时、7-8 时、10-12 时、14-16 时、22-24 时；高峰：8-10 时、16-17 时、19-22 时；尖峰：17-19 时。（三）其他季节（每年 2、3、4、5 月及 9、10 月）低谷：1-7 时、12-14 时；平段：0-1 时、7-8 时、10-12 时、14-16 时、22-24 时；高峰：8-10 时、16-22 时。",
		"广西":    "广西壮族自治区发展和改革委员会关于优化峰谷分时电价机制的通知（桂发改价格〔2023〕609 号）一、优化峰平谷时段划分 峰谷时段按每日 24 小时分为高峰、平段、低谷三段，时长调整为各 8 个小时，时段划分调整为：峰时段：11:00—13:00、17:00—23:00；平时段：7:00—11:00、13:00—17:00；谷时段：23:00—24:00、00:00—7:00。",
		"山西":    "关于完善分时电价机制有关事项的通知 发文字号：晋发改商品发〔2021〕479 号 发布日期：2022-01-05 各市发展改革委，国网山西省电力公司、山西地方电力有限公司、山西电力交易中心：根据《国家发展改革委关于进一步完善分时电价机制的通知》（发改价格〔2021〕1093 号）、《国家发展改革委关于进一步深化燃煤发电上网电价市场化改革的通知》（发改价格〔2021〕1439 号）要求，为充分发挥分时电价作用，更好引导用户削峰填谷，服务新能源为主体的新型电力系统建设，保障电力安全稳定供应，现就进一步完善我省分时电价政策有关事项通知如下：一、分时电价执行范围 除国家有专门规定的电气化铁路牵引用电外的执行工商业电价的电力用户（包括一般工商业和大工业用电）。二、优化峰谷时段划分 根据我省电力供需状况、系统用电负荷特性、可再生能源发展和季节性负荷变化等情况，将现行峰谷时段优化调整为：高峰时段 08：00－11：00、17：00－23：00；低谷时段 00：00－07：00、11：00－13：00；平时段 07：00－08：00、13：00－17：00、23：00－24：00。三、适当扩大峰谷价差 根据我省电力系统峰谷差率、新能源消纳和系统调节能力等情况，将我省峰谷价差调整为 3.6：1。即高峰时段电价在平时段电价基础上上浮 60％，低谷时段电价在平时段电价基础上下浮 55％，尖峰时段电价在高峰时段电价基础上上浮 20％。保障居民农业用电价格稳定新增损益、代理购电新增线损损益、辅助服务费用、政府性基金及附加、容（需）量电价、代理购电偏差电费不参与浮动。四、实施季节性尖峰电价 每年冬、夏两季对大工业电力用户实施尖峰电价政策，其中 1 月、7 月、8 月、12 月 18：00－20：00 为尖峰时段，尖峰时段电价在峰时段电价基础上上浮 20％。",
		"宁夏":    "自治区发展改革委关于做好 2025 年电力中长期交易有关事项的通知（宁发改运行〔2024〕952 号）：四、时段划分 1. 为高效衔接现货市场，中长期交易按日划分 24 小时时段，各市场主体根据自身发电特性和用电需求合理参与分时段交易。2. 为引导市场主体形成合理分时段交易价格，根据《自治区发展改革委关于优化峰谷分时电价机制的通知》（宁发改价格（管理）〔2023〕7 号），结合宁夏电网电力时段性供需情况，将 24 小时时段归为峰（含尖峰）、平、谷（含深谷）三类，具体为：峰段：7:00－9:00，17:00-23:00；谷段：9:00-17:00；平段：0:00-7:00，23:00－0:00。3. 根据区内电力供需情况，适时调整峰、平、谷时段划分。",
		"重庆":    "《关于进一步完善我市分时电价制有关事项的通知》（渝发改规范〔2021〕14 号）二、时段划分 高峰时段、平段和谷段分别各 8 小时，具体如下：（一）高峰时段：11∶00—17∶00、20∶00—22∶00。其中，夏季 7 月、8 月以及冬季 12 月、1 月的 12∶00—14∶00 为尖峰时段。（二）平段：08∶00—11∶00、17∶00—20∶00、22∶00—24∶00。（三）低谷时段：00∶00—08∶00。",
		"辽宁":    "省发展改革委关于进一步完善分时电价机制有关事项的通知（辽发改价格〔2023〕441 号）二、完善有关机制 （一）峰谷时段划分。根据我省系统用电负荷特性、新能源消纳等情况，每日用电分为高峰、平时、低谷三个时段，各 8 小时。具体为：高峰时段：07:30-10:30、16:00-21:00；低谷时段：11:30-12:30，22:00-05:00；其余为平时时段。其中，每年夏季（7 月、8 月）、冬季（1 月、12 月）每日 17:00-19:00 为尖峰时段，执行尖峰电价。",
	}

	outputs := []DlgdPeriod{
		{
			Area:  "安徽",
			DocNo: "皖发改价格[2025]302号",
			Hours: []DlgdHour{
				{Name: "峰段", Value: "16:00-24:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-02:00,09:00-11:00,13:00-16:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "02:00-09:00,11:00-13:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "15:00-23:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "06:00-12:00,14:00-15:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "12:00-14:00,23:00-06:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "06:00-08:00,16:00-22:00", Temp: "", Months: []int64{2, 3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "08:00-11:00,14:00-16:00,22:00-23:00", Temp: "", Months: []int64{2, 3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "11:00-14:00,23:00-06:00", Temp: "", Months: []int64{2, 3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "11:00-15:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{"3天及以上节假日"}, Categories: []string{"容量315千伏安及以上", "两部制"}},
			},
		},
		{
			Area:  "北京",
			DocNo: "京发改规[2023]11号",
			Hours: []DlgdHour{
				{Name: "尖段", Value: "11:00-13:00,16:00-17:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "18:00-21:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "10:00-13:00,17:00-22:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "07:00-10:00,13:00-17:00,22:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "23:00-07:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "重庆",
			DocNo: "渝发改规范[2021]14号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "11:00-17:00,20:00-22:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "12:00-14:00", Temp: "", Months: []int64{7, 8, 12, 1}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "08:00-11:00,17:00-20:00,22:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-08:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "福建",
			DocNo: "闽发改规[2023]8号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "10:00-12:00,15:00-20:00,21:00-22:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-08:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "11:00-12:00,17:00-18:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "甘肃",
			DocNo: "甘发改价格[2024]424号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "06:00-08:00,18:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "23:00-06:00,08:00-10:00,16:00-18:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "10:00-16:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "广东",
			DocNo: "粤发改价格[2021]331号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "10:00-12:00,14:00-19:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-08:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "11:00-12:00,15:00-17:00", Temp: "其他月份中广州日最高气温达到35℃", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "广西",
			DocNo: "桂发改价格[2023]609号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "11:00-13:00,17:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "07:00-11:00,13:00-17:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "23:00-24:00,00:00-07:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "贵州",
			DocNo: "黔发改价格[2023]481号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "10:00-13:00,17:00-22:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "08:00-10:00,13:00-17:00,22:00-00:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-08:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "海南",
			DocNo: "琼发改规[2021]18号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "10:00-12:00,16:00-22:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "07:00-10:00,12:00-16:00,22:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "23:00-07:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "20:00-22:00", Temp: "", Months: []int64{5, 6, 7}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "河北",
			DocNo: "冀发改能价[2025]1115号",

			Hours: []DlgdHour{
				{Name: "深谷", Value: "12:00-15:00", Temp: "", Months: []int64{3, 4, 5}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "03:00-07:00,11:00-12:00", Temp: "", Months: []int64{3, 4, 5}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-03:00,07:00-11:00,15:00-16:00", Temp: "", Months: []int64{3, 4, 5}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-24:00", Temp: "", Months: []int64{3, 4, 5}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "01:00-07:00,12:00-14:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-01:00,07:00-12:00,14:00-16:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-19:00,22:00-24:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "19:00-22:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "12:00-14:00", Temp: "", Months: []int64{9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "02:00-06:00,11:00-12:00,14:00-15:00", Temp: "", Months: []int64{9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-02:00,06:00-11:00,15:00-16:00", Temp: "", Months: []int64{9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-24:00", Temp: "", Months: []int64{9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "02:00-06:00,11:00-15:00", Temp: "", Months: []int64{12, 1, 2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-02:00,06:00-07:00,09:00-11:00,15:00-17:00,23:00-24:00", Temp: "", Months: []int64{12, 1, 2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "07:00-09:00,19:00-23:00", Temp: "", Months: []int64{12, 1, 2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "17:00-19:00", Temp: "", Months: []int64{12, 1, 2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "河北",
			DocNo: "冀发改能价[2023]1711号",

			Hours: []DlgdHour{
				{Name: "谷段", Value: "00:00-07:00,23:00-24:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "07:00-10:00,12:00-16:00,22:00-23:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "10:00-12:00,16:00-17:00,20:00-22:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "17:00-20:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "01:00-07:00,12:00-14:00", Temp: "", Months: []int64{11, 12, 1}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-01:00,07:00-08:00,10:00-12:00,14:00-16:00,22:00-24:00", Temp: "", Months: []int64{11, 12, 1}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "08:00-10:00,16:00-17:00,19:00-22:00", Temp: "", Months: []int64{11, 12, 1}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "17:00-19:00", Temp: "", Months: []int64{11, 12, 1}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "01:00-07:00,12:00-14:00", Temp: "", Months: []int64{2, 3, 4, 5, 9, 10}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-01:00,07:00-08:00,10:00-12:00,14:00-16:00,22:00-24:00", Temp: "", Months: []int64{2, 3, 4, 5, 9, 10}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "08:00-10:00,16:00-22:00", Temp: "", Months: []int64{2, 3, 4, 5, 9, 10}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "黑龙江",
			DocNo: "黑发改价格函[2024]291号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "07:00-08:00,09:00-11:30,15:30-20:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "12:00-14:00,23:30-05:30", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "河南",
			DocNo: "豫发改价管[2024]283号",

			Hours: []DlgdHour{
				{Name: "尖段", Value: "17:00-19:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-24:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-07:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-24:00", Temp: "", Months: []int64{3, 4, 5, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-06:00,11:00-14:00", Temp: "", Months: []int64{3, 4, 5, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: []int64{3, 4, 5, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "20:00-23:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-24:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-07:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "湖北",
			DocNo: "鄂发改价管[2024]77号",

			Hours: []DlgdHour{
				{Name: "尖段", Value: "20:00-22:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "18:00-20:00", Temp: "", Months: []int64{1, 2, 3, 4, 5, 6, 9, 10, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-20:00,22:00-24:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-18:00,20:00-24:00", Temp: "", Months: []int64{1, 2, 3, 4, 5, 6, 9, 10, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "06:00-12:00,14:00-16:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-06:00,12:00-14:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "湖南",
			DocNo: "湘发改价调[2025]385号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "16:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-20:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-18:00,22:00-24:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "20:00-24:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "18:00-22:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "06:00-12:00,14:00-16:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-06:00,12:00-14:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "吉林",
			DocNo: "吉发改价格[2025]11号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "08:00-10:00,16:00-21:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "05:00-08:00,10:00-11:00,14:00-16:00,21:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-05:00,11:00-14:00,23:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "江苏",
			DocNo: "苏发改价格发[2025]426号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "14:00-22:00", Temp: "", Months: []int64{6, 7, 8, 12, 1, 2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "06:00-11:00,13:00-14:00,22:00-24:00", Temp: "", Months: []int64{6, 7, 8, 12, 1, 2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-06:00,11:00-13:00", Temp: "", Months: []int64{6, 7, 8, 12, 1, 2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "15:00-22:00", Temp: "", Months: []int64{3, 4, 5, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "06:00-10:00,14:00-15:00,22:00-02:00", Temp: "", Months: []int64{3, 4, 5, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "02:00-06:00,10:00-14:00", Temp: "", Months: []int64{3, 4, 5, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "江西",
			DocNo: "赣发改价管[2025]463号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "09:00-12:00,18:00-21:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "18:00-20:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-06:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-22:00", Temp: "", Months: []int64{2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-06:00", Temp: "", Months: []int64{2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: []int64{2}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "20:30-22:30", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "17:00-23:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "01:00-05:00,11:30-14:30", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "12:00-14:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-22:00", Temp: "", Months: []int64{3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "01:00-05:00,11:30-14:30", Temp: "", Months: []int64{3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "12:00-14:00", Temp: "", Months: []int64{3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: []int64{3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "12:00-14:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{"春节", "劳动节", "国庆节"}, Categories: nil},
			},
		},
		{
			Area:  "辽宁",
			DocNo: "辽发改价格[2023]441号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "07:30-10:30,16:00-21:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "11:30-12:30,22:00-05:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "17:00-19:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "内蒙古",
			DocNo: "内发改价费字[2023]1630号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "06:00-08:00,18:00-22:00", Temp: "", Months: []int64{1, 2, 3, 4, 5, 9, 10, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "04:00-06:00,08:00-11:00,16:00-18:00,22:00-24:00", Temp: "", Months: []int64{1, 2, 3, 4, 5, 9, 10, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-04:00,11:00-16:00", Temp: "", Months: []int64{1, 2, 3, 4, 5, 9, 10, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "06:00-08:00,18:00-22:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-06:00,08:00-11:00,16:00-18:00,22:00-24:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "11:00-16:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "19:00-21:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "13:00-15:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "内蒙古",
			DocNo: "内发改价费字[2023]1631号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "06:00-09:00,17:00-22:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "05:00-06:00,09:00-11:00,14:00-17:00,22:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "11:00-14:00,00:00-05:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "18:00-20:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "12:00-14:00", Temp: "", Months: []int64{6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "宁夏",
			DocNo: "宁发改价格(管理)[2023]7号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "07:00-09:00,17:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "09:00-17:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-07:00,23:00-00:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "青海",
			DocNo: "青发改价格[2024]179号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "07:00-09:00,17:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "09:00-17:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "山东",
			DocNo: "鲁发改价格[2023]914号",

			Hours: []DlgdHour{
				{Name: "谷段", Value: "02:00-06:00,10:00-15:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "11:00-14:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "07:00-09:00,16:00-21:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "16:00-19:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "10:00-15:00", Temp: "", Months: []int64{3, 4, 5}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "11:00-14:00", Temp: "", Months: []int64{3, 4, 5}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "17:00-22:00", Temp: "", Months: []int64{3, 4, 5}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "17:00-20:00", Temp: "", Months: []int64{3, 4, 5}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "07:00-12:00", Temp: "", Months: []int64{6}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-23:00", Temp: "", Months: []int64{6}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "17:00-22:00", Temp: "", Months: []int64{6}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "01:00-06:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-23:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "17:00-22:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "10:00-15:00", Temp: "", Months: []int64{9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "11:00-14:00", Temp: "", Months: []int64{9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-21:00", Temp: "", Months: []int64{9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "17:00-19:00", Temp: "", Months: []int64{9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "山西",
			DocNo: "晋发改商品发[2021]479号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "08:00-11:00,17:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-07:00,11:00-13:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "07:00-08:00,13:00-17:00,23:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "18:00-20:00", Temp: "", Months: []int64{1, 7, 8, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "陕西",
			DocNo: "陕发改价格[2025]1034号",

			Hours: []DlgdHour{
				{Name: "尖段", Value: "19:00-21:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "18:00-20:00", Temp: "", Months: []int64{12, 1}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "16:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-06:00,11:00-14:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "上海",
			DocNo: "沪发改价管[2022]50号,沪发改价管[2024]28号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "06:00-22:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"单一制"}},
				{Name: "谷段", Value: "22:00-06:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"单一制"}},
				{Name: "峰段", Value: "08:00-15:00,18:00-21:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"两部制"}},
				{Name: "平段", Value: "06:00-08:00,15:00-18:00,21:00-22:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"两部制"}},
				{Name: "谷段", Value: "22:00-06:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"两部制"}},
				{Name: "尖段", Value: "12:00-14:00", Temp: "", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"两部制"}},
				{Name: "峰段", Value: "08:00-11:00,18:00-21:00", Temp: "", Months: []int64{1, 2, 3, 4, 5, 6, 10, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"两部制"}},
				{Name: "平段", Value: "06:00-08:00,11:00-18:00,21:00-22:00", Temp: "", Months: []int64{1, 2, 3, 4, 5, 6, 10, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"两部制"}},
				{Name: "谷段", Value: "22:00-06:00", Temp: "", Months: []int64{1, 2, 3, 4, 5, 6, 10, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"两部制"}},
				{Name: "尖段", Value: "19:00-21:00", Temp: "", Months: []int64{1, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: []string{"两部制"}},
				{Name: "深谷", Value: "00:00-06:00及22:00-24:00", Temp: "", Months: nil, WeekendMonths: []int64{2, 3, 4, 5, 6, 9, 10, 11}, Holidays: []string{"春节", "劳动节", "国庆节", "元旦", "清明节", "端午节", "中秋节"}, Categories: []string{"大工业用电"}},
			},
		},
		{
			Area:  "深圳",
			DocNo: "深发改[2021]1005号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "10:00-12:00,14:00-19:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-08:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "11:00-12:00,15:00-17:00", Temp: "其他月份中广州日最高气温达到35℃", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "四川",
			DocNo: "川发改价格[2025]185号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "10:00-12:00,17:00-22:00", Temp: "", Months: []int64{3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "08:00-10:00,12:00-17:00", Temp: "", Months: []int64{3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "22:00-08:00", Temp: "", Months: []int64{3, 4, 5, 6, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "11:00-18:00,20:00-23:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "07:00-11:00,18:00-20:00,23:00-01:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "01:00-07:00", Temp: "", Months: []int64{7, 8, 9}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "10:00-12:00,16:00-22:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "08:00-10:00,12:00-16:00,22:00-24:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-08:00", Temp: "", Months: []int64{1, 2, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "13:00-14:00,21:00-23:00", Temp: "其他月份连续三日最高气温≥35℃", Months: []int64{7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "天津",
			DocNo: "津发改价综[2021]395号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "09:00-12:00,16:00-21:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "23:00-07:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "07:00-09:00,12:00-16:00,21:00-23:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "西藏",
			DocNo: "藏政办发[2023]28号",

			Hours: []DlgdHour{
				{Name: "平段", Value: "00:00-03:00,06:00-11:00,17:00-24:00", Temp: "", Months: []int64{5, 6, 7, 8, 9, 10}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "11:00-17:00", Temp: "", Months: []int64{5, 6, 7, 8, 9, 10}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "03:00-06:00", Temp: "", Months: []int64{5, 6, 7, 8, 9, 10}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-09:00,13:00-19:00", Temp: "", Months: []int64{1, 2, 3, 4, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "09:00-13:00,22:00-24:00", Temp: "", Months: []int64{1, 2, 3, 4, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "19:00-22:00", Temp: "", Months: []int64{1, 2, 3, 4, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "新疆",
			DocNo: "新发改规[2023]11号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "08:00-11:00,19:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "11:00-13:00,17:00-19:00,00:00-04:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "04:00-08:00,13:00-17:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "21:00-23:00", Temp: "", Months: []int64{7}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "19:00-21:00", Temp: "", Months: []int64{1, 11, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "深谷", Value: "14:00-16:00", Temp: "", Months: []int64{5, 6, 7, 8}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "云南",
			DocNo: "云发改价格[2024]924号",

			Hours: []DlgdHour{
				{Name: "峰段", Value: "07:00-09:00,18:00-24:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "00:00-02:00,06:00-07:00,09:00-12:00,16:00-18:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "02:00-06:00,12:00-16:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
		{
			Area:  "浙江",
			DocNo: "浙发改价格[2024]21号",

			Hours: []DlgdHour{
				{Name: "深谷", Value: "10:00-14:00", Temp: "", Months: nil, WeekendMonths: nil, Holidays: []string{"春节", "劳动节", "国庆节"}, Categories: nil},
				{Name: "峰段", Value: "08:00-11:00,13:00-17:00", Temp: "", Months: []int64{2, 3, 4, 5, 6, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "17:00-24:00", Temp: "", Months: []int64{2, 3, 4, 5, 6, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-08:00,11:00-13:00", Temp: "", Months: []int64{2, 3, 4, 5, 6, 9, 10, 11}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "尖段", Value: "09:00-11:00,15:00-17:00", Temp: "", Months: []int64{1, 7, 8, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "峰段", Value: "08:00-09:00,17:00-23:00", Temp: "", Months: []int64{1, 7, 8, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "平段", Value: "13:00-15:00,23:00-24:00", Temp: "", Months: []int64{1, 7, 8, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
				{Name: "谷段", Value: "00:00-08:00,11:00-13:00", Temp: "", Months: []int64{1, 7, 8, 12}, WeekendMonths: nil, Holidays: []string{}, Categories: nil},
			},
		},
	}

	for area, input := range inputs {
		period := NewPeriod(input)
		period.Area = strings.Split(area, "_")[0]
		inStr, _ := json.Marshal(period)
		match := false
		for _, output := range outputs {
			outStr, _ := json.Marshal(output)
			if string(inStr) == string(outStr) {
				match = true
				break
			}
		}

		if !match {
			t.Errorf("area %s, input %s, not match", period.Area, input)
		}
	}

}
