package cronx

import "math"

// PowerFactors 定义标准的功率因数考核值
var PowerFactors = [3]float64{0.8, 0.85, 0.9}

// AdjustPowerFactorFee 计算功率因数调整电费系数
// 根据(83)水电财字第215号文《功率因数调整电费办法》计算
// 参考: http://www.js.sgcc.com.cn/html/files/2021-09/29/20210929145633338232143.pdf
//
// 参数:
//   - pf: 功率因数标准值 (0.8/0.85/0.9)
//   - ep: 总有功电量 (单位: 千瓦时)
//   - eq: 总无功电量 (单位: 千乏时)
//
// 返回值:
//   float64: 调整系数
//     - 正值表示减收电费比例 (奖励)
//     - 负值表示增收电费比例 (惩罚)
//     - 0 表示不调整
func AdjustPowerFactorFee(pf, ep, eq float64) float64 {
	// 计算实际功率因数 (保留2位小数)
	actualPF := math.Abs(ep) / math.Sqrt(ep*ep+eq*eq)
	actualPF = math.Round(actualPF*100) / 100

	// 根据不同的标准值分别计算
	switch pf {
	case 0.9: // 适用于160kVA以上的高压供电工业用户
		switch {
		case actualPF >= 0.95: // 0.95及以上: 减收0.75%
			return -0.0075
		case actualPF >= 0.9: // 0.90-0.94: 每低于0.01减收0.15%
			return (0.9 - actualPF) * 0.15
		case actualPF >= 0.70: // 0.70-0.89: 每低于0.01增收0.5%
			return (0.9 - actualPF) * 0.50
		case actualPF >= 0.65: // 0.65-0.69: 增收10% + 每再低于0.01增收1%
			return 0.1 + (0.7-actualPF)*1
		default: // 0.64及以下: 增收15% + 每再低于0.01增收2%
			return 0.15 + (0.65-actualPF)*2
		}

	case 0.85: // 适用于100kVA以上的其他工业用户
		switch {
		case actualPF >= 0.94: // 0.94及以上: 减收1%
			return -0.01
		case actualPF >= 0.91: // 0.91-0.93: 减收0.65% + 每低于0.01减收0.15%
			return (0.91-actualPF)*0.15 - 0.0065
		case actualPF >= 0.85: // 0.85-0.90: 每低于0.01增收0.1%
			return (0.85 - actualPF) * 0.1
		case actualPF >= 0.65: // 0.65-0.84: 每低于0.01增收0.5%
			return (0.85 - actualPF) * 0.5
		case actualPF >= 0.60: // 0.60-0.64: 增收10% + 每再低于0.01增收1%
			return 0.1 + (0.65-actualPF)*1
		default: // 0.59及以下: 增收15% + 每再低于0.01增收2%
			return 0.15 + (0.6-actualPF)*2
		}

	case 0.8: // 适用于100kVA及以上的农业用户
		switch {
		case actualPF >= 0.92: // 0.92及以上: 减收1.3%
			return -0.013
		case actualPF == 0.91: // 0.91: 减收1.15%
			return -0.0115
		case actualPF >= 0.8: // 0.80-0.90: 每低于0.01增收0.1%
			return (0.8 - actualPF) * 0.1
		case actualPF >= 0.6: // 0.60-0.79: 每低于0.01增收0.5%
			return (0.8 - actualPF) * 0.5
		case actualPF >= 0.55: // 0.55-0.59: 增收10% + 每再低于0.01增收1%
			return 0.1 + (0.6-actualPF)*1
		default: // 0.54及以下: 增收15% + 每再低于0.01增收2%
			return 0.15 + (0.55-actualPF)*2
		}
	}

	return 0 // 不符合任何调整条件
}
