// 辅助函数
package game

import (
	"strconv"
	"strings"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/config"
)

type SettleResult struct {
	TheorySettle int64 // 理论结算值
	ActualSettle int64 // 实际结算值
	CurAmount    int64 // 携带金额
}

// FillWinnerAmount 折算多余赢家金额，补足应赢金额小于携带金额但是 按比例补足会触发防一小博大，则补到携带金额大小
// leftAmount 剩余多赢金额
// leftTheoryWinSum 剩余理论赢钱合值
// WinnerList 赢家列表
func FillWinnerAmount(leftAmount *int64, leftTheoryWinSum *int64, WinnerList map[int64]*SettleResult) {

	newLeftAmount := *leftAmount
	// 按比例折扣 会 触发防止一小博大机制，则先补足金额
	for userID, v := range WinnerList {

		if v.ActualSettle >= v.CurAmount {
			continue
		}

		// 按比例折算金额
		convertAmount := *leftAmount * v.TheorySettle / *leftTheoryWinSum

		if convertAmount+v.ActualSettle > v.CurAmount {
			*leftAmount -= v.CurAmount - v.ActualSettle
			WinnerList[userID].ActualSettle = v.CurAmount
			*leftTheoryWinSum -= v.TheorySettle
			break
		}
	}

	// 剩余钱无变化，不需要再补足，跳出循环
	if newLeftAmount != *leftAmount {
		FillWinnerAmount(leftAmount, leftTheoryWinSum, WinnerList)
	}
}

// ConvertWinnerAmount 按比例折算多余赢家金额
func ConvertWinnerAmount(leftAmount int64, leftTheoryWinSum int64, WinnerList map[int64]*SettleResult) {
	if leftAmount <= 0 || leftTheoryWinSum <= 0 {
		log.Errorf("按比例折算多余赢家金额出现错误，剩余金额：%d，剩余赢钱理论合值：%d", leftAmount, leftTheoryWinSum)
		return
	}

	var (
		convertCount int   // 需要补足多余金额玩家个数
		winCounter   int   // 赢家计数器
		winAcc       int64 // 赢钱累加器
	)

	for _, v := range WinnerList {
		if v.ActualSettle < v.CurAmount {
			convertCount++
		}
	}

	// 按比例折扣
	for userID, v := range WinnerList {

		if v.ActualSettle < v.CurAmount {

			// 最后一个需要补足多余金额到赢家
			if convertCount-winCounter == 1 {
				WinnerList[userID].ActualSettle += leftAmount - winAcc
				break
			}

			// 按比例折算金额
			convertAmount := leftAmount * v.TheorySettle / leftTheoryWinSum
			WinnerList[userID].ActualSettle += convertAmount
			winAcc += convertAmount
			winCounter++
		}
	}
}

// FillLoserAmount 折算多余输家金额，补足应输金额小于携带金额但是 按比例补足会触发防一小博大，则补到携带金额大小
// leftAmount 剩余多输金额
// leftTheoryLoseSum 剩余理论输钱合值
// LoserList 输家列表
func FillLoserAmount(leftAmount *int64, leftTheoryLoseSum *int64, LoserList map[int64]*SettleResult) {

	newLeftAmount := *leftAmount
	// 按比例折扣 会 触发防止一小博大机制，则先补足金额
	for userID, v := range LoserList {

		if v.ActualSettle <= -v.CurAmount {
			continue
		}

		// 按比例折算金额
		convertAmount := *leftAmount * v.TheorySettle / *leftTheoryLoseSum

		if convertAmount+v.ActualSettle < -v.CurAmount {
			*leftAmount += v.CurAmount + v.ActualSettle
			LoserList[userID].ActualSettle = -v.CurAmount
			*leftTheoryLoseSum -= v.TheorySettle
			break
		}
	}

	// 剩余钱无变化，不需要再补足，跳出循环
	if newLeftAmount != *leftAmount {
		FillLoserAmount(leftAmount, leftTheoryLoseSum, LoserList)
	}
}

// ConvertLoserAmount 按比例折算多余输家金额
func ConvertLoserAmount(leftAmount int64, leftTheoryLoseSum int64, LoserList map[int64]*SettleResult) map[int64]*SettleResult {
	if leftAmount >= 0 || leftTheoryLoseSum >= 0 {
		log.Errorf("按比例折算多余输家金额出现错误，剩余金额：%d，剩余输钱理论合值：%d", leftAmount, leftTheoryLoseSum)
		return LoserList
	}

	var (
		convertCount int   // 需要补足多余金额玩家个数
		loseCounter  int   // 输家计数器
		loseAcc      int64 // 输钱累加器
	)

	for _, v := range LoserList {
		if v.ActualSettle > -v.CurAmount {
			convertCount++
		}
	}

	// 按比例折扣
	for userID, v := range LoserList {

		if v.ActualSettle > -v.CurAmount {

			// 最后一个需要补足多余金额到输家
			if convertCount-loseCounter == 1 {
				LoserList[userID].ActualSettle += leftAmount - loseAcc
				break
			}

			// 按比例折算金额
			convertAmount := leftAmount * v.TheorySettle / leftTheoryLoseSum
			LoserList[userID].ActualSettle += convertAmount
			loseAcc += convertAmount
			loseCounter++
		}
	}

	return LoserList
}

// PaoMaDeng 跑马灯
func (game *Game) PaoMaDeng(Gold int64, userInter player.PlayerInterface, special string) {
	configs := game.Table.GetMarqueeConfig()

	for _, conf := range configs {

		if len(conf.SpecialCondition) > 0 && len(special) > 0 {
			log.Tracef("跑马灯有特殊条件 : %s", conf.SpecialCondition)

			specialIndex := game.GetSpecialSlogan(special)
			specialArr := strings.Split(conf.SpecialCondition, ",")
			for _, specialStr := range specialArr {
				specialInt, err := strconv.Atoi(specialStr)
				if err != nil {
					log.Errorf("解析跑马灯特殊条件出错 : %s", conf.SpecialCondition)
					continue
				}

				// 金额与特殊条件同时满足
				if specialInt == specialIndex && Gold >= conf.AmountLimit {
					err := game.Table.CreateMarquee(userInter.GetNike(), Gold, special, conf.RuleId)
					if err != nil {
						log.Errorf("创建跑马灯错误：%v", err)
					}
					return
				}
			}
		}
	}

	// 未触发特殊条件
	for _, conf := range configs {
		if len(conf.SpecialCondition) > 0 {
			continue
		}

		// 只需要满足金钱条件
		if Gold >= conf.AmountLimit {
			err := game.Table.CreateMarquee(userInter.GetNike(), Gold, special, conf.RuleId)
			if err != nil {
				log.Errorf("创建跑马灯错误：%v", err)
			}
			return
		}
	}
}

// GetSpecialSlogan 获取跑马灯触发特殊条件下标
func (game *Game) GetSpecialSlogan(special string) int {
	switch special {
	case "至尊青龙":
		return 1
	case "一条龙":
		return 2
	default:
		return 0
	}
}

// checkProb 检测作弊率
func (game *Game) checkProb(prob int32) (probIndex int) {
	probIndex = -1
	for index, rate := range config.CheatConf.ControlRate {
		if prob == rate {
			probIndex = index
		}
	}

	return
}
