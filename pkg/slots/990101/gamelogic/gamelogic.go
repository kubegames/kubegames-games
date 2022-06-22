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
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	laba "github.com/kubegames/kubegames-games/pkg/slots/990101/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	table         table.TableInterface   // table interface
	user          player.PlayerInterface //用户
	lbcfg         *config.LabaConfig     //拉霸配置
	xml           *xiaomali.XiaoMaLiCfg  //小玛丽配置
	icon          iconlogic.Iconinfo     //图形算法逻辑
	FreeGameTimes int                    //剩余免费游戏次数
	FreeGameGold  int64                  //免费游戏获取到的金币
	XiaoMaLiTimes int                    //小玛丽次数
	xiaoMaLiGold  int64                  //小玛丽获取的金币
	LastBet       int64                  //最近一次用户下注的钱
	SmallGameOdds int                    //小游戏中奖的倍数
	BetArr        []int32                //下注配置
	UserTotalWin  int64                  //玩家总赢钱，算产出
	curr          int64
	AllBet        int64 // 所有下注
	testMsg       *laba.TestMsg
	CheatValue    string //个人 系统
}

//初始化
func (g *Game) Init(lb *config.LabaConfig, xml *xiaomali.XiaoMaLiCfg) {
	g.lbcfg = lb
	g.xml = xml
	g.FreeGameTimes = 0
	g.SmallGameOdds = 0
}

//绑定机器人接口
func (g *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	return nil
}

//用户押注
func (g *Game) OnUserBet(b []byte) {

	data := &laba.UserBet{}
	senddata := new(laba.BetRes)
	proto.Unmarshal(b, data)
	if !g.CheckUserBet(data.BetMoney) {
		return
	}

	cheat := g.GetCheatValue()
	g.dealWithTest(senddata)
	bfree := g.FreeGameTimes > 0
	tmpfree := g.FreeGameTimes
	g.XiaoMaLiTimes = 0
	g.SmallGameOdds = 0
	g.xiaoMaLiGold = 0
	//小游戏和免费游戏不同时出现
	for {
	TEST_LABEL:
		freegame, smallgame := g.GetIconRes(int64(cheat))
		if g.testMsg != nil && g.testMsg.Result == 0 && g.icon.Odds < 120 {
			goto TEST_LABEL
		}

		if freegame == 0 || smallgame == 0 {
			if bfree && smallgame != 0 {
				continue
			}
			g.FreeGameTimes += freegame
			g.XiaoMaLiTimes += smallgame
			break
		}

	}

	//免费游戏减一

	odds := g.icon.Odds

	if tmpfree != g.FreeGameTimes {
		senddata.BEnterFree = true
	} else {
		senddata.BEnterFree = false
	}

	g.dealWithTest(senddata)
	if bfree {
		g.FreeGameTimes--
		g.FreeGameGold += int64(odds) * int64(g.LastBet)
	} else {
		g.FreeGameGold = 0
	}

	senddata.Odds = int32(odds)
	senddata.Gold = int64(odds) * int64(g.LastBet)
	//senddata.Cheat = int32(cheat)
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("结算前", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), senddata.Gold, g.table.GetRoomRate())
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("结算后", score.GetScoreStr(g.user.GetScore())))

	//需要检查用户是否是最大的下注，todo
	rat := g.icon.Getjackpot(g.lbcfg)
	if rat > 0 {
		_, JackpotGold, _ := g.table.GetCollect(int32(rat), g.user.GetID())
		senddata.Jackpot = JackpotGold
		g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("奖金池结算前:", score.GetScoreStr(g.user.GetScore())))
		g.user.SetScore(g.table.GetGameNum(), senddata.Jackpot, g.table.GetRoomRate())
		g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("奖金池结算后:", score.GetScoreStr(g.user.GetScore())))
	}
	g.UserTotalWin += senddata.Gold + senddata.Jackpot
	senddata.UserGold = g.user.GetScore()
	senddata.XiaoMaLi = int32(g.XiaoMaLiTimes)
	senddata.FreeGames = int32(g.FreeGameTimes)
	senddata.IconArr = append(senddata.IconArr, g.icon.Iconarr...)
	for _, v := range g.icon.Line {
		li := new(laba.LineInfo)
		li.Count = v.Count
		li.Gold = v.Gold * int64(g.LastBet)
		li.Index = v.Index
		senddata.Line = append(senddata.Line, li)
	}

	AddGold := data.BetMoney * int32(g.lbcfg.LineCount)
	if bfree {
		AddGold = 0
	}

	// arrstr := fmt.Sprint(senddata.IconArr)
	// g.user.SetEndCards(arrstr)
	str := fmt.Sprint(time.Now().Format("2006-1-2 15:04:05"), " 用户ID：", g.user.GetID(), g.CheatValue, "作弊率：", cheat, " 结果数组：", senddata.IconArr, "，扣钱：", AddGold/100, ".", AddGold%100,
		"，加钱：", (senddata.Gold+senddata.Jackpot)/100, ".", (senddata.Gold+senddata.Jackpot)%100,
		"免费游戏次数:", senddata.FreeGames, "小玛丽次数:", senddata.XiaoMaLi)
	g.table.WriteLogs(g.user.GetID(), str)

	g.user.SendMsg(int32(laba.ReMsgIDS2C_BetRet), senddata)

	if senddata.Jackpot != 0 {
		g.PaoMaDeng(senddata.Gold+senddata.Jackpot, 1)
	} else {
		g.PaoMaDeng(senddata.Gold+senddata.Jackpot, 0)
	}
	g.user.SendChip(int64(AddGold))
	g.testMsg = nil
}

func (g *Game) GetIconRes(cheatvalue int64) (freegame int, smallgame int) {
	var iconFunc func(cheatvalue int64)
	if g.testMsg != nil && g.testMsg.Result == 3 {
		iconFuncs := []func(cheatvalue int64){
			func(cheatvalue int64) {
				g.icon.Iconarr[0] = 9
				g.icon.Iconarr[3] = 9
				g.icon.Iconarr[6] = 9
			},
			func(cheatvalue int64) {
				g.icon.Iconarr[0] = 9
				g.icon.Iconarr[3] = 9
				g.icon.Iconarr[6] = 9
				g.icon.Iconarr[9] = 9
			},
			func(cheatvalue int64) {
				g.icon.Iconarr[0] = 9
				g.icon.Iconarr[3] = 9
				g.icon.Iconarr[6] = 9
				g.icon.Iconarr[9] = 9
				g.icon.Iconarr[12] = 9
			},
		}
		iconFunc = iconFuncs[rand.Intn(len(iconFuncs))]
	} else {
		iconFunc = nil
	}

	g.icon.GetIcon(cheatvalue, g.lbcfg, g.FreeGameTimes > 0, false, nil, iconFunc)

	//测试用
	/*g.icon.Iconarr = make([]int32, 0)
	tmp := []int32{1, 1, 3, 0, 4, 0, 0, 1, 1, 5, 10, 10, 1, 10, 2}
	for _, v := range tmp {
		g.icon.Iconarr = append(g.icon.Iconarr, v)
	}
	g.icon.Gettotalodds(g.lbcfg)
	*/
	count := g.icon.Getfreegametimes(g.lbcfg)
	for _, v := range count {
		freegame += int(g.lbcfg.FreeGame.Times[v])
	}
	wildcount := g.icon.GetOnLineIconCount(g.lbcfg.Wild.IconId, g.lbcfg)
	for _, v := range wildcount {
		if v == 3 {
			smallgame += 1
		} else if v == 4 {
			smallgame += 2
		} else if v >= 5 {
			smallgame += 3
		}
	}

	return freegame, smallgame
}

func (g *Game) GetIconOdds() int {
	return g.icon.Gettotalodds(g.lbcfg)
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
		if iconid == 7 && g.SmallGameOdds == 0 {
			continue
		}

		tmpodds := xiaomali.GetOdds(int64(iconid), iniconarr, g.xml.IconAward)
		if iconid == 7 && tmpodds > 0 {
			continue
		}

		if g.SmallGameOdds > CheatCfg.Limit && iconid == 7 {
			break
		}

		if (tmpodds + g.SmallGameOdds) > CheatCfg.Limit {
			continue
		}

		break
	}

	if iconid == 7 {
		g.XiaoMaLiTimes--
		if g.XiaoMaLiTimes == 0 {
			g.SmallGameOdds = 0
		}
	} else {
		Odds = xiaomali.GetOdds(int64(iconid), iniconarr, g.xml.IconAward)
		if Odds != 0 {
			g.SmallGameOdds += Odds
			g.xiaoMaLiGold += int64(Odds) * int64(g.LastBet) * int64(g.lbcfg.LineCount)
			g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("小玛莉结算前:", score.GetScoreStr(g.user.GetScore())))
			g.user.SetScore(g.table.GetGameNum(), int64(Odds)*int64(g.LastBet)*int64(g.lbcfg.LineCount), g.table.GetRoomRate())
			g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("小玛莉结算后:", score.GetScoreStr(g.user.GetScore())))
			g.UserTotalWin += int64(Odds) * int64(g.LastBet) * int64(g.lbcfg.LineCount)
		}
	}

	senddata := new(laba.XiaoMaLiRes)
	for i := 0; i < len(iniconarr); i++ {
		senddata.InIcon = append(senddata.InIcon, (int32(iniconarr[i])))
	}

	senddata.Times = int32(g.XiaoMaLiTimes)
	senddata.Gold = int64(Odds) * int64(g.LastBet) * int64(g.lbcfg.LineCount)

	senddata.OutIndex = int32(outindex)
	if g.XiaoMaLiTimes == 0 {
		senddata.Exit = true
	} else {
		senddata.Exit = false
	}

	g.user.SendMsg(int32(laba.ReMsgIDS2C_XiaMaLiRet), senddata)

	str := fmt.Sprint("作弊率：", cheat, " 小玛丽结果：", senddata.OutIndex, " ", senddata.InIcon,
		"，加钱：", senddata.Gold/100, ".", senddata.Gold%100)
	// arrstr := fmt.Sprint(senddata.InIcon)
	// g.user.SetEndCards(arrstr)
	g.table.WriteLogs(g.user.GetID(), str)
	if g.XiaoMaLiTimes == 0 {
		g.PaoMaDeng(g.xiaoMaLiGold, 2)
	}
}

func (g *Game) CheckUserBet(BetMoney int32) bool {
	if g.FreeGameTimes > 0 {
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
		msg := new(laba.BetFail)
		msg.FailID = 2
		msg.Reson = "数据异常"
		g.user.SendMsg(int32(laba.ReMsgIDS2C_BetFailID), msg)
		return false
	}

	//这里检查用户的钱是否足够
	if g.user.GetScore() < int64(BetMoney) {
		msg := new(laba.BetFail)
		msg.FailID = 1
		msg.Reson = "金币不足，请充值！"

		g.user.SendMsg(int32(laba.ReMsgIDS2C_BetFailID), msg)
		log.Tracef("金币不足，请充值！")
		return false
	}

	//这里检查用户输入参数是否有问题，是否在下注的范围内
	//这里扣钱
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("下注前:", score.GetScoreStr(g.user.GetScore())))
	g.user.SetScore(g.table.GetGameNum(), int64(-BetMoney*int32(g.lbcfg.LineCount)), g.table.GetRoomRate())
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

//此函数给测试用
func (g *Game) GetXiaoMaLiOdds(cheatvalue int64) int {
	Odds := 0

	var iconid int
	var iniconarr []int
	CheatCfg := g.xml.GetIconWeight(cheatvalue)
	for {
		iconid, _, iniconarr = g.xml.GetIconInfo(CheatCfg.Innerweight, CheatCfg)
		tmpodds := xiaomali.GetOdds(int64(iconid), iniconarr, g.xml.IconAward)
		if iconid == 7 && g.SmallGameOdds == 0 {
			continue
		}

		if iconid == 7 && tmpodds > 0 {
			continue
		} else if (tmpodds + g.SmallGameOdds) > CheatCfg.Limit {
			continue
		}

		break
	}

	if iconid == 7 {
		g.XiaoMaLiTimes--
		if g.XiaoMaLiTimes == 0 {
			g.SmallGameOdds = 0
		}
	} else {
		Odds = xiaomali.GetOdds(int64(iconid), iniconarr, g.xml.IconAward)
		if Odds != 0 {
			g.SmallGameOdds += Odds
		}
	}

	return Odds * 9
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
		g.XiaoMaLiTimes, _ = strconv.Atoi(arrstr[2])
		g.xiaoMaLiGold, _ = strconv.ParseInt(arrstr[3], 10, 0)
		g.LastBet, _ = strconv.ParseInt(arrstr[4], 10, 0)
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
		log.Tracef("获取到群控的值为：%v", tmp)
		Prob = int32(tmp)
		if tmp == 0 {
			g.CheatValue += " 获取到系统作弊率为0 "
			Prob = 1000
		}
	}

	log.Tracef("获取到的值为：%v", Prob)
	return int(Prob)
}

func (g *Game) GetIconCount() int {
	return g.icon.Count
}

func (g *Game) PaoMaDeng(Gold int64, Type int) {
	configs := g.table.GetMarqueeConfig()
	for _, v := range configs {
		if Type != 0 {
			special, _ := strconv.Atoi(v.SpecialCondition)
			if Type == 1 && special == 1 && Gold >= v.AmountLimit {
				err := g.table.CreateMarquee(g.user.GetNike(), Gold, "彩金池", v.RuleId)
				if err != nil {
					log.Debugf("创建跑马灯错误：%v", err)
				}
			} else if Type == 2 && special == 2 && Gold >= v.AmountLimit {
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

func (game *Game) dealWithTest(br *laba.BetRes) {
	if game.testMsg == nil {
		return
	}
	switch game.testMsg.Result {
	case 1:
		br.BEnterFree = true
		br.FreeGames = 6
		game.FreeGameTimes = int(br.FreeGames)
	case 2:
		game.XiaoMaLiTimes = 3

	}
}

func (game *Game) handleTest(bts []byte) {
	msg := new(laba.TestMsg)
	if err := proto.Unmarshal(bts, msg); err != nil {
		return
	}
	switch msg.Result {
	case 0, 1, 2, 3:
	default:
		return
	}
	game.testMsg = msg
}
