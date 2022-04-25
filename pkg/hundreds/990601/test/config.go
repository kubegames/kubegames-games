package main

import (
	"encoding/json"
	"fmt"
	"game_LaBa/birdAnimal/config"
	"game_LaBa/birdAnimal/model"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/tidwall/gjson"
)

type testConfig struct {
	RoomProb     int32 `json:"roomProb"`     // 血池值
	SharkBetProb int32 `json:"sharkBetProb"` // 特殊下注概率（万分率）
	TestTimes    int64 `json:"testTimes"`    // 测试次数
	InitGold     int64 `json:"initGold"`     // 初始货币
	TakeGold     int64 // 携带金币
	Tax          int64 `json:"tax"`     // 税点
	BetGold      int64 `json:"betGold"` // 下注额度

	// BirdAreaNum   int64 `json:"bridAreaaNum"`  // 下飞禽区域数量（0-4）超过4最多下4个区域
	// AnimalAreaNum int64 `json:"animalAreaNum"` // 下走兽区域数量

	BetArea [10]int64 `json:"betArea"` // 下注区的下注倍数
}

var Conf testConfig

func init() {
	data, err := ioutil.ReadFile("./testconfig.json")
	if err != nil {
		log.Println("open file error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	json.Unmarshal([]byte(result.String()), &Conf)
	Conf.RoomProb = int32(result.Get("roomProb").Int())
	Conf.SharkBetProb = int32(result.Get("sharkBetProb").Int())
	Conf.TestTimes = int64(result.Get("testTimes").Int())
	Conf.InitGold = int64(result.Get("initGold").Int())
	Conf.Tax = int64(result.Get("tax").Int())
	Conf.BetGold = int64(result.Get("betGold").Int())

	betArea := result.Get("betArea").Array()
	if len(betArea) != 10 {
		panic("betArea length must be 10")
	}
	for i, v := range betArea {
		Conf.BetArea[i] = v.Int()
	}
	checkTestConfig()
	initTestConfig()
}

// 初始化config文件
func initTestConfig() {
	Conf.TakeGold = Conf.InitGold
}

// 检查config文件
func checkTestConfig() {
	if Conf.TestTimes <= 0 {
		Conf.TestTimes = 10000
	}
	if Conf.InitGold <= 0 {
		Conf.InitGold = 100000
	}
	if Conf.Tax < 0 {
		Conf.Tax = 5
	}
	if Conf.BetGold <= 0 {
		Conf.BetGold = 10
	}
	switch Conf.RoomProb {
	case -3000, -2000, -1000, 1000, 2000, 3000:
	default:
		log.Fatalln("roomProb must be one of [-3000,-2000,-1000,1000,2000,3000]")
	}
}

type TestResult struct {
	EndGold      int64   // 结束货币
	GoldExchange int64   // 货币变化
	GoldShark    int64   // 金鲨次数
	SilverShark  int64   // 银鲨次数
	AllKill      int64   // 通杀次数
	AllPay       int64   // 通赔次数
	AllBet       int64   // 总投入
	AllWin       int64   // 总收益
	AllTax       float64 // 总抽税
	WinProb      float64 // 收益率

	Bird   int64 // 飞禽类
	Animal int64 // 走兽类

	TestTimes int64 // 真实下注次数

	UserWinTimes int64 // 玩家赢的次数

	NotInBack int64 // 不在返奖率范围中的次数
}

func (tr TestResult) write2file() {
	str := fmt.Sprintf(
		"初始货币：%d\n结束货币：%d\n货币变化：%d\n金鲨次数：%d\n银鲨次数：%d\n通杀次数：%d\n通赔次数：%d\n飞禽：%d\n走兽：%d\n总投入：%d\n总收益：%d\n总抽税：%f\n收益率：%f%%\n实际游戏次数：%d\n玩家胜率：%f%%\n不在随机返奖率区间内次数：%d\n",
		Conf.InitGold,
		Conf.TakeGold+tr.AllWin,
		tr.GoldExchange,
		tr.GoldShark,
		tr.SilverShark,
		tr.AllKill,
		tr.AllPay,
		tr.Bird,
		tr.Animal,
		tr.AllBet,
		tr.AllWin,
		tr.AllTax,
		tr.WinProb,
		tr.TestTimes+1,
		(float64(tr.UserWinTimes)/float64(tr.TestTimes))*100,
		tr.NotInBack,
	)

	name := fmt.Sprintf("飞禽走兽测试结果-%s.txt", time.Now().Format("20060102150405"))
	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	file.Write([]byte(str))
}

// 计算测试结果
func (tr *TestResult) calc() {
	tr.GoldExchange = Conf.TakeGold - Conf.InitGold
	tr.AllTax = float64(Conf.Tax) / float64(100) * float64(tr.AllWin)
	tr.WinProb = (float64(tr.AllWin) - tr.AllTax) / float64(tr.AllBet) * float64(100)
	// tr.WinProb=float64(tr.)
}

func getOdds(i int) int {
	// if i < 0 || i >= 12 {
	// 	return 2
	// }
	return config.BirdAnimaConfig.BirdAnimals.GetByID(i).OddsNow
}

func allBet() (gold int64) {
	for _, v := range TotalBet {
		gold += v
	}
	return
}

// 统计
func (tr *TestResult) count(result []*model.Element) {
	var win int64
	allBetGold := allBet()

	oddsList := getRandOddsInfo()
	// 只有一个结果
	if len(result) == 1 {
		switch result[0].ID {
		case ALL_KILL_ID:
			tr.AllKill++
			win = 0
		case ALL_PAY_ID:
			tr.AllPay++
			for index, v := range TotalBet {
				win += v * int64(int64(oddsList[index].Odds))
			}
		case 0, 1, 6, 7:
			tr.Bird++
			win = TotalBet[result[0].ID] * int64(oddsList[result[0].ID].Odds)
			win += TotalBet[BIRD_BET_INDEX] * int64(2)
		case 4, 5, 10, 11:
			tr.Animal++
			win = TotalBet[result[0].ID] * int64(oddsList[result[0].ID].Odds)
			win += TotalBet[ANIMAL_BET_INDEX] * int64(2)
		default:
			win = TotalBet[result[0].ID] * int64(oddsList[result[0].ID].Odds)
		}
		tr.AllWin += win
		if win > allBetGold {
			tr.UserWinTimes++
		}
		return
	}

	var win1, win2 int64
	switch result[0].ID {
	case GOLD_SHARK_ID:
		tr.GoldShark++
		win1 = TotalBet[result[0].ID] * int64(oddsList[result[0].ID].Odds)
		tr.AllWin += win1
	case SILVER_SHARK_ID:
		tr.SilverShark++
		win1 = TotalBet[result[0].ID] * int64(oddsList[result[0].ID].Odds)
		tr.AllWin += win1
	}

	win2 = TotalBet[result[1].ID] * int64(oddsList[result[1].ID].Odds)
	switch result[1].ID {
	case 0, 1, 6, 7:
		tr.Bird++
		win += TotalBet[BIRD_BET_INDEX] * int64(2)
	case 4, 5, 10, 11:
		tr.Animal++
		win += TotalBet[ANIMAL_BET_INDEX] * int64(2)
	}
	tr.AllWin += win2
	if (win1 + win2) > allBetGold {
		tr.UserWinTimes++
	}
	return

}
