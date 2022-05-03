package gamelogic

import (
	"common/score"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"fmt"
	"game_LaBa/labacom/config"
	"game_LaBa/labacom/iconlogic"
	roomconfig "game_LaBa/yhhwd/config"
	yhhwd "game_LaBa/yhhwd/msg"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	table         table.TableInterface   // table interface
	user          player.PlayerInterface //用户
	lbcfg         *config.LabaConfig     //拉霸配置
	icon          iconlogic.Iconinfo     //图形算法逻辑
	FreeGameTimes int                    //剩余免费游戏次数
	FreeGameGold  int64                  //免费游戏获取到的金币
	LastBet       int64                  //最近一次用户下注的钱
	BetArr        []int32                //下注配置
	Line          int32                  //线数
	Type          int32                  //游戏类型
	UserTotalWin  int64                  //玩家总赢钱，算产出
	CruentIndex   [25]int32              //当前获奖图标下标
	curr          int64
	CruentOdd     int //当前倍数

	AllBet int64 // 总下注

	testMsg    *yhhwd.TestMsg
	WildArr    []int32
	Stage      int         //阶段1 2 3 4 5
	CheatValue string      //个人 系统
	StageTree  [5][5]*Node //5个阶段得起始树

}

//33445
var Matrix1 = [][]int64{
	{-1, 1, 2, 3, -1},
	{-1, 6, 7, 8, -1},
	{-1, 11, 12, 13, 14},
	{-1, 16, 17, 18, 19},
	{20, 21, 22, 23, 24},
}

//34455
var Matrix2 = [][]int64{
	{-1, 1, 2, 3, -1},
	{-1, 6, 7, 8, 9},
	{-1, 11, 12, 13, 14},
	{15, 16, 17, 18, 19},
	{20, 21, 22, 23, 24},
}

//44555
var Matrix3 = [][]int64{
	{-1, 1, 2, 3, 4},
	{-1, 6, 7, 8, 9},
	{10, 11, 12, 13, 14},
	{15, 16, 17, 18, 19},
	{20, 21, 22, 23, 24},
}

//45555
var Matrix4 = [][]int64{
	{-1, 1, 2, 3, 4},
	{5, 6, 7, 8, 9},
	{10, 11, 12, 13, 14},
	{15, 16, 17, 18, 19},
	{20, 21, 22, 23, 24},
}

//5*5
var Matrix5 = [][]int64{
	{0, 1, 2, 3, 4},
	{5, 6, 7, 8, 9},
	{10, 11, 12, 13, 14},
	{15, 16, 17, 18, 19},
	{20, 21, 22, 23, 24},
}

//初始化
func (g *Game) Init(lb *config.LabaConfig) {
	g.lbcfg = lb
	g.FreeGameTimes = 0
	g.Line = int32(lb.LineCount)
	g.CreatTree()
}

//用户押注
func (g *Game) OnUserBet(b []byte) {
	g.Stage = 1
	g.CruentOdd = 0
	data := &yhhwd.UserBet{}
	senddata := new(yhhwd.BetRes)

	proto.Unmarshal(b, data)
	if !g.CheckUserBet(data.BetMoney) {
		return
	}
	g.Type = 0
	tmpfree := g.FreeGameTimes
	cheat := g.GetCheatValue()
	bfree := g.FreeGameTimes > 0
	//免费游戏减一
	usertatolwin := int64(0)
TEST_LABEL:
	sendSingleresdata := new(yhhwd.SingleRes)
	g.ChangeMatrix(g.Stage)
	//log.Traceln("==========",g.Stage,g.lbcfg.Matrix)
	g.GetIconRes(int64(cheat))
	odds := g.icon.Odds
	log.Traceln("当前倍数", odds, g.CruentOdd)

	//if g.testMsg != nil && g.testMsg.Result == 0 && odds < 120 {
	//
	//}
	//免费游戏
	log.Traceln("免费游戏次数==", g.FreeGameTimes, tmpfree)
	if g.FreeGameTimes != tmpfree {
		senddata.BEnterFree = true
	} else {
		senddata.BEnterFree = false
	}

	//if g.FreeGameTimes != tmpfree {
	//	senddata.BEnterFree = true
	//} else {
	//	senddata.BEnterFree = false
	//}
	//
	//if bfree {
	//	g.FreeGameTimes--
	//	g.FreeGameGold += int64(odds) * int64(g.LastBet)
	//} else {
	//	g.FreeGameGold = 0
	//}
	//免费游戏赢取金额统计
	log.Traceln("xiazhu", g.LastBet)
	if bfree {
		//g.FreeGameTimes--
		g.FreeGameGold += int64(odds) * int64(g.LastBet)
	}
	//单次倍数
	sendSingleresdata.Odds = int32(odds)
	sendSingleresdata.Gold = int64(odds) * int64(g.LastBet)

	//senddata.Cheat = int32(cheat)
	//senddata.BloodPool = roomconfig.TestRoomConfig.Pool
	//玩家总赢钱
	usertatolwin += sendSingleresdata.Gold
	//g.user.SetScore(g.table.GetGameNum(), sendSingleresdata.Gold, g.table.GetRoomRate())
	//g.UserTotalWin +=sendSingleresdata.Gold
	//senddata.UserGold = g.user.GetScore()
	//senddata.FreeGames = int32(g.FreeGameTimes)
	//图型
	sendSingleresdata.IconArr = append(sendSingleresdata.IconArr, g.icon.Iconarr...)
	//线型
	for k, v := range g.CruentIndex {
		if v != -1 {
			sendSingleresdata.Line = append(sendSingleresdata.Line, int32(k))
		}
	}
	log.Traceln("下标====", g.CruentIndex, sendSingleresdata.Line, sendSingleresdata.IconArr)
	g.dealWithTest(senddata)
	//所有结果
	senddata.AllResArr = append(senddata.AllResArr, sendSingleresdata)
	//log.Traceln("发送数据",senddata,g.icon.Iconarr)
	//如果中奖继续出将 中将得情况下不出免费游戏，免费游戏只在第一次出

	if g.FreeGameTimes > 0 && senddata.BEnterFree {

	} else {
		if g.Stage > 5 {
			if len(sendSingleresdata.Line) == 25 {

			} else if odds > g.CruentOdd {
				g.CruentOdd = odds
				goto TEST_LABEL
			}
		} else if g.Stage <= 5 {
			if len(sendSingleresdata.Line) == 25 {

			} else if odds > 0 {
				g.CruentOdd = odds
				g.Stage++
				goto TEST_LABEL
			}
		}
	}
	if bfree {
		g.FreeGameTimes--

	} else {
		g.FreeGameGold = 0
	}
	//上下分 免费游戏次数 玩家金额
	g.table.WriteLogs(g.user.GetId(), fmt.Sprintln("结算前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), usertatolwin, g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetId(), fmt.Sprintln("结算前:", score.GetScoreStr(g.user.GetScore())))
	g.UserTotalWin += usertatolwin
	senddata.UserTotalWin = usertatolwin
	senddata.UserGold = g.user.GetScore()
	senddata.FreeGames = int32(g.FreeGameTimes)
	senddata.Count = int32(len(senddata.AllResArr))
	g.user.SendMsg(int32(yhhwd.ReMsgIDS2C_BetRet), senddata)
	//g.dealWithTest(senddata)
	BetGold := data.BetMoney * g.Line
	if bfree {
		BetGold = 0
	}
	str := fmt.Sprint(time.Now().Format("2006-1-2 15:04:05"), "用户ID：", g.user.GetID(), g.CheatValue, "作弊率：", cheat, " 结果数组：", senddata.AllResArr, "，扣钱：", BetGold/100, ".",
		BetGold%100, "，加钱：", senddata.UserTotalWin/100, ".", senddata.UserTotalWin%100, "，免费次数：", senddata.FreeGames)
	log.Debugf("%v", str)
	// arrstr := fmt.Sprint(senddata.IconArr)
	//g.user.SetEndCards(arrstr)
	g.table.WriteLogs(g.user.GetId(), str)
	g.PaoMaDeng(senddata.UserTotalWin)
	g.user.SendChip(int64(BetGold))
	g.testMsg = nil
}

func (g *Game) GetIconRes(cheatvalue int64) {
	if g.FreeGameTimes == 0 {
		g.icon.GetIcon(cheatvalue, g.lbcfg, g.FreeGameTimes > 0, false, g.GetNewOdds, g.ChangeIconRet)
	} else {
		g.icon.GetIcon(cheatvalue, g.lbcfg, g.FreeGameTimes > 0, false, g.GetNewOdds, g.ChangeIconRet)
	}

	/*
		//测试用
		g.icon.Iconarr = make([]int32, 0)
		tmp := [...]int32{1, 10, 8, 5, 8, 10, 9, 10, 1, 2, 1, 2, 6, 7, 0}
		for _, v := range tmp {
			g.icon.Iconarr = append(g.icon.Iconarr, v)
		}
	*/
	if g.Stage == 1 {
		count := g.icon.Getfreegametimes(g.lbcfg)
		log.Traceln("免费游戏", count)
		for _, v := range count {
			g.FreeGameTimes += int(g.lbcfg.FreeGame.Times[v])

		}
	}

}

func (g *Game) GetIconOdds() int {
	return g.icon.Odds
}

func (g *Game) CheckUserBet(BetMoney int32) bool {
	if g.FreeGameTimes > 0 {
		g.Stage = 3
		return true
	}
	//判断客户端下注金币是否和筹码配置一样。如果不一样下注失败
	temp := false
	for _, v := range g.BetArr {
		if BetMoney == v {
			temp = true
			break
		}
	}
	if !temp {
		msg := new(yhhwd.BetFail)
		msg.FailID = 2
		msg.Reson = "数据异常"
		g.user.SendMsg(int32(yhhwd.ReMsgIDS2C_BetFailID), msg)
		return false
	}
	//这里检查用户的钱是否足够
	if g.user.GetScore() < int64(BetMoney) {
		msg := new(yhhwd.BetFail)
		msg.FailID = 1
		msg.Reson = "金币不足，请充值！"

		g.user.SendMsg(int32(yhhwd.ReMsgIDS2C_BetFailID), msg)
		log.Tracef("金币不足，请充值！")
		return false
	}

	//这里检查用户输入参数是否有问题，是否在下注的范围内
	//这里扣钱
	g.table.WriteLogs(g.user.GetId(), fmt.Sprintln("结算前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), int64(-BetMoney*g.Line), g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetId(), fmt.Sprintln("结算前:", score.GetScoreStr(g.user.GetScore())))
	if int64(BetMoney) != g.LastBet {
		g.LastBet = int64(BetMoney)
	}
	g.AllBet += int64(BetMoney)

	return true
}

//获取奖金池游戏奖励
func (g *Game) GetJackpotMoney(Jackpot int64) float64 {
	gold := float64(Jackpot) * float64(g.icon.Getjackpot(g.lbcfg)) / 10000.0
	return gold
}

func (g *Game) GetRoomconfig() {
	g.BetArr = make([]int32, 0)
	index := g.table.GetLevel()
	for i := 0; i < len(g.lbcfg.BetConfig[index-1]); i++ {
		g.BetArr = append(g.BetArr, int32(g.lbcfg.BetConfig[index-1][i]))
	}
}

func (g *Game) GetRebackInfo() {
	data := g.user.GetTableData()
	if len(data) != 0 {
		arrstr := strings.Split(data, ",")

		g.FreeGameTimes, _ = strconv.Atoi(arrstr[0])
		g.FreeGameGold, _ = strconv.ParseInt(arrstr[1], 10, 0)

		g.LastBet, _ = strconv.ParseInt(arrstr[2], 10, 0)
		g.user.DelTableData()
	}
}

func (g *Game) GetCheatValue() int {
	//先获取用户的
	Prob := g.user.GetProb()
	g.CheatValue = "点控"
	if Prob == 0 {
		tmp, _ := g.table.GetRoomProb()
		g.CheatValue = "系统"
		Prob = int32(tmp)
		if tmp == 0 {
			g.CheatValue += " 获取到系统作弊率为0 "
			Prob = 1000
		}
	}

	return int(Prob)
}

func (g *Game) GetIconCount() int {
	return g.icon.Count
}

//替换wild图标
func (g *Game) ChangeIconRet(cheatvalue int64) {
	log.Traceln("阶段", g.Stage)
	if g.FreeGameTimes > 0 {
		if g.Stage <= 3 {
			log.Traceln("不替换")
			return
		}

	}
	if g.Stage > 1 && g.Stage <= 5 {
		log.Traceln("-----jie", g.Stage)
		for i := 5; i <= 24; i++ {
			if g.CruentIndex[i] != -1 {
				g.icon.Iconarr[i-5] = g.CruentIndex[i]
			}
		}
	} else if g.Stage > 5 {
		log.Traceln("-----jie5", g.Stage)
		for i := 0; i <= 24; i++ {
			if g.CruentIndex[i] != -1 {
				g.icon.Iconarr[i] = g.CruentIndex[i]
			}
		}
	}

	g.Type = 1
}

func (g *Game) NormalChangeIconRet(cheatvalue int64) {
	g.WildArr = make([]int32, 0)
	pro := roomconfig.CSDConfig.GetProByCheat(int(cheatvalue))
	r := rand.Intn(10000)
	if r < pro {
		for i := 7; i < 11; i++ {
			g.icon.Iconarr[i] = int32(g.lbcfg.Wild.IconId)
		}
		g.rangeThreeWild()
		g.Type = 1
	}
}

func (g *Game) PaoMaDeng(Gold int64) {
	configs := g.table.GetMarqueeConfig()
	log.Debugf("跑马灯配置%v", configs)
	for _, v := range configs {
		special, _ := strconv.Atoi(v.SpecialCondition)
		if g.Type == 1 && special == 1 && Gold >= v.AmountLimit {
			err := g.table.CreateMarquee(g.user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Debugf("创建跑马灯错误：%v", err)
			}
		} else if Gold >= v.AmountLimit && len(v.SpecialCondition) == 0 {
			err := g.table.CreateMarquee(g.user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Debugf("创建跑马灯错误：%v", err)
			}
		}
	}
}

func (game *Game) dealWithTest(br *yhhwd.BetRes) {
	log.Traceln("测试")
	if game.testMsg == nil {
		return
	}
	switch game.testMsg.Result {
	case 0:
	case 1:
		br.BEnterFree = true
		br.FreeGames = 5
		game.FreeGameTimes = 5
	}
}

func (game *Game) handleTest(bts []byte) {
	msg := new(yhhwd.TestMsg)
	if err := proto.Unmarshal(bts, msg); err != nil {
		return
	}
	log.Traceln("测试", msg)
	switch msg.Result {
	case 0, 1:
	default:
		return
	}
	game.testMsg = msg
}

//随机替换3个图标wild
func (g *Game) rangeThreeWild() {
	totalweight := 0
	//记录替换的3个图标位置
	wild1 := -1
	wild2 := -1
	wild3 := -1
	count := 0
	for _, v := range roomconfig.CSDConfig.WildCheatPro {
		totalweight += v
	}
	for {
		r := rand.Intn(totalweight)
		for k, v := range roomconfig.CSDConfig.WildCheatPro {

			if k == wild1 || k == wild2 || k == wild3 {
				continue
			}
			//免费游戏的图标不替换
			if g.icon.Iconarr[k] == int32(g.lbcfg.FreeGame.IconId) {
				continue
			}
			if r < v {
				count++
				switch count {
				case 1:
					wild1 = k
				case 2:
					wild2 = k
				case 3:
					wild3 = k
				}
				g.icon.Iconarr[k] = int32(g.lbcfg.Wild.IconId)
				g.WildArr = append(g.WildArr, int32(k))
				if count == 3 {
					return
				}
			}
			r -= v
		}
	}
}

func (g *Game) isLongMu() bool {
	//判断是否是龙母图标
	for i := 7; i < 11; i++ {
		if g.icon.Iconarr[i] == int32(g.lbcfg.Wild.IconId) {
		} else {
			return false
		}
	}
	return true
}

//出龙母概率
func (g *Game) TestChangeIconRet(cheatvalue int64) {
	g.WildArr = make([]int32, 0)

	for i := 7; i < 11; i++ {
		g.icon.Iconarr[i] = int32(g.lbcfg.Wild.IconId)
	}
	g.rangeThreeWild()
	g.Type = 1
}

func (g *Game) ChangeMatrix(MatrixId int) {
	if MatrixId >= 5 {
		g.lbcfg.Matrix = Matrix5
		return
	}
	switch MatrixId {
	case 1:
		//图形33445
		g.lbcfg.Matrix = Matrix1
	case 2:
		//图形34455
		g.lbcfg.Matrix = Matrix2
	case 3:
		//图形44555
		g.lbcfg.Matrix = Matrix3
	case 4:
		//图形45555
		g.lbcfg.Matrix = Matrix4

	}
}

var MatrixArr1 = []int32{
	-1, 1, 2, 3, -1,
	-1, 6, 7, 8, -1,
	-1, 11, 12, 13, 14,
	-1, 16, 17, 18, 19,
	20, 21, 22, 23, 24,
}

//34455
var MatrixArr2 = []int32{
	-1, 1, 2, 3, -1,
	-1, 6, 7, 8, 9,
	-1, 11, 12, 13, 14,
	15, 16, 17, 18, 19,
	20, 21, 22, 23, 24,
}

//44555
var MatrixArr3 = []int32{
	-1, 1, 2, 3, 4,
	-1, 6, 7, 8, 9,
	10, 11, 12, 13, 14,
	15, 16, 17, 18, 19,
	20, 21, 22, 23, 24,
}

//45555
var MatrixArr4 = []int32{
	-1, 1, 2, 3, 4,
	5, 6, 7, 8, 9,
	10, 11, 12, 13, 14,
	15, 16, 17, 18, 19,
	20, 21, 22, 23, 24,
}

//5*5
var MatrixArr5 = []int32{
	0, 1, 2, 3, 4,
	5, 6, 7, 8, 9,
	10, 11, 12, 13, 14,
	15, 16, 17, 18, 19,
	20, 21, 22, 23, 24,
}

func getMatrixArr5(i int) []int32 {
	switch i {
	case 0:
		return MatrixArr1
	case 1:
		return MatrixArr2
	case 2:
		return MatrixArr3
	case 3:
		return MatrixArr4
	case 4:
		return MatrixArr5

	}
	return nil
}
func (g *Game) CreatTree() {
	//var StageTree [5][5]*Node
	//var TreePath  [][][]int
	var result []*Node //当前查找结果值
	var temp []*Node   //中间值
	var Matrix []int32
	//从0-4 开始查找
	for q := 0; q <= 4; q++ {
		Matrix = getMatrixArr5(q)
		b := 0
		for j := 0; j <= 4; j++ {
			//值为-1跳过
			if Matrix[j] == -1 {
				continue
			}
			g.StageTree[q][j] = g.StageTree[q][j].Insert(j, 1)
			//新循当局结果初始化
			result = make([]*Node, 0)
			result = append(result, g.StageTree[q][j])
			b++
			//Allresult=append(Allresult,t)
			//查找节点值
			//遍历后面4列数据

			for i := 0; i < 4; i++ {
				var first, fi, fj, last, li, lj, mi, mj int
				//每列循环 头尾 中间数字找的相邻位置，5个阶段每个阶段相邻的不一样
				if q == 0 {
					switch i {
					case 0:
						first = 1
						fi = 5
						fj = 6
						last = 3
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 1:
						first = 6
						fi = 5
						fj = 6
						last = 8
						li = 5
						lj = 6
						mi = 5
						mj = 6
					case 2:
						first = 11
						fi = 5
						fj = 6
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 3:
						first = 16
						fi = 4
						fj = 5
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 5
					}
				} else if q == 1 {
					switch i {
					case 0:
						first = 1
						fi = 5
						fj = 6
						last = 3
						li = 5
						lj = 6
						mi = 5
						mj = 6
					case 1:
						first = 6
						fi = 5
						fj = 6
						last = 9
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 2:
						first = 11
						fi = 4
						fj = 5
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 5
					case 3:
						first = 15
						fi = 5
						fj = 6
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 6
					}

				} else if q == 2 {
					switch i {
					case 0:
						first = 1
						fi = 5
						fj = 6
						last = 4
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 1:
						first = 6
						fi = 4
						fj = 5
						last = 9
						li = 4
						lj = 5
						mi = 4
						mj = 5
					case 2:
						first = 10
						fi = 5
						fj = 6
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 3:
						first = 15
						fi = 5
						fj = 6
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 6
					}
				} else if q == 3 {
					switch i {
					case 0:
						first = 1
						fi = 4
						fj = 5
						last = 4
						li = 4
						lj = 5
						mi = 4
						mj = 5
					case 1:
						first = 5
						fi = 5
						fj = 6
						last = 9
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 2:
						first = 10
						fi = 5
						fj = 6
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 3:
						first = 15
						fi = 5
						fj = 6
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 6
					}
				} else if q == 4 {
					switch i {
					case 0:
						first = 0
						fi = 5
						fj = 6
						last = 4
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 1:
						first = 5
						fi = 5
						fj = 6
						last = 9
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 2:
						first = 10
						fi = 5
						fj = 6
						last = 14
						li = 4
						lj = 5
						mi = 4
						mj = 6
					case 3:
						first = 15
						fi = 5
						fj = 6
						last = 19
						li = 4
						lj = 5
						mi = 4
						mj = 6
					}
				}
				//上局结果赋值给中间变量
				temp = result
				//初始化当局结果
				result = make([]*Node, 0)
				//上局是否有结果 if有则继续下一列查找 else 退出本次查找
				//遍历上局结果，用每个结果便利下一下列查找是否有相邻的
				for _, v := range temp {
					index := 0 //赋值到树的下表，012 对应Left mid right
					var tempi, tempj int
					if v.value == first {
						tempi = fi
						tempj = fj
					} else if v.value == last {
						tempi = li
						tempj = lj
					} else {
						tempi = mi
						tempj = mj
					}
					//每个值需要查找相邻的下标循环
					for k := tempi; k <= tempj; k++ {
						tr := v
						switch index {
						case 0:
							tr.left = CreateNode(v.value+k, i+2)
							result = append(result, tr.left)
						case 1:
							tr.mid = CreateNode(v.value+k, i+2)
							result = append(result, tr.mid)
						case 2:
							tr.right = CreateNode(v.value+k, i+2)
							result = append(result, tr.right)
						}
						index++
					}
				}
			}
		}
	}

	return

}
func (g *Game) GetNewOdds(cheat int64) {
	log.Traceln(g.StageTree)
	g.icon.Odds = 0
	g.CurrenticonidInit()
	stage := g.Stage
	if g.Stage >= 5 {
		stage = 5
	}

	for i := 1; i <= 5; i++ {
		if stage == i {
			for _, v := range g.StageTree[i-1] {
				if v == nil {
					continue
				}
				log.Traceln("=====jieguo=====", v.value)
				var temp []int
				g.OneTreeOdds(v, g.icon.Iconarr[v.value], temp)
			}

		}
	}

}

//获取
func (g *Game) OneTreeOdds(n *Node, iconid int32, indexarr []int) {

	currenid := iconid
	if iconid == int32(g.lbcfg.Wild.IconId) {
		currenid = g.icon.Iconarr[n.value]
	}
	bEnter := true
	indexarr = append(indexarr, n.value)

	if n.left != nil {
		if g.icon.Iconarr[n.left.value] == int32(g.lbcfg.Wild.IconId) || g.icon.Iconarr[n.left.value] == currenid || currenid == int32(g.lbcfg.Wild.IconId) {
			var tempindexarr []int
			tempindexarr = append(tempindexarr, indexarr...)
			g.OneTreeOdds(n.left, currenid, tempindexarr)
			bEnter = false
		}
	}
	if n.mid != nil {
		if g.icon.Iconarr[n.mid.value] == int32(g.lbcfg.Wild.IconId) || g.icon.Iconarr[n.mid.value] == currenid || currenid == int32(g.lbcfg.Wild.IconId) {
			var tempindexarr []int
			tempindexarr = append(tempindexarr, indexarr...)
			g.OneTreeOdds(n.mid, currenid, tempindexarr)
			bEnter = false
		}

	}
	if n.right != nil {
		if g.icon.Iconarr[n.right.value] == int32(g.lbcfg.Wild.IconId) || g.icon.Iconarr[n.right.value] == currenid || currenid == int32(g.lbcfg.Wild.IconId) {
			var tempindexarr []int
			tempindexarr = append(tempindexarr, indexarr...)
			g.OneTreeOdds(n.right, currenid, tempindexarr)
			bEnter = false
		}

	}
	if bEnter {
		//log.Traceln("=====",n.hight,n.value,currenid)
		//log.Traceln()
		if n.hight > 2 {
			//log.Traceln("中奖下标路劲",indexarr,"值",n.value,"高度",n.hight,currenid)
			for _, v := range indexarr {
				g.CruentIndex[v] = g.icon.Iconarr[v]
			}

			g.icon.Odds += int(g.lbcfg.IconAward[int64(currenid)][int64(n.hight-1)])
		}
		return
	}

}

func (g *Game) CurrenticonidInit() {
	for i := 0; i <= 24; i++ {
		g.CruentIndex[i] = -1
	}
}
