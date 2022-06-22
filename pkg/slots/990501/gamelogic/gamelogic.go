package gamelogic

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/config"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/iconlogic"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	_777 "github.com/kubegames/kubegames-games/pkg/slots/990501/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
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
	LastWildIndex int64                  //wild出现的位置
	UserTotalWin  int64                  //玩家总赢钱，算产出

	curr int64

	AllBet        int64 // 总下注
	GameIconLogic ConfigLogic
	CheatValue    string //个人 系统
}

//初始化
func (g *Game) Init(lb *config.LabaConfig) {
	g.lbcfg = lb
	g.FreeGameTimes = 0
	g.Line = int32(lb.LineCount)
}

func (lbr *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	return nil
}

//用户押注
func (g *Game) OnUserBet(b []byte) {
	data := &_777.UserBet{}
	senddata := new(_777.BetRes)
	proto.Unmarshal(b, data)
	if !g.CheckUserBet(data.BetMoney) {
		return
	}

	cheat := g.GetCheatValue()
	bfree := g.FreeGameTimes > 0
	g.GetIconRes(int64(cheat))
	//免费游戏减一

	odds := g.icon.Odds
	// busstype := int32(201801)
	if bfree {
		// busstype = 201802
		g.FreeGameTimes--
		g.FreeGameGold += int64(odds) * int64(g.LastBet)
	} else {
		g.FreeGameGold = 0
	}

	//免费不触发免费
	if g.FreeGameTimes != 0 {
		senddata.BEnterFree = true
	} else {
		senddata.BEnterFree = false
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
		li := new(_777.LineInfo)
		li.Count = v.Count
		li.Gold = v.Gold * int64(g.LastBet)
		li.Index = v.Index
		senddata.Line = append(senddata.Line, li)
	}

	g.user.SendMsg(int32(_777.ReMsgIDS2C_BetRet), senddata)

	BetGold := data.BetMoney * g.Line
	if bfree {
		BetGold = 0
	}

	str := fmt.Sprint(time.Now().Format("2006-1-2 15:04:05"), " 用户ID：", g.user.GetID(), g.CheatValue, "作弊率：", cheat, " 结果数组：", senddata.IconArr, "，扣钱：", BetGold/100, ".",
		BetGold%100, "，加钱：", senddata.Gold/100, ".", senddata.Gold%100, "，免费次数：", senddata.FreeGames)
	// arrstr := fmt.Sprint(senddata.IconArr)
	// g.user.SetEndCards(arrstr)
	g.table.WriteLogs(g.user.GetID(), str)
	g.PaoMaDeng(senddata.Gold)
	g.user.SendChip(int64(BetGold))

	log.Debugf("使用的作弊率为：%v", cheat)
}

func (g *Game) GetIconRes(cheatvalue int64) {
	g.icon.Iconarr = make([]int32, 0)
	if g.FreeGameTimes > 0 {
		arr := g.GameIconLogic.GetIconArr(int(cheatvalue), true, int(g.LastWildIndex))
		for i := 0; i < 3; i++ {
			g.icon.Iconarr = append(g.icon.Iconarr, int32(arr[i]))
		}
		g.icon.Gettotalodds(g.lbcfg)
		g.ExterIcon(0)
		//g.icon.GetIcon(cheatvalue, g.lbcfg, g.FreeGameTimes > 0, g.ExterIcon, g.ChangeIcon)
		g.LastWildIndex = -1
	} else {
		arr := g.GameIconLogic.GetIconArr(int(cheatvalue), false, -1)
		for i := 0; i < 3; i++ {
			g.icon.Iconarr = append(g.icon.Iconarr, int32(arr[i]))
		}
		g.icon.Gettotalodds(g.lbcfg)
		g.ExterIcon(0)
		//g.icon.GetIcon(cheatvalue, g.lbcfg, g.FreeGameTimes > 0, g.ExterIcon, nil)
		g.LastWildIndex = -1
		freegame := 0
		count := 0
		for i, v := range g.icon.Iconarr {
			if v == 6 {
				g.LastWildIndex = int64(i)
				freegame = 1
				count++
			}
		}

		if count == 3 {
			g.LastWildIndex = -1
			freegame = 0
		}

		g.FreeGameTimes = freegame
	}

	/*
		//测试用
		g.icon.Iconarr = make([]int32, 0)
		tmp := [...]int32{1, 10, 8, 5, 8, 10, 9, 10, 1, 2, 1, 2, 6, 7, 0}
		for _, v := range tmp {
			g.icon.Iconarr = append(g.icon.Iconarr, v)
		}
	*/
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
		msg := new(_777.BetFail)
		msg.FailID = 2
		msg.Reson = "数据异常"
		g.user.SendMsg(int32(_777.ReMsgIDS2C_BetFailID), msg)
		return false
	}
	//这里检查用户的钱是否足够
	if g.user.GetScore() < int64(BetMoney) {
		msg := new(_777.BetFail)
		msg.FailID = 1
		msg.Reson = "金币不足，请充值！"

		g.user.SendMsg(int32(_777.ReMsgIDS2C_BetFailID), msg)
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

func (g *Game) GetRoomconfig() {
	/*str := g.table.GetAdviceConfig()
	log.Tracef("配置 %v", str)
	js, err := simplejson.NewJson([]byte(str))
	if err != nil {
		fmt.Printf("解析房间配置失败 err%v\n", err)
		fmt.Printf("%v\n", str)
		return
	}
	*/
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
		g.LastWildIndex, _ = strconv.ParseInt(arrstr[2], 10, 0)
		g.LastBet, _ = strconv.ParseInt(arrstr[3], 10, 0)
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

//额外的图片情形
var iconarr2 = [6][3]int32{{0, 3, 0}, {3, 0, 3}, {4, 1, 4}, {1, 4, 1}, {2, 5, 2}, {5, 2, 5}}

func (g *Game) ExterIcon(v int64) {
	if g.icon.Odds > 0 {
		return
	}

	isarr := true
	for j := 0; j < 3; j++ {
		if g.icon.Iconarr[j] != 3 && g.icon.Iconarr[j] != 4 && g.icon.Iconarr[j] != 5 &&
			g.icon.Iconarr[j] != 6 {
			isarr = false
			break
		}
	}

	if isarr {
		g.icon.Odds = 2
		return
	}

	bar := -1
	seven := -1
	three := -1
	for j := 0; j < 3; j++ {
		if g.icon.Iconarr[j] <= int32(BarIconID[2]) && bar == -1 {
			bar = int(g.icon.Iconarr[j])
		} else if g.icon.Iconarr[j] <= int32(SevenIconID[2]) &&
			g.icon.Iconarr[j] > int32(BarIconID[2]) && seven == -1 {
			seven = int(g.icon.Iconarr[j])
		} else {
			three = int(g.icon.Iconarr[j])
		}
	}

	if bar+3 == seven && (bar == three || three == seven || three == 6) {
		g.icon.Odds = 1
	}
}

func (g *Game) ChangeIcon(cheatvalue int64) {
	g.icon.Iconarr[g.LastWildIndex] = 6
}

func (g *Game) PaoMaDeng(Gold int64) {
	configs := g.table.GetMarqueeConfig()
	for _, v := range configs {
		if Gold >= v.AmountLimit {
			err := g.table.CreateMarquee(g.user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Debugf("创建跑马灯错误：%v", err)
			}
		}
	}
}
