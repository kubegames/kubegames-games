package test

import (
	"fmt"
	"io/ioutil"
	"strconv"

	//"strconv"
	"go-game-sdk/example/game_LaBa/990701/gamelogic"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"os"
	"strings"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/sipt/GoJsoner"
)

var lineNum = 30

type testconf struct {
	testtime    int //测试次数
	poolinit    int //血池初始值
	bet         int //下注值
	cheatvalue  int //作弊值
	tax         int //税收万分比
	userwintax  int //税收万分比
	jackpot     int //每次投注加入奖池的钱万分比
	jackpotgold int //奖池初始值
	poolprint   int //血池打印间隔
}

//测试结果统计
type testres struct {
	tc testconf //测试配置
	//概况
	testtime    int     //总测试次数
	totalbet    int     //总下注额
	totalreturn int     //总的返奖
	totaltax    float64 //总税收
	outscore    int     //总吐分
	eatscore    int     //总吃分
	pooltax     float64 //血池税收
	bettax      float64 //押注税收
	//免费游戏
	freegamegetgold         int     //免费游戏获取总金币
	freegametimes           int     //免费游戏总次数
	enterfreegametimescount int     //进入免费游戏次数统计
	freegametype            [3]int  //免费游戏类型统计,5,8,10触发的次数
	freeodds                [11]int //免费游戏倍数区间
	awardcount              int     //中奖次数统计
	//小游戏统计
	littlegamegetgold    int    //小游戏总获取金币
	enterlittlegametimes int    //小游戏触发次数
	littlegametype       [3]int //小游戏类型1,2,3次
	littlegameodds       [8]int //小游戏返奖区间统计
	//特殊游戏总返奖
	specialgamegetgold    int //小游戏总获取金币
	enterspecialgametimes int //小游戏触发次数
	//specialgametype       [3]int //小游戏类型1,2,3次
	specialgameodds [11]int //小游戏返奖区间统计

	//彩金游戏统计
	jackpotgetgold    float64 //玩家获取的彩金金币
	enterjackpottimes int     //彩金游戏触发次数
	jackpotgametype   [3]int  //彩金游戏类型1,2,3次

	odds [11]int //常规倍数区间

	normalgametime int //普通游戏次数
	normalgetgold  int //普通游戏获取金币
	normalgamerate int //普通游戏中奖率
	freegamerate   int //免费游戏中奖率

	jackpotgold float64 //奖池值
	pool        float64 //血池的值
	SpCount     int
}

func (t *testconf) gettestconfig() {
	data, err := ioutil.ReadFile("test.txt")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}

	result, _ := GoJsoner.Discard(string(data))
	strarr := strings.Split(result, ",")
	i := 0
	t.testtime, err = strconv.Atoi(strarr[i])
	i++
	t.poolinit, _ = strconv.Atoi(strarr[i])
	i++
	t.bet, _ = strconv.Atoi(strarr[i])
	i++
	t.cheatvalue, _ = strconv.Atoi(strarr[i])
	i++
	t.tax, _ = strconv.Atoi(strarr[i])
	i++
	t.userwintax, _ = strconv.Atoi(strarr[i])
	i++
	t.jackpot, _ = strconv.Atoi(strarr[i])
	i++
	t.jackpotgold, _ = strconv.Atoi(strarr[i])
	i++
	t.poolprint, _ = strconv.Atoi(strarr[i])
	i++
}

func Test(lb *config.LabaConfig) {
	var tres testres
	tres.tc.gettestconfig()
	total := tres.tc.testtime

	tres.pool = float64(tres.tc.poolinit)
	tres.jackpotgold = float64(tres.tc.jackpotgold)
	var g gamelogic.Game
	g.Init(lb)

	os.Remove("./blood.txt")
	file, _ := os.OpenFile("./blood.txt", os.O_RDWR|os.O_CREATE, 0766)

	for i := 0; i < total; i++ {
		//累加测试次数
		tres.testtime++
		freegame := g.FreeGameTimes

		bfree := g.FreeGameTimes > 0
		//这里算的图形
		g.GetIconRes(int64(tres.tc.cheatvalue))
		//如果两次免费不一样就表示新的免费
		tmp := g.FreeGameTimes - freegame
		if tmp != 0 && tmp != 5 && tmp != 8 && tmp != 10 {
			log.Traceln("免费游戏次数", tmp)
		}
		if freegame != g.FreeGameTimes {
			total += g.FreeGameTimes - freegame
			tres.freegamefun(g.FreeGameTimes - freegame)
			if bfree {
				//免费游戏触发免费游戏
				log.Traceln("免费游戏触发免费游戏")
				select {}
			}
		}
		//log.Traceln("longmusuiji", len(g.WildArr),g.WildArr)
		if bfree {
			g.FreeGameTimes--
			//统计免费游戏
			tres.freegame(g.GetIconOdds())
		} else {
			//下注为单线乘以总线数
			tres.userbet(tres.tc.bet * lb.LineCount)
			if len(g.WildArr) == 3 {
				//log.Traceln("龙母")
				tres.specialgame(g.GetIconOdds())
			} else {
				//统计普通游戏
				tres.normalgame(g.GetIconOdds())
			}

		}

		//彩金游戏统计
		if false {
			gold := g.GetJackpotMoney(int64(tres.jackpotgold))
			tax := float64(gold) * float64(tres.tc.userwintax) / 10000.0
			tres.pool -= float64(gold) + tax
			tres.pooltax += tax
			tres.totaltax += tax
			tres.totalreturn += int(gold)
			tres.jackpotgetgold += gold
			if gold > 0 {
				tres.enterjackpottimes++
			}
		}

		if i%tres.tc.poolprint == 0 {
			str := fmt.Sprintf("%f,\r", tres.pool)
			file.WriteString(str)
		}

	}

	tres.SpCount = g.GetIconCount()
	file.Close()
	tres.writefile()
}

func (tres *testres) userbet(bet int) {
	jackpot := float64(bet) * float64(tres.tc.jackpot) / 10000.0
	tax := float64(bet) * float64(tres.tc.tax) / 10000.0
	tres.totalbet += bet
	tres.totaltax += tax
	tres.bettax += tax
	tres.pool += float64(bet) - tax - jackpot
}

func (tres *testres) freegamefun(times int) {
	tres.enterfreegametimescount++
	tres.freegametimes += times
	switch times {
	case 5:
		tres.freegametype[0]++
		break
	case 10:
		tres.freegametype[1]++
		break
	case 15:
		tres.freegametype[2]++
		break
	}
}

//每次免费游戏统计
func (tres *testres) freegame(odds int) {
	if odds > 0 {
		tres.awardcount++
	}
	tres.subpool(odds)
	win := tres.tc.bet * odds
	tres.freegamegetgold += win
	arr := [...]int{0, lineNum * 1, lineNum * 3, lineNum * 5, lineNum * 10, lineNum * 30, lineNum * 50, lineNum * 100, lineNum * 200, lineNum * 500, 99999999}
	// arr := [...]int{0, 1, 3, 5, 10, 30, 50, 100, 200, 500,9999999}
	for i := 0; i < len(arr); i++ {
		if odds <= arr[i] {
			tres.freeodds[i]++
			break
		}
	}
}

func (tres *testres) littlegamefun(times int) {
	tres.enterlittlegametimes++
	switch times {
	case 1:
		tres.littlegametype[0]++
		break
	case 2:
		tres.littlegametype[1]++
		break
	case 3:
		tres.littlegametype[2]++
		break
	}
}

func (tres *testres) littlegame(odds int) {
	if odds > 0 {
		//tres.awardcount++
	}
	tres.subpool(odds)
	tres.littlegamegetgold += tres.tc.bet * odds
	arr := [...]int{0, lineNum * 1, lineNum * 3, lineNum * 5, lineNum * 10, lineNum * 20, lineNum * 50, 99999999}
	//arr2 := [...]int{0, 1, 3, 5, 10, 20, 50, 99999999}
	for i := 0; i < len(arr); i++ {
		if odds <= arr[i] {
			tres.littlegameodds[i]++
			break
		}
	}
}

func (tres *testres) normalgame(odds int) {
	if odds > 0 {
		tres.awardcount++
	}
	tres.subpool(odds)
	tres.normalgetgold += odds * tres.tc.bet
	arr := [...]int{0, lineNum * 1, lineNum * 3, lineNum * 5, lineNum * 10, lineNum * 30, lineNum * 50, lineNum * 100, lineNum * 200, lineNum * 500, 99999999}
	//	arr3 := [...]int{0, 1, 3, 5, 10, 30, 50, 100, 200, ,500,99999999}
	for i := 0; i < len(arr); i++ {
		if odds <= arr[i] {
			tres.odds[i]++
			break
		}
	}
}
func (tres *testres) specialgame(odds int) {
	if odds > 0 {
		tres.awardcount++
	}
	tres.enterspecialgametimes++
	tres.subpool(odds)
	tres.specialgamegetgold += odds * tres.tc.bet
	arr := [...]int{0, lineNum * 1, lineNum * 3, lineNum * 5, lineNum * 10, lineNum * 30, lineNum * 50, lineNum * 100, lineNum * 200, lineNum * 500, 99999999}
	//	arr3 := [...]int{0, 1, 3, 5, 10, 30, 50, 100, 200,500, 99999999}
	for i := 0; i < len(arr); i++ {
		if odds <= arr[i] {
			tres.specialgameodds[i]++
			break
		}
	}
}
func (tres *testres) subpool(odds int) {
	gold := tres.tc.bet * odds
	tax := float64(gold) * float64(tres.tc.userwintax) / 10000.0
	tres.pool -= float64(gold) + tax
	tres.pooltax += tax
	tres.totaltax += tax
	tres.totalreturn += gold
}

func (tres *testres) writefile() {
	os.Remove("./result.txt")
	file, _ := os.OpenFile("./result.txt", os.O_RDWR|os.O_CREATE, 0766)
	str := fmt.Sprintf("测试次数：%v\r", tres.tc.testtime)
	file.WriteString(str)

	str = fmt.Sprintf("总测试次数：%v\r", tres.testtime)
	file.WriteString(str)

	str = fmt.Sprintf("当前作弊率：%v\r", tres.tc.cheatvalue)
	file.WriteString(str)

	str = fmt.Sprintf("总下注额度：%v\r", tres.totalbet)
	file.WriteString(str)

	str = fmt.Sprintf("总返奖额度：%v\r", tres.totalreturn)
	file.WriteString(str)

	str = fmt.Sprintf("总税收额度：%v\r", tres.totaltax)
	file.WriteString(str)

	outscore := float64(tres.totalreturn) + tres.pooltax
	str = fmt.Sprintf("总吐分数值：%f\r", outscore)
	file.WriteString(str)

	eatscore := float64(tres.totalbet) - tres.bettax
	str = fmt.Sprintf("总吃分数值：%f\r", eatscore)
	file.WriteString(str)

	str = fmt.Sprintf("吞吐率：%v\r", outscore/eatscore*100.0)
	file.WriteString(str)

	str = fmt.Sprintf("返奖率：%v\r", float64(tres.totalreturn)/float64(tres.totalbet)*100.0)
	file.WriteString(str)

	str = fmt.Sprintf("中奖率：%v\r", float64(tres.awardcount)/float64(tres.testtime)*100.0)
	file.WriteString(str)

	str = fmt.Sprintf("彩金触发次数：%v\r", tres.enterjackpottimes)
	file.WriteString(str)

	str = fmt.Sprintf("彩金返奖：%f\r", tres.jackpotgetgold)
	file.WriteString(str)
	//免费游戏统计
	str = fmt.Sprintf("免费游戏返奖金额：%v\r", tres.freegamegetgold)
	file.WriteString(str)

	str = fmt.Sprintf("免费触发次数：%v\r", tres.enterfreegametimescount)
	file.WriteString(str)

	str = fmt.Sprintf("免费平均返奖值：%v\r", float64(tres.freegamegetgold)/float64(tres.enterfreegametimescount))
	file.WriteString(str)

	str = fmt.Sprintf("免费游戏触发概率：%v\r", float64(tres.enterfreegametimescount)/float64(tres.tc.testtime)*100.0)
	file.WriteString(str)

	str = fmt.Sprintf("免费游戏转动次数：%v\r", tres.freegametimes)
	file.WriteString(str)

	str = fmt.Sprintf("免费转动平均返奖值：%v\r", float64(tres.freegamegetgold)/float64(tres.freegametimes))
	file.WriteString(str)

	str = fmt.Sprintf("免费游戏触发5次：%v\r", tres.freegametype[0])
	file.WriteString(str)

	str = fmt.Sprintf("免费游戏触发8次：%v\r", tres.freegametype[1])
	file.WriteString(str)

	str = fmt.Sprintf("免费游戏触发10次：%v\r", tres.freegametype[2])
	file.WriteString(str)

	str = fmt.Sprintf("免费游戏返奖区间0：%v\r", tres.freeodds[0])
	file.WriteString(str)
	arr := [...]int{0, 1, 3, 5, 10, 30, 50, 100, 200, 500, 99999999}
	for i := 0; i < len(arr)-1; i++ {
		str = fmt.Sprintf("免费游戏返奖区间%v~%v：%v\r", arr[i], arr[i+1], tres.freeodds[i+1])
		file.WriteString(str)
	}
	//特殊游戏统计
	str = fmt.Sprintf("特殊游戏返奖金额：%v\r", tres.specialgamegetgold)
	file.WriteString(str)

	str = fmt.Sprintf("特殊触发次数：%v\r", tres.enterspecialgametimes)
	file.WriteString(str)

	str = fmt.Sprintf("特殊游戏触发概率：%v\r", float64(tres.enterspecialgametimes)/float64(tres.tc.testtime)*100.0)
	file.WriteString(str)

	str = fmt.Sprintf("特殊游戏返奖区间0：%v\r", tres.specialgameodds[0])
	file.WriteString(str)
	arr4 := [...]int{0, 1, 3, 5, 10, 30, 50, 100, 200, 500, 99999999}
	for i := 0; i < len(arr4)-1; i++ {
		str = fmt.Sprintf("特殊游戏返奖区间%v~%v：%v\r", arr4[i], arr4[i+1], tres.specialgameodds[i+1])
		file.WriteString(str)
	}

	//小游戏统计
	//str = fmt.Sprintf("小游戏返奖金额：%v\r", tres.littlegamegetgold)
	//file.WriteString(str)
	//
	//str = fmt.Sprintf("小游戏触发次数：%v\r", tres.enterlittlegametimes)
	//file.WriteString(str)
	//
	//str = fmt.Sprintf("小游戏平均返奖值：%v\r", float64(tres.littlegamegetgold)/float64(tres.enterlittlegametimes))
	//file.WriteString(str)
	//
	//str = fmt.Sprintf("小游戏触发概率：%v\r", float64(tres.enterlittlegametimes)/float64(tres.tc.testtime)*100.0)
	//file.WriteString(str)
	//
	//str = fmt.Sprintf("小游戏触发5次：%v\r", tres.littlegametype[0])
	//file.WriteString(str)
	//
	//str = fmt.Sprintf("小游戏触发10次：%v\r", tres.littlegametype[1])
	//file.WriteString(str)
	//
	//str = fmt.Sprintf("小游戏触发15次：%v\r", tres.littlegametype[2])
	//file.WriteString(str)
	//
	//str = fmt.Sprintf("小游戏返奖区间0：%v\r", tres.littlegameodds[0])
	//file.WriteString(str)
	//arr2 := [...]int{0, 1, 3, 5, 10, 30, 50, 100, 200, 500, 99999999}
	//for i := 0; i < len(arr2)-1; i++ {
	//	str = fmt.Sprintf("小游戏返奖区间%v~%v：%v\r", arr2[i], arr2[i+1], tres.littlegameodds[i+1])
	//	file.WriteString(str)
	//}

	str = fmt.Sprintf("常规转动返获取金币：%v\r", tres.normalgetgold)
	file.WriteString(str)

	str = fmt.Sprintf("常规转动返奖区间0：%v\r", tres.odds[0])
	file.WriteString(str)

	arr3 := [...]int{0, 1, 3, 5, 10, 30, 50, 100, 200, 500, 99999999}
	for i := 0; i < len(arr3)-1; i++ {
		str = fmt.Sprintf("常规转动返奖区间%v~%v：%v\r", arr3[i], arr3[i+1], tres.odds[i+1])
		file.WriteString(str)
	}

	str = fmt.Sprintf("使用特殊配置的次数：%v\r", tres.SpCount)
	file.WriteString(str)

	file.Close()
}
