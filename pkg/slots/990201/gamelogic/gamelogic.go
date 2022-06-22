package gamelogic

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/config"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/iconlogic"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/xiaomali"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/slots/990201/bibei"
	shz "github.com/kubegames/kubegames-games/pkg/slots/990201/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	table            table.TableInterface   // table interface
	user             player.PlayerInterface //用户
	lbcfg            *config.LabaConfig     //拉霸配置
	xml              *xiaomali.XiaoMaLiCfg  //小玛丽配置
	bbc              *bibei.BiBeiConf       //比倍配置
	icon             iconlogic.Iconinfo     //图形算法逻辑
	XiaoMaLiTimes    int                    //小玛丽次数
	LastBet          int64                  //最近一次用户下注的钱
	LastWin          int64                  //最后一次赢的钱
	SmallGameGetOdds int                    //小游戏本次中奖倍数之和
	XiaoMaLiGold     int64                  //小玛丽获取金币
	BetArr           []int32
	Line             int32    //线数
	Fi               FullIcon //全图
	UserTotalWin     int64    //玩家总赢钱，算产出
	curr             int64

	AllBet int64 // 所有下注

	testMsg    *shz.TestMsg
	CheatValue string //个人 系统
}

//初始化
func (g *Game) Init(lb *config.LabaConfig, xml *xiaomali.XiaoMaLiCfg, bbc *bibei.BiBeiConf) {
	g.lbcfg = lb
	g.xml = xml
	g.bbc = bbc
	g.SmallGameGetOdds = 0
	g.Line = int32(lb.LineCount)
}

func (lbr *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	return nil
}

//用户押注
func (g *Game) OnUserBet(b []byte) {
	data := &shz.UserBet{}
	senddata := new(shz.BetRes)
	proto.Unmarshal(b, data)

	if !g.CheckUserBet(data.BetMoney) {
		return
	}

	cheat := g.GetCheatValue()

	g.GetIconRes(int64(cheat))
	if g.testMsg != nil && g.testMsg.Result == 2 {
		g.XiaoMaLiTimes = 3
	}

	// busstype := int32(200501)
	// if g.Fi.Type != 0 {
	// busstype = 200504
	// }
	odds := g.icon.Odds
	senddata.Odds = int32(odds)
	senddata.XiaoMaLi = int32(g.XiaoMaLiTimes)
	senddata.Gold = int64(odds) * int64(g.LastBet)
	senddata.ResType = int32(g.Fi.Type)
	//senddata.Cheat = int32(cheat)
	//给玩家加钱
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("结算前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), senddata.Gold, g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("结算后:", score.GetScoreStr(g.user.GetScore())))
	g.UserTotalWin += senddata.Gold
	senddata.UserGold = g.user.GetScore()
	g.LastWin = int64(odds) * int64(g.LastBet)
	senddata.IconArr = append(senddata.IconArr, g.icon.Iconarr...)
	for _, v := range g.icon.Line {
		li := new(shz.LineInfo)
		li.Count = v.Count
		li.Gold = v.Gold * int64(g.LastBet)
		li.Index = v.Index
		li.Start = v.Start
		senddata.Line = append(senddata.Line, li)
	}

	//log.Traceln("下注数据", senddata)
	g.user.SendMsg(int32(shz.ReMsgIDS2C_BetRet), senddata)

	str := fmt.Sprint(time.Now().Format("2006-1-2 15:04:05"), " 用户ID：", g.user.GetID(), g.CheatValue, "作弊率：", cheat, " 常规结果数组：", senddata.IconArr, "，扣钱：", (data.BetMoney*g.Line)/100,
		".", (data.BetMoney*g.Line)%100, "，加钱：", senddata.Gold/100, ".", senddata.Gold%100)
	g.table.WriteLogs(g.user.GetID(), str)
	g.PaoMaDeng(senddata.Gold, false)
	g.user.SendChip(int64(data.BetMoney * g.Line))
	g.testMsg = nil
}

func (g *Game) GetIconRes(cheatvalue int64) {
TEST_LABEL:
	isFull := g.Fi.GetFullIcon(int(cheatvalue))
	if g.testMsg != nil && g.testMsg.Result == 3 && !isFull {
		goto TEST_LABEL
	}
	if isFull {
		g.icon.Iconarr = g.Fi.Iconarr
		g.icon.Gettotalodds(g.lbcfg)
		g.icon.Odds += g.Fi.Odd
	} else {
		for {
			g.icon.GetIcon(cheatvalue, g.lbcfg, false, false, nil, nil)
			if !g.Fi.IsFullIcon(g.icon.Iconarr) {
				break
			}
		}
	}

	/*
		g.icon.Iconarr = make([]int32, 0)
		tmp := [...]int32{5, 3, 1, 8, 4, 1, 2, 7, 2, 0, 8, 5, 6, 0, 8}
		for _, v := range tmp {
			g.icon.Iconarr = append(g.icon.Iconarr, v)
		}

		g.icon.Gettotalodds(g.lbcfg)
	*/
	g.XiaoMaLiTimes = 0
	g.SmallGameGetOdds = 0
	g.XiaoMaLiGold = 0
	//计算压线上分散元素的个数
	wildcount := g.icon.GetOnLineScatterIconCount(g.lbcfg.Wild.IconId, g.lbcfg)
	for _, v := range wildcount {
		if v == 3 {
			g.XiaoMaLiTimes += 1
		} else if v == 4 {
			g.XiaoMaLiTimes += 2
		} else if v >= 5 {
			g.XiaoMaLiTimes += 3
		}
	}
}

func (g *Game) GetIconOdds() int {
	return g.icon.Odds
}

func (g *Game) XiaoMaLi(b []byte) {
	if g.XiaoMaLiTimes <= 0 {
		return
	}

	Odds := 0
	var iconid int
	var outindex int
	cheat := g.GetCheatValue()
	CheatCfg := g.xml.GetIconWeight(int64(cheat))
	var iniconarr []int

	for {
		iconid, outindex, iniconarr = g.xml.GetIconInfo(CheatCfg.Innerweight, CheatCfg)
		if iconid == 8 && g.SmallGameGetOdds == 0 {
			continue
		}

		tmpodds := xiaomali.GetOdds(int64(iconid), iniconarr, g.xml.IconAward)
		if iconid == 8 && tmpodds > 0 {
			continue
		}

		if g.SmallGameGetOdds > CheatCfg.Limit && iconid == 8 {
			break
		}

		if (tmpodds + g.SmallGameGetOdds) > CheatCfg.Limit {
			continue
		}

		break
	}

	//log.Traceln(iniconarr, iconid, outindex)
	if iconid == 8 {
		g.XiaoMaLiTimes--
		if g.XiaoMaLiTimes == 0 {
			g.SmallGameGetOdds = 0
		}
	} else {
		Odds = xiaomali.GetOdds(int64(iconid), iniconarr, g.xml.IconAward)
		if Odds != 0 {
			g.SmallGameGetOdds += Odds
			g.XiaoMaLiGold += int64(Odds) * int64(g.LastBet) * int64(g.Line)
			g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("小玛莉结算前:", score.GetScoreStr(g.user.GetScore())))
			g.user.SetScore(g.table.GetGameNum(), int64(Odds)*int64(g.LastBet)*int64(g.Line),
				g.table.GetRoomRate())
			g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("小玛莉结算后:", score.GetScoreStr(g.user.GetScore())))
			g.UserTotalWin += int64(Odds) * int64(g.LastBet) * int64(g.Line)
		}
	}

	senddata := new(shz.XiaoMaLiRes)
	for i := 0; i < len(iniconarr); i++ {
		senddata.InIcon = append(senddata.InIcon, (int32(iniconarr[i])))
	}

	senddata.Times = int32(g.XiaoMaLiTimes)
	senddata.Gold = int64(Odds) * int64(g.LastBet) * int64(g.Line)

	senddata.OutIndex = int32(outindex)

	if g.XiaoMaLiTimes == 0 {
		senddata.Exit = true
	} else {
		senddata.Exit = false
	}
	// arrstr := fmt.Sprint(senddata.InIcon)
	// g.user.SetEndCards(arrstr)
	g.user.SendMsg(int32(shz.ReMsgIDS2C_XiaMaLiRet), senddata)
	str := fmt.Sprint("作弊率：", cheat, " 小玛丽结果：", senddata.OutIndex, " ", senddata.InIcon,
		"，加钱：", senddata.Gold/100, ".", senddata.Gold%100)
	g.table.WriteLogs(g.user.GetID(), str)
	if g.XiaoMaLiTimes == 0 {
		g.PaoMaDeng(g.XiaoMaLiGold, true)
	}
}

//此函数给测试用
func (g *Game) GetXiaoMaLiOdds(cheatvalue int64) int {
	Odds := 0

	var iconid int
	var iniconarr []int
	CheatCfg := g.xml.GetIconWeight(cheatvalue)
	for {
		iconid, _, iniconarr = g.xml.GetIconInfo(CheatCfg.Innerweight, CheatCfg)
		tmpodds := xiaomali.GetOdds(int64(iconid), iniconarr, g.xml.IconAward)
		if iconid == 8 && g.SmallGameGetOdds == 0 {
			continue
		}

		if iconid == 8 && tmpodds > 0 {
			continue
		} else if (tmpodds + g.SmallGameGetOdds) > CheatCfg.Limit {
			continue
		}

		break
	}

	if iconid == 8 {
		g.XiaoMaLiTimes--
		if g.XiaoMaLiTimes == 0 {
			g.SmallGameGetOdds = 0
		}
	} else {
		//log.Traceln("小玛丽内圈 ", iniconarr, iconid, "小玛丽外圈", g.xml.IconAward, CheatCfg.Limit)
		Odds = xiaomali.GetOdds(int64(iconid), iniconarr, g.xml.IconAward)
		if Odds != 0 {
			g.SmallGameGetOdds += Odds
		}
	}

	return Odds * 9
}

func (g *Game) GetIconCount() int {
	return g.icon.Count
}

func (g *Game) UserBiBei(b []byte) {
	if g.LastWin == 0 {
		return
	}
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("比倍扣钱前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), -g.LastWin, g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("比倍扣钱后:", score.GetScoreStr(g.user.GetScore())))
	// g.user.SetChip(g.LastWin)
	data := &shz.C2SBiBei{}
	proto.Unmarshal(b, data)
	//获得两个骰子

	cheat := g.GetCheatValue()
	d1, d2, bWin := bibei.GetBiBeiRes(g.bbc, cheat, int(data.Result))

	senddata := new(shz.S2CBiBeiRes)
	LastWin := g.LastWin
	//如果玩家赢了获取倍数
	if bWin {
		senddata.Gold = int64(bibei.GetOdds(g.bbc, d1, d2)+1) * g.LastWin
		g.LastWin = senddata.Gold
		g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("比倍结算前:", score.GetScoreStr(g.user.GetScore())))
		g.user.SetScore(g.table.GetGameNum(), g.LastWin, g.table.GetRoomRate())
		g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("比倍结算后:", score.GetScoreStr(g.user.GetScore())))
		g.UserTotalWin += g.LastWin
	} else {
		g.LastWin = 0
	}

	senddata.UserGold = g.user.GetScore()
	senddata.Dice1 = int32(d1)
	senddata.Dice2 = int32(d2)

	str := fmt.Sprintf("比倍结果：用户押注：%v,押注金额：%v,开出塞子点数：%v，%v，用户赢钱：%v", data.Result, LastWin, d1, d2, g.LastWin)
	g.table.WriteLogs(g.user.GetID(), str)

	// g.user.SendChip()
	g.user.SendMsg(int32(shz.ReMsgIDS2C_BiBeiRet), senddata)
}

func (g *Game) CheckUserBet(BetMoney int32) bool {
	//这里检查用户的钱是否足够
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
		msg := new(shz.BetFail)
		msg.FailID = 2
		msg.Reson = "数据异常"
		g.user.SendMsg(int32(shz.ReMsgIDS2C_BetFailID), msg)
		return false
	}
	if g.user.GetScore() < int64(BetMoney) {
		msg := new(shz.BetFail)
		msg.FailID = 1
		msg.Reson = "金币不足，请充值！"

		g.user.SendMsg(int32(shz.ReMsgIDS2C_BetFailID), msg)
		return false
	}

	//这里扣钱
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("下注前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), int64(-BetMoney*g.Line), g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("下注后:", score.GetScoreStr(g.user.GetScore())))
	if int64(BetMoney) != g.LastBet {
		//这里检查用户输入参数是否有问题
		g.LastBet = int64(BetMoney)
	}
	g.AllBet += int64(BetMoney)

	return true
}

//检查结果是否符合要求
func (g *Game) CheckIconRet(BetMoney int32) bool {
	iconid := g.icon.Iconarr[0]
	ret := true
	for i := 1; i < len(g.icon.Iconarr); i++ {
		//这里算全图
		if iconid != g.icon.Iconarr[i] {
			//不是全图，看下是否是花图
			if (iconid < 3 && g.icon.Iconarr[i] > 3) ||
				(iconid > 3 && g.icon.Iconarr[i] < 3) {
				ret = false
				break
			}
		}
	}
	return ret
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

		g.XiaoMaLiTimes, _ = strconv.Atoi(arrstr[0])
		g.XiaoMaLiGold, _ = strconv.ParseInt(arrstr[1], 10, 0)
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

func (g *Game) PaoMaDeng(Gold int64, bXiaoMali bool) {
	configs := g.table.GetMarqueeConfig()
	for _, v := range configs {
		if bXiaoMali {
			special, _ := strconv.Atoi(v.SpecialCondition)
			if special == 1 && Gold >= v.AmountLimit {
				err := g.table.CreateMarquee(g.user.GetNike(), Gold, "小玛丽游戏", v.RuleId)
				if err != nil {
					log.Debugf("创建跑马灯错误：%v", err)
				}
			}
		} else if Gold >= v.AmountLimit && len(v.SpecialCondition) == 0 {
			err := g.table.CreateMarquee(g.user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Debugf("创建跑马灯错误：%v", err)
			}
		}
	}
}

func (game *Game) handleTest(bts []byte) {
	msg := new(shz.TestMsg)
	if err := proto.Unmarshal(bts, msg); err != nil {
		return
	}
	switch msg.Result {
	case 2, 3:
	default:
		return
	}
	game.testMsg = msg
}
