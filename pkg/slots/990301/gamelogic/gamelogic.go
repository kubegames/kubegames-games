package gamelogic

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/config"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/iconlogic"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	roomconfig "github.com/kubegames/kubegames-games/pkg/slots/990301/config"
	wlzb "github.com/kubegames/kubegames-games/pkg/slots/990301/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	table             table.TableInterface   // table interface
	user              player.PlayerInterface //用户
	lbcfg             *config.LabaConfig     //拉霸配置
	icon              iconlogic.Iconinfo     //图形算法逻辑
	EnterFreeGame     bool                   //是否进入免费
	FreeGameTimes     int                    //剩余免费游戏次数
	FreeGameGold      int64                  //免费游戏获取到的金币
	LastBet           int64                  //最近一次用户下注的钱
	BetArr            []int32                //下注配置
	Line              int32                  //线数
	FreeGameIndex     int32                  //选择的免费游戏的下标
	FreeGameExtraOdds int32                  //免费游戏额外乘以的倍数
	UserTotalWin      int64                  //玩家总赢钱，算产出

	curr int64

	AllBet int64 // 总下注

	test       *wlzb.TestMsg // 测试消息
	CheatValue string        //个人 系统
}

//初始化
func (g *Game) Init(lb *config.LabaConfig) {
	g.lbcfg = lb
	g.FreeGameTimes = 0
	g.EnterFreeGame = false
	g.FreeGameExtraOdds = 0
	g.Line = int32(lb.LineCount)
}

func (lbr *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	return nil
}

//用户押注
func (g *Game) OnUserBet(b []byte) {
	data := &wlzb.UserBet{}
	senddata := new(wlzb.BetRes)
	proto.Unmarshal(b, data)

	if !g.CheckUserBet(data.BetMoney) {
		return
	}
	cheat := g.GetCheatValue()
	bfree := g.FreeGameTimes > 0

	g.GetIconRes(int64(cheat))
	//免费游戏减一
	// busstype := int32(201701)
	if bfree {
		// busstype = 201702
		g.FreeGameTimes--
	}

	odds := g.icon.Odds
	g.settleWithTest()
	senddata.BEnterFree = g.EnterFreeGame
	senddata.Odds = int32(odds)
	senddata.Gold = int64(odds) * int64(g.LastBet)
	//senddata.Cheat = int32(cheat)
	senddata.FreeGameExtraOdd = g.FreeGameExtraOdds

	g.FreeGameExtraOdds = 0
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("结算前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), senddata.Gold, g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("结算后:", score.GetScoreStr(g.user.GetScore())))
	g.UserTotalWin += senddata.Gold
	senddata.UserGold = g.user.GetScore()

	senddata.IconArr = append(senddata.IconArr, g.icon.Iconarr...)
	for _, v := range g.icon.Line {
		li := new(wlzb.LineInfo)
		li.Count = v.Count
		li.Gold = v.Gold * int64(g.LastBet)
		li.Index = v.Index
		senddata.Line = append(senddata.Line, li)
	}

	senddata.FreeGames = int32(g.FreeGameTimes)
	//log.Debugf("%v", senddata)
	g.user.SendMsg(int32(wlzb.ReMsgIDS2C_BetRet), senddata)

	BetGold := data.BetMoney * g.Line
	if bfree {
		BetGold = 0
	}
	str := fmt.Sprint(time.Now().Format("2006-1-2 15:04:05"), " 用户ID：", g.user.GetID(), g.CheatValue, "作弊率：", cheat, " 结果数组：", senddata.IconArr, "，扣钱：", BetGold/100, ".", BetGold%100,
		"，加钱：", senddata.Gold/100, ".", senddata.Gold%100, "，免费次数：", g.FreeGameTimes)
	// arrstr := fmt.Sprint(senddata.IconArr)
	// g.user.SetEndCards(arrstr)
	g.table.WriteLogs(g.user.GetID(), str)
	g.PaoMaDeng(senddata.Gold, bfree)
	g.user.SendChip(int64(BetGold))
	g.test = nil
}

func (g *Game) GetIconRes(cheatvalue int64) {
	if g.FreeGameTimes > 0 {
	L1:
		g.icon.GetIcon(cheatvalue, g.lbcfg, true, false, g.HasWildOnLine, nil)
		if g.test != nil && g.test.ResultType == 0 && g.icon.Odds < 120 {
			goto L1
		}
	} else {
	L2:
		g.icon.GetIcon(cheatvalue, g.lbcfg, false, false, nil, nil)
		if g.test != nil && g.test.ResultType == 0 && g.icon.Odds < 120 {
			goto L2
		}
	}
	/*g.icon.Iconarr = make([]int32, 0)
	tmp := [...]int32{4, 8, 5, 10, 2, 12, 12, 9, 4, 10, 12, 9, 6, 12, 0}
	for _, v := range tmp {
		g.icon.Iconarr = append(g.icon.Iconarr, v)
	}

	odds := g.icon.Gettotalodds(g.lbcfg)

	log.Tracef("倍数为%v", odds)*/

	count := g.icon.GetLineIconCount(g.lbcfg.FreeGame.IconId, g.lbcfg)

	if count >= 3 {
		g.EnterFreeGame = true
	}
}

func (g *Game) GetIconOdds() int {
	return g.icon.Odds
}

//场景消息
func (g *Game) SendSence() {
	senddata := new(wlzb.Sence)
	senddata.BetValue = append(senddata.BetValue, g.BetArr...)
	senddata.Gold = 0

	g.user.SendMsg(int32(wlzb.ReMsgIDS2C_SenceID), senddata)
}

func (g *Game) CheckUserBet(BetMoney int32) bool {
	//这里检查用户的钱是否足够
	if g.FreeGameTimes > 0 {
		log.Tracef("免费游戏：%v", g.FreeGameTimes)
		return true
	}

	if BetMoney <= 0 {
		return false
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
		msg := new(wlzb.BetFail)
		msg.FailID = 2
		msg.Reson = "数据异常"
		g.user.SendMsg(int32(wlzb.ReMsgIDS2C_BetFailID), msg)
		return false
	}
	if g.user.GetScore() < int64(BetMoney) {
		msg := new(wlzb.BetFail)
		msg.FailID = 1
		msg.Reson = "金币不足，请充值！"

		g.user.SendMsg(int32(wlzb.ReMsgIDS2C_BetFailID), msg)
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

//用户选择免费游戏
func (g *Game) ChoseFreeGameTimes(buffer []byte) {
	if !g.EnterFreeGame {
		return
	}

	msg := &wlzb.FreeGames{}
	proto.Unmarshal(buffer, msg)
	g.FreeGameIndex = msg.FreeGamesIndex % 5

	msg.FreeGamesIndex += 1
	cheat := roomconfig.RoomCfg.GetFreeGameWeight(int32(g.GetCheatValue()))
	g.FreeGameTimes += cheat.Fgw[g.FreeGameIndex].FreeGameTimes

	g.user.SendMsg(int32(wlzb.ReMsgIDS2C_ChoseFreeGameRet), msg)
	g.EnterFreeGame = false
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
		g.EnterFreeGame, _ = strconv.ParseBool(arrstr[2])
		tmp, _ := strconv.Atoi(arrstr[3])
		g.FreeGameIndex = int32(tmp)
		g.LastBet, _ = strconv.ParseInt(arrstr[4], 10, 0)

		g.user.DelTableData()
	}
}

func (g *Game) GetCheatValue() int {
	//先获取用户的
	if g.user == nil {
		return 0
	}

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

	//return -3000
	return int(Prob)
}

func (g *Game) HasWildOnLine(cheatvalue int64) {
	if g.icon.Odds == 0 {
		return
	}

	for _, v := range g.icon.Line {
		if v.WildCount > 0 {
			log.Tracef("有wild")
			g.icon.Odds *= g.GetFreeGameExtraOdds(cheatvalue)
			break
		}
	}
}

func (g *Game) GetFreeGameExtraOdds(cheatvalue int64) int {
	fgw := roomconfig.RoomCfg.GetFreeGameWeight(int32(cheatvalue))
	r := rand.Intn(fgw.Fgw[g.FreeGameIndex].FreeGameTimes_TotalWeight)
	for i := 0; i < 3; i++ {
		if r < fgw.Fgw[g.FreeGameIndex].FreeGameTimes_Weight[i] {
			g.FreeGameExtraOdds = int32(fgw.Fgw[g.FreeGameIndex].ExtraOdds[i])
			return fgw.Fgw[g.FreeGameIndex].ExtraOdds[i]
		}

		r -= fgw.Fgw[g.FreeGameIndex].FreeGameTimes_Weight[i]
	}

	return 1
}

func (g *Game) PaoMaDeng(Gold int64, bfree bool) {
	configs := g.table.GetMarqueeConfig()
	log.Debugf("跑马灯配置%v", configs)
	for _, v := range configs {
		special, _ := strconv.Atoi(v.SpecialCondition)
		if bfree && special == 1 && Gold >= v.AmountLimit {
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

func (g *Game) settleWithTest() {
	if g.test == nil {
		return
	}
	switch g.test.ResultType {
	default:
	case 1:
		if !g.EnterFreeGame {
			g.EnterFreeGame = true
		}
	}
}

func (g *Game) handleTestMsg(bts []byte) {
	msg := new(wlzb.TestMsg)
	if err := proto.Unmarshal(bts, msg); err != nil {
		return
	}
	switch msg.ResultType {
	case 0, 1:
	default:
		return
	}
	g.test = msg
}
