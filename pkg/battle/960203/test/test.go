package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960203/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/sipt/GoJsoner"
)

// 参数
type Params struct {
	ControlCfg struct {
		ControlRate      []int32 `json:"control_rate"`       //  作弊率等级
		BiggestCardsRate []int   `json:"biggest_cards_rate"` // 最大牌概率分布
		SecondCardsRate  []int   `json:"second_cards_rate"`  // 第二大牌概率分布
	} `json:"control_cfg"` // 作弊率配置表

	BetMultiple int64 `json:"bet_multiple"` // 默认投注倍数
	RobMultiple int64 `json:"rob_multiple"` // 默认抢庄倍数
	RoomProb    int32 `json:"room_prob"`    // 作弊率
	InitAmount  int64 `json:"init_amount"`  // 初始金额
	LoopCount   int   `json:"loop_count"`   // 循环局数
	Tax         int64 `json:"tax"`          // 税点 百分率
	RoomCost    int64 `json:"room_cost"`    // 底注
	RobRate     int   `json:"rob_rate"`     // 抢庄概率 百分率
}

var params Params

// testUser 测试玩家
type testUser struct {
	ID        int64 // id
	CurAmount int64 // 当前拥有资金
}

// testGame 测试游戏桌
type testGame struct {
	UserList        map[int64]*testUser        // 游戏列表
	Banker          *testUser                  // 庄家
	Poker           *poker.GamePoker           // 牌堆
	ControlledCards map[int64]*poker.HoldCards // 控制的牌堆
	CardsSequence   []*poker.HoldCards         // 牌组序列
}

// outData 输出数据
type outData struct {
	roomProb     int32   // 作弊率
	gameCount    int     // 测试局数
	initAmount   int64   // 初始货币
	tax          int64   // 税点
	betAmount    int64   // 下注额度
	robRate      int     // 抢庄概率
	finalAmount  int64   // 结束货币
	changeAmount int64   // 变化货币 = 结束货币 - 初始货币
	winCount     int     // 赢局数
	lossCount    int     // 输局数
	drawCount    int     // 和局数
	winRate      float64 // 胜率
	bankerCount  int     // 庄家局数
	biggestCount int     // 1号牌组次数
	biggestRate  float64 // 1号牌组概率 = 1号牌组次数 / 局数 * 100%
	secondCount  int     // 2号牌组次数
	secondRate   float64 // 2号牌组概率 = 2号牌组次数 / 局数 * 100%
	totalIn      int64   // 总投入
	totalOut     int64   // 总收益
	totalTaxes   float64 // 总税收
	returnRate   float64 // 收益率 = 总收益/中投入 * 100%
}

var data outData

func main() {

	LoadCfg()

	// 初始化一个ID为1，资金为100000的用户
	player := &testUser{
		ID:        1,
		CurAmount: params.InitAmount,
	}

	data = outData{
		roomProb:     params.RoomProb,
		gameCount:    0,
		initAmount:   params.InitAmount,
		tax:          params.Tax,
		betAmount:    params.RoomCost,
		robRate:      params.RobRate,
		finalAmount:  params.InitAmount,
		changeAmount: 0,
		winCount:     0,
		lossCount:    0,
		drawCount:    0,
		winRate:      0,
		bankerCount:  0,
		biggestCount: 0,
		biggestRate:  0,
		secondCount:  0,
		secondRate:   0,
		totalIn:      0,
		totalOut:     0,
		totalTaxes:   0,
		returnRate:   0,
	}

	// 初始化游戏桌子
	game := new(testGame)
	game.UserList = make(map[int64]*testUser)

	// 开始游戏
	game.startGame(player)

}

// LoadCfg 加载配置文件
func LoadCfg() {
	data, err := ioutil.ReadFile("./test.json")
	if err != nil {
		log.Errorf("File reading error : %v", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))

	if err := json.Unmarshal([]byte(result), &params); err != nil {
		log.Errorf("Unmarshal json error : %v", err)
		return
	}

}

// startGame 开始游戏
func (game *testGame) startGame(player *testUser) {
	// 清空桌子
	game.Poker = nil
	game.CardsSequence = []*poker.HoldCards{}
	game.ControlledCards = make(map[int64]*poker.HoldCards)
	for id := range game.UserList {
		delete(game.UserList, id)
	}

	// 玩家加入游戏列表
	game.UserList[player.ID] = player

	// 添加三个机器人, id 为2，3，4
	for i := 2; i <= 4; i++ {
		robotID := int64(i)
		robot := &testUser{
			ID:        robotID,
			CurAmount: params.InitAmount,
		}
		game.UserList[robotID] = robot
	}

	// 定庄
	game.confirmBanker()

	// 设置牌组序列
	game.SetCardsSequence()

	// 配牌
	game.ControlCards()

	// 结算
	game.Settle()

	// 结束
	game.endGame()
}

// confirmBanker 定庄
func (game *testGame) confirmBanker() {

	var (
		candidates []int64 // 抢庄候选人
		bankerID   int64   // 庄家ID
	)
	weight := rand.RandInt(1, 101)

	if weight <= data.robRate {
		bankerID = 1

		data.bankerCount++
	} else {
		for id := range game.UserList {
			if id == 1 {
				continue
			}
			candidates = append(candidates, id)
		}
		index := rand.RandInt(0, len(candidates))

		bankerID = candidates[index]

	}

	// 定庄
	game.Banker = game.UserList[bankerID]
}

// SetCardsSequence 设置牌组序列
func (game *testGame) SetCardsSequence() {
	// 洗牌
	game.Poker = new(poker.GamePoker)
	game.Poker.InitPoker()
	for i := 0; i < len(game.UserList); i++ {
		cards := game.Poker.DrawCard()
		holdCards := &poker.HoldCards{
			Cards:     cards,
			CardsType: poker.GetCardsType(cards),
		}

		// 牌组序列为空，加入一张牌组
		if len(game.CardsSequence) == 0 {
			game.CardsSequence = append(game.CardsSequence, holdCards)
			continue
		}

		// 插入排序法从小到大依次排列
		var newSequence []*poker.HoldCards
		for k, v := range game.CardsSequence {

			// holdCards < v
			if poker.ContrastCards(holdCards, v) {

				rear := append([]*poker.HoldCards{}, game.CardsSequence[k:]...)
				newSequence = append(append(game.CardsSequence[:k], holdCards), rear...)
				break
			} else {
				newSequence = append(game.CardsSequence, holdCards)
			}
		}
		game.CardsSequence = newSequence
	}
}

// ControlCards 配牌
func (game *testGame) ControlCards() {
	// 最大牌，和第二大牌概率分布
	biggestCardsRatePlace := make(map[int64]int)
	secondCardsRatePlace := make(map[int64]int)

	for id := range game.UserList {
		roomProb := data.roomProb

		// 机器人取反
		if id > 1 {
			roomProb = -roomProb
		}
		probIndex := game.checkProb(roomProb)
		if probIndex == -1 {
			// 默认 1000 作弊率的 索引
			probIndex = 2

			// 机器人取反
			if id > 1 {
				probIndex = 4
			}
		}

		biggestCardsRatePlace[id] = params.ControlCfg.BiggestCardsRate[probIndex]
		secondCardsRatePlace[id] = params.ControlCfg.SecondCardsRate[probIndex]
	}

	// 总概率值
	var totalRate int

	// 总概率值
	for _, rate := range biggestCardsRatePlace {
		totalRate += rate
	}

	// 未满10000的剩余概率值剩余平均概率
	lessAverageRate := (10000 - totalRate) / len(biggestCardsRatePlace)

	if lessAverageRate < 0 {
		lessAverageRate = 0
	}

	// 更新新概率值，让概率变得更加平缓
	for id, rate := range biggestCardsRatePlace {
		biggestCardsRatePlace[id] = lessAverageRate + rate
		totalRate += lessAverageRate
	}

	// 权重
	weight := rand.RandInt(1, totalRate+1)

	// 概率累加值
	addRate := 0

	// 最大牌userID
	var biggestCardsUserID int64
	for id, rate := range biggestCardsRatePlace {

		if weight > addRate && weight <= addRate+rate {
			biggestCardsUserID = id
			break
		}
		addRate += rate
	}

	// 分配最大牌
	game.ControlledCards[biggestCardsUserID] = game.CardsSequence[len(game.CardsSequence)-1]
	game.CardsSequence = game.CardsSequence[:len(game.CardsSequence)-1]

	if biggestCardsUserID == 0 {
		fmt.Printf("权重 %d", weight)
		fmt.Printf("第一大牌分布 %v", biggestCardsRatePlace)
	}

	// 踢出已那最大牌用户的概率
	delete(secondCardsRatePlace, biggestCardsUserID)

	// 最大牌计数加1
	if biggestCardsUserID == 1 {
		data.biggestCount++
	}

	// 重置总概率值
	totalRate = 0

	// 计算拿第二大牌的总概率值
	for _, rate := range secondCardsRatePlace {
		totalRate += rate
	}

	// 剩余平均概率
	lessAverageRate = (10000 - totalRate) / len(secondCardsRatePlace)

	if lessAverageRate < 0 {
		lessAverageRate = 0
	}

	// 更新新概率值，让概率变得更加平缓
	for id, rate := range secondCardsRatePlace {
		secondCardsRatePlace[id] = lessAverageRate + rate
		totalRate += lessAverageRate
	}

	// 权重
	weight = rand.RandInt(1, totalRate+1)

	// 概率累加值
	addRate = 0

	// 最二大牌userID
	var secondCardsUserID int64
	for id, rate := range secondCardsRatePlace {

		if weight > addRate && weight <= addRate+rate {
			secondCardsUserID = id
			break
		}
		addRate += rate
	}

	// 分配最第二大牌
	game.ControlledCards[secondCardsUserID] = game.CardsSequence[len(game.CardsSequence)-1]
	game.CardsSequence = game.CardsSequence[:len(game.CardsSequence)-1]

	if secondCardsUserID == 0 {
		fmt.Printf("权重 %d", weight)
		fmt.Printf("第二大牌分布 %v", secondCardsRatePlace)
	}

	// 第二大牌计数加1
	if secondCardsUserID == 1 {
		data.secondCount++
	}

	// 分配剩下的牌
	for id := range game.UserList {
		if _, ok := game.ControlledCards[id]; !ok {
			index := rand.RandInt(0, len(game.CardsSequence))
			game.ControlledCards[id] = game.CardsSequence[index]
			game.CardsSequence = append(game.CardsSequence[:index], game.CardsSequence[index+1:]...) // 删除中间1个元素
		}
	}
}

// ControlCards 结算
func (game *testGame) Settle() {

	// 玩家结果，庄家结果，净盈利
	var playerResult, bankerResult, netprofit int64

	for id := range game.UserList {
		if id == game.Banker.ID {
			continue
		}

		// 牌倍数
		var cardsMultiple int64
		if id == game.Banker.ID {
			fmt.Print(game.Banker.ID)
			fmt.Print(id)
			fmt.Print("\n")
		}

		if poker.ContrastCards(game.ControlledCards[game.Banker.ID], game.ControlledCards[id]) {
			// 闲家
			cardsMultiple = poker.GetCardsMultiple(game.ControlledCards[id].CardsType)
		} else {
			// 闲家输
			cardsMultiple = poker.GetCardsMultiple(game.ControlledCards[game.Banker.ID].CardsType) * -1
		}

		// 闲家输赢
		result := params.RoomCost * params.RobMultiple * params.BetMultiple * cardsMultiple
		if id == 1 {
			playerResult = result
		}

		bankerResult += -result
	}

	if game.Banker.ID == 1 {
		playerResult = bankerResult
	}

	// 统计总投入
	if playerResult < 0 {
		data.lossCount++
		data.totalIn += -1 * playerResult
		netprofit = playerResult
	}

	if playerResult == 0 {
		data.drawCount++
	}

	// 统计总收益，总抽税，统计
	if playerResult > 0 {
		data.winCount++
		data.totalOut += playerResult
		// 税
		taxes := float64(playerResult) * 5 / 100
		// 净盈利
		netprofit = playerResult - int64(math.Ceil(taxes))
		data.totalTaxes += taxes
	}
	//fmt.Println(playerResult)

	// 统计结束金额
	data.finalAmount += netprofit
}

// ControlCards 结算
func (game *testGame) endGame() {

	// 游戏局数+1
	data.gameCount++

	// 变化货币
	data.changeAmount = data.finalAmount - data.initAmount

	// 胜率
	data.winRate = float64(data.winCount) * 100 / float64(data.gameCount)

	// 1号牌组概率
	data.biggestRate = float64(data.biggestCount) * 100 / float64(data.gameCount)

	// 2号牌组概率
	data.secondRate = float64(data.secondCount) * 100 / float64(data.gameCount)

	// 收益率
	if data.totalIn > 0 {
		data.returnRate = float64(data.totalOut) * 100 / float64(data.totalIn)
	}

	player := &testUser{
		ID:        1,
		CurAmount: data.finalAmount,
	}

	// 开启下一局
	if data.gameCount < params.LoopCount && data.finalAmount > 0 {
		game.startGame(player)
	} else {
		outPutData()
	}
}

// 检测作弊率
func (game *testGame) checkProb(prob int32) (probIndex int) {
	probIndex = -1
	for index, rate := range params.ControlCfg.ControlRate {
		if prob == rate {
			probIndex = index
		}
	}

	return
}

func outPutData() {
	fileName := "test.log"
	_, err := os.Stat(fileName)
	if err == nil {
		os.Remove(fileName)
	}

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	output := "看牌抢庄牛牛日志\n"

	output += "作弊率:" + fmt.Sprintf(`%d`, data.roomProb) + "\n"
	output += "测试局数:" + fmt.Sprintf(`%d`, data.gameCount) + "\n"
	output += "初始货币:" + fmt.Sprintf(`%d`, data.initAmount) + "\n"
	output += "税点:" + fmt.Sprintf(`%d`, data.tax) + "%\n"
	output += "下注额度:" + fmt.Sprintf(`%d`, data.betAmount) + "\n"
	output += "抢庄概率:" + fmt.Sprintf(`%d`, data.robRate) + "%\n"
	output += "结束货币:" + fmt.Sprintf(`%d`, data.finalAmount) + "\n"
	output += "货币变化:" + fmt.Sprintf(`%d`, data.changeAmount) + "\n"
	output += "赢局数:" + fmt.Sprintf(`%d`, data.winCount) + "\n"
	output += "输局数:" + fmt.Sprintf(`%d`, data.lossCount) + "\n"
	output += "和局数:" + fmt.Sprintf(`%d`, data.drawCount) + "\n"
	output += "胜率:" + fmt.Sprintf(`%.2f`, data.winRate) + "%\n"
	output += "庄家局数:" + fmt.Sprintf(`%d`, data.bankerCount) + "\n"
	output += "1号牌组次数:" + fmt.Sprintf(`%d`, data.biggestCount) + "\n"
	output += "1号牌组概率:" + fmt.Sprintf(`%.2f`, data.biggestRate) + "\n"
	output += "2号牌组次数:" + fmt.Sprintf(`%d`, data.secondCount) + "\n"
	output += "2号牌组概率:" + fmt.Sprintf(`%.2f`, data.secondRate) + "\n"
	output += "总投入:" + fmt.Sprintf(`%d`, data.totalIn) + "\n"
	output += "总收益:" + fmt.Sprintf(`%d`, data.totalOut) + "\n"
	output += "总抽税:" + fmt.Sprintf(`%.2f`, data.totalTaxes) + "\n"
	output += "收益率:" + fmt.Sprintf(`%.2f`, data.returnRate) + "%\n"
	content := []byte(output)
	_, err = file.Write(content)
	if err != nil {
		fmt.Println(err)
	}
}
