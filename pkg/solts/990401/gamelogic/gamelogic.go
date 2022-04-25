package gamelogic

import (
	"fmt"
	roomconfig "go-game-sdk/example/game_LaBa/990401/config"
	csd "go-game-sdk/example/game_LaBa/990401/msg"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"go-game-sdk/example/game_LaBa/labacom/iconlogic"
	"go-game-sdk/inter"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/golang/protobuf/proto"
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

	curr int64

	AllBet int64 // 总下注

	testMsg    *csd.TestMsg
	CheatValue string //个人 系统
}

//初始化
func (g *Game) Init(lb *config.LabaConfig) {
	g.lbcfg = lb
	g.FreeGameTimes = 0
	g.Line = int32(lb.LineCount)
}

func (lbr *Game) BindRobot(ai inter.AIUserInter) player.RobotHandler {
	return nil
}

//用户押注
func (g *Game) OnUserBet(b []byte) {
	data := &csd.UserBet{}
	senddata := new(csd.BetRes)
	proto.Unmarshal(b, data)
	if !g.CheckUserBet(data.BetMoney) {
		return
	}

	g.Type = 0
	tmpfree := g.FreeGameTimes
	cheat := g.GetCheatValue()
	bfree := g.FreeGameTimes > 0
	//免费游戏减一
TEST_LABEL:
	g.GetIconRes(int64(cheat))
	odds := g.icon.Gettotalodds(g.lbcfg)
	if g.testMsg != nil && g.testMsg.Result == 0 && odds < 120 {
		goto TEST_LABEL
	}

	if g.FreeGameTimes != tmpfree {
		senddata.BEnterFree = true
	} else {
		senddata.BEnterFree = false
	}

	if bfree {
		g.FreeGameTimes--
		g.FreeGameGold += int64(odds) * int64(g.LastBet)
	} else {
		g.FreeGameGold = 0
	}

	senddata.Odds = int32(odds)
	senddata.Gold = int64(odds) * int64(g.LastBet)
	//senddata.Cheat = int32(cheat)
	//senddata.BloodPool = roomconfig.TestRoomConfig.Pool
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("结算前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), senddata.Gold, g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("结算后:", score.GetScoreStr(g.user.GetScore())))
	g.UserTotalWin += senddata.Gold
	senddata.UserGold = g.user.GetScore()
	senddata.FreeGames = int32(g.FreeGameTimes)
	senddata.IconArr = append(senddata.IconArr, g.icon.Iconarr...)
	for _, v := range g.icon.Line {
		li := new(csd.LineInfo)
		li.Count = v.Count
		li.Gold = v.Gold * int64(g.LastBet)
		li.Index = v.Index
		senddata.Line = append(senddata.Line, li)
	}
	g.dealWithTest(senddata)
	g.user.SendMsg(int32(csd.ReMsgIDS2C_BetRet), senddata)

	BetGold := data.BetMoney * g.Line
	if bfree {
		BetGold = 0
	}
	str := fmt.Sprint(time.Now().Format("2006-1-2 15:04:05"), " 用户ID：", g.user.GetID(), g.CheatValue, "作弊率：", cheat, " 结果数组：", senddata.IconArr, "，扣钱：", BetGold/100, ".",
		BetGold%100, "，加钱：", senddata.Gold/100, ".", senddata.Gold%100, "，免费次数：", senddata.FreeGames)
	log.Debugf("%v", str)
	// arrstr := fmt.Sprint(senddata.IconArr)
	// g.user.SetEndCards(arrstr)
	g.table.WriteLogs(g.user.GetID(), str)
	g.PaoMaDeng(senddata.Gold)
	g.user.SendChip(int64(BetGold))
	g.testMsg = nil
}

func (g *Game) GetIconRes(cheatvalue int64) {

	if g.FreeGameTimes == 0 {
		g.icon.GetIcon(cheatvalue, g.lbcfg, g.FreeGameTimes > 0, false, nil, g.NormalChangeIconRet)
	} else {
		g.icon.GetIcon(cheatvalue, g.lbcfg, g.FreeGameTimes > 0, false, nil, g.ChangeIconRet)
	}

	/*
		//测试用
		g.icon.Iconarr = make([]int32, 0)
		tmp := [...]int32{1, 10, 8, 5, 8, 10, 9, 10, 1, 2, 1, 2, 6, 7, 0}
		for _, v := range tmp {
			g.icon.Iconarr = append(g.icon.Iconarr, v)
		}
	*/
	count := g.icon.Getfreegametimes(g.lbcfg)
	for _, v := range count {
		g.FreeGameTimes += int(g.lbcfg.FreeGame.Times[v])
	}
}

func (g *Game) GetIconOdds() int {
	return g.icon.Odds
}

func (g *Game) CheckUserBet(BetMoney int32) bool {
	if g.FreeGameTimes > 0 {
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
		msg := new(csd.BetFail)
		msg.FailID = 2
		msg.Reson = "数据异常"
		g.user.SendMsg(int32(csd.ReMsgIDS2C_BetFailID), msg)
		return false
	}
	//这里检查用户的钱是否足够
	if g.user.GetScore() < int64(BetMoney) {
		msg := new(csd.BetFail)
		msg.FailID = 1
		msg.Reson = "金币不足，请充值！"

		g.user.SendMsg(int32(csd.ReMsgIDS2C_BetFailID), msg)
		log.Tracef("金币不足，请充值！")
		return false
	}

	//这里检查用户输入参数是否有问题，是否在下注的范围内
	//这里扣钱
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("下注前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), int64(-BetMoney*g.Line), g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("下注后:", score.GetScoreStr(g.user.GetScore())))
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
		tmp := g.table.GetRoomProb()
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
	arr := [3]int{6, 7, 8}
	for i := 0; i < 3; i++ {
		g.icon.Iconarr[arr[i]] = int32(g.lbcfg.Wild.IconId)
	}
	g.Type = 1
}

func (g *Game) NormalChangeIconRet(cheatvalue int64) {
	pro := roomconfig.CSDConfig.GetProByCheat(int(cheatvalue))
	r := rand.Intn(10000)
	if r < pro {
		for i := 0; i < 3; i++ {
			g.icon.Iconarr[i] = int32(g.lbcfg.Wild.IconId)
		}

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

func (game *Game) dealWithTest(br *csd.BetRes) {
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
	msg := new(csd.TestMsg)
	if err := proto.Unmarshal(bts, msg); err != nil {
		return
	}
	switch msg.Result {
	case 0, 1:
	default:
		return
	}
	game.testMsg = msg
}
