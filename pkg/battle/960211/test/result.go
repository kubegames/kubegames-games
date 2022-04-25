package main

import (
	"fmt"
	"os"
	"time"
)

type result struct {
	name     string // 游戏名称
	roomProb int    // 作弊率

	// ---------------------------分割线-----------------------------
	setTimes  int     // 设定局数
	runTimes  int     // 运行次数（实际运行的局数（金币不足或异常等情况中断））
	winTimes  int     // 赢次数（本局最终收益为正记为赢局）
	winGold   int64   // 赢得金额
	lossTimes int     // 输次数（本局最终收益为负记为输局）
	lossGold  int64   // 输得金额
	heTimes   int     // 和次数（本局最终收益为零记为和局）
	winProb   float64 // 胜率=赢局数/运行局数×100%，精确到小数点后四位

	// ---------------------------分割线-----------------------------
	firstTimes  int     // 1号牌组次数(拿到1号牌组的次数)
	firstProb   float64 // 1号牌组概率(1号牌组次数/运行局数×100%)
	secondTimes int     // 2号牌组次数(拿到2号牌组的次数)
	secondProb  float64 // 2号牌组概率(2号牌组次数/运行局数×100%)

	// ---------------------------分割线-----------------------------
	beginGold   int64   // 初始货币
	endGold     int64   // 结束货币
	changeGold  int64   // 货币变化
	betGold     int64   // 下注额度
	tax         int     // 税点
	winBalance  int64   // 赢局平均赢额(赢局统计中，平均每局赢的额度)
	lossBalance int64   // 输局平均输额(输局统计中，平均每局输的额度)
	allBet      int64   // 总投入(一切出账均纳入记数，包括初始押注、中途追加、结算给出等)
	allWin      int64   // 总收益(一切入账均纳入记数，包括押注退还、结算除税收益等)
	allTax      int64   // 总抽税(所有赢局最终赢额×税点)
	allWinProb  float64 // 收益率（总收益/总投入×100%）
}

func (r result) println() {
	str := fmt.Sprintf(
		`
游戏名称:%v
作弊率:%v

------------分割线-----------
设定局数:%v
实际游戏局数:%v
赢次数:%v
输次数:%v
和次数:%v
胜率:%v

------------分割线-----------
1号牌组次数:%v
1号牌组概率:%v
2号牌组次数:%v
2号牌组概率:%v

------------分割线-----------
初始货币:%v
结束货币:%v
货币变化:%v
下注额度:%v
税点:%v
赢局平均赢额:%v
输局平均输额:%v
总投入:%v
总收益:%v
总抽税:%v
总益率:%v

`, r.name,
		r.roomProb,
		// ------------分割线-----------
		r.setTimes,
		r.runTimes,
		r.winTimes,
		r.lossTimes,
		r.heTimes,
		r.winProb,
		// ------------分割线-----------
		r.firstTimes,
		r.firstProb,
		r.secondTimes,
		r.secondProb,

		// ------------分割线-----------
		r.beginGold,   // 初始货币
		r.endGold,     // 结束货币
		r.changeGold,  // 货币变化
		r.betGold,     // 下注额度
		r.tax,         // 税点
		r.winBalance,  // 赢局平均赢额(赢局统计中，平均每局赢的额度)
		r.lossBalance, // 输局平均输额(输局统计中，平均每局输的额度)
		r.allBet,      // 总投入(一切出账均纳入记数，包括初始押注、中途追加、结算给出等)
		r.allWin,      // 总收益(一切入账均纳入记数，包括押注退还、结算除税收益等)
		r.allTax,      // 总抽税(所有赢局最终赢额×税点)
		r.allWinProb,  // 收益率（总收益/总投入×100%）
	)

	tt := time.Now().Format("20060102150405")
	file, err := os.Create(fmt.Sprintf("%v-%s.txt", r.name, tt))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(str)
}

// 统计
func (r *result) count() {
	// 默认用户0是玩家
	user := userTable[1]

	// ------------分割线-----------
	r.runTimes++
	if user.winGoldActual > 0 {
		r.winTimes++
		r.winGold += user.winGoldActual
		r.allWin += user.winGoldActual
	} else if user.winGoldActual < 0 {
		r.lossTimes++
		r.lossGold += -1 * user.winGoldActual
		r.allBet += -1 * user.winGoldActual
	} else {
		r.heTimes++
	}

	// ------------分割线-----------
	switch user.paiIndex {
	case 1:
		r.firstTimes++
	case 2:
		r.secondTimes++
	}
	r.endGold = user.gold
	r.allTax += user.tax
	r.betGold += int64(user.bet)

}

func (r *result) endCalc() {
	r.winProb = float64(r.winTimes) / float64(r.runTimes) * float64(100)
	r.changeGold = r.endGold - r.beginGold
	r.tax = Conf.Tax
	if r.winTimes != 0 {
		r.winBalance = r.winGold / int64(r.winTimes)
	}
	if r.lossTimes != 0 {
		r.lossBalance = r.lossGold / int64(r.lossTimes)
	}
	r.allWinProb = float64(r.allWin) / float64(r.allBet) * float64(100)
	r.firstProb = float64(r.firstTimes) / float64(r.runTimes) * float64(100)
	r.secondProb = float64(r.secondTimes) / float64(r.runTimes) * float64(100)
}
