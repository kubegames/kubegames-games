package gamelogic

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	majiangcom "github.com/kubegames/kubegames-games/internal/pkg/majiang"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/battle/940101/config"
	"github.com/kubegames/kubegames-games/pkg/battle/940101/def"
	errenmajiang "github.com/kubegames/kubegames-games/pkg/battle/940101/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
	"github.com/tidwall/gjson"
)

const (
	UserNum = 2
)

var IsTest = false

var (
	TestStatus0 = 1 // 配置发牌
	TestStatus1 = 2 // 配置摸牌
)

type Game struct {
	table                table.TableInterface               //框架桌子接口
	MaJiang              majiangcom.MaJiang                 //麻将类
	ud                   [UserNum]UserData                  //用户数据使用座位号为索引
	CurrOperatUser       int                                //当前操作的玩家椅子号
	Zhuang               int                                //庄家椅子号
	CurrOperation        int                                //本轮有那些操作
	OptLevel             int                                //已经操作玩家最高的座位号
	TimerJob             *table.Job                         //job
	LastOutCard          int                                //最后一次出的牌
	Round                int                                // 游戏轮次
	AllOutCardMaps       [majiangcom.MaxCardValue]int       // 打出去的牌
	IsStartGame          bool                               // 是否开始游戏
	SendSceneIds         [UserNum]*errenmajiang.RepeatedInt // 桌上用户信息
	UserAvatar           [UserNum]string                    // 桌上用户信息
	UserName             [UserNum]string                    // 桌上用户信息
	UserMoney            [UserNum]int64                     // 桌上用户信息
	UserCity             [UserNum]string                    // 桌上用户城市
	TestUserIdx          int                                // 配牌用户下表
	ReUpLineOutCardsMaps map[int][majiangcom.MaxCardValue]int
	TouZi                []int32
	STARTCARDS           [2]string
	SETTELCARDS          [2]string
}

func (g *Game) CloseTable() {
}

//定庄
func (g *Game) SetZhuang() {
	g.IsStartGame = true
	if g.TestUserIdx > -1 && g.ud[g.TestUserIdx].TestFlag == int32(TestStatus0) {
		if len(majiangcom.GetHandCards(g.ud[g.TestUserIdx].HandCards)) == 14 {
			g.Zhuang = g.TestUserIdx
		} else {
			g.Zhuang = (g.TestUserIdx + 1) % UserNum
		}
	} else {
		g.Zhuang = rand.Intn(UserNum)
	}
	//二人麻将有万，字牌，白板，花牌
	HasCard := majiangcom.CardTypeWan | majiangcom.CardTypeZi |
		majiangcom.CardTypeBai | majiangcom.CardTypeHua
	g.MaJiang.InitMaJiang(HasCard)
	for i := 0; i < UserNum; i++ {
		if g.TestUserIdx == i {
			continue
		}
		g.ud[i].ResetData()
	}
	g.ReUpLineOutCardsMaps = map[int][majiangcom.MaxCardValue]int{}
	g.SendTouZi()
}

func (g *Game) SendTouZi() {
	msg := new(errenmajiang.StartMovieMsg)
	msg.MovieTime = config.ErRenMaJiang.StartMovie
	r := [2]int32{rand.Int31n(6) + 1, rand.Int31n(6) + 1}
	g.TouZi = r[:]
	msg.Zhuang = int32(g.Zhuang)
	msg.TouZi = g.TouZi
	g.table.Broadcast(int32(errenmajiang.ReMsgIDS2C_StartMovie), msg)
	if g.TestUserIdx > -1 && g.ud[g.TestUserIdx].TestFlag == int32(TestStatus0) {
		g.TimerJob, _ = g.table.AddTimer(int64(config.ErRenMaJiang.StartMovie), g.TestDealCards) // 开始动画时间
	} else {
		log.Debugf("开始动画，开始发牌 %v", config.ErRenMaJiang.StartMovie)
		g.TimerJob, _ = g.table.AddTimer(int64(config.ErRenMaJiang.StartMovie), g.DealCards) // 开始动画时间
	}
}

//开局发牌
func (g *Game) DealCards() {
	g.IsStartGame = true
	if g.TestUserIdx != -1 {
		return
	}
	g.ud[g.Zhuang].HandCards = [majiangcom.MaxCardValue]int{}
	g.ud[(g.Zhuang+1)%UserNum].HandCards = [majiangcom.MaxCardValue]int{}
	g.MaJiang.FlushCards()
	var tmp [UserNum][]int32
	//给玩家发牌
	//庄多一张
	//tmp[g.Zhuang] = append(tmp[g.Zhuang], int32(v))
	for i := 0; i < UserNum; i++ {
		for m := 0; m < 13; m++ {
			v := g.MaJiang.DealCard()
			tmp[i] = append(tmp[i], int32(v))
		}
	}
	tmp = g.HanCardsWeight(tmp)
	v := g.MaJiang.DealCard()
	g.ud[g.Zhuang].HandCards[v] += 1
	tmp[g.Zhuang] = append(tmp[g.Zhuang], int32(v))
	g.ud[g.Zhuang].LastAddCard = v

	for i := 0; i < UserNum; i++ {
		msg := new(errenmajiang.DealCardMsg)
		g.ud[i].HandCards = majiangcom.InitTestCards(tmp[i])
		msg.HandCards = tmp[i]
		msg.ChairId = int32(i)
		g.ud[i].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_DealCard), msg)
		if IsTest {
			g.ud[(i+1)%UserNum].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_DealCard), msg)
		}
	}

	// 记录开始手牌
	next := (g.Zhuang + 1) % UserNum
	cheatvalue := g.GetCheatValue(int32(g.Zhuang))
	robotstr := "否"
	if g.ud[g.Zhuang].User.IsRobot() {
		robotstr = "是"
	}

	g.STARTCARDS[g.Zhuang] = fmt.Sprintf("用户id:%v 机器人：%v %v作弊率:%v\t\t庄家:%v\t 开局手牌:%v \r\n",
		g.ud[g.Zhuang].User.GetID(), robotstr, g.ud[g.Zhuang].CheatSrc, cheatvalue, "是",
		majiangcom.GetHandCardString(g.ud[g.Zhuang].HandCards))
	cheatvalue = g.GetCheatValue(int32(next))
	robotstr = "否"
	if g.ud[next].User.IsRobot() {
		robotstr = "是"
	}
	g.STARTCARDS[next] = fmt.Sprintf("用户id:%v 机器人：%v %v作弊率:%v\t\t 庄家:%v\t开局手牌:%v \r\n",
		g.ud[next].User.GetID(), robotstr, g.ud[g.Zhuang].CheatSrc, cheatvalue, "否",
		majiangcom.GetHandCardString(g.ud[next].HandCards))

	// 游戏轮次
	g.Round = 0
	g.TimerJob, _ = g.table.AddTimer(int64(6000), g.SendFirstBuHua) // 补花动画消息
}

//发送第一次发牌补花
func (g *Game) SendFirstBuHua() {
	// 补花
	temp := 0
	temp += g.BuHua()
	if temp != 0 {
		temp = int(config.ErRenMaJiang.BuHuaMovie)
	}

	// 摸牌
	g.ud[g.Zhuang].MoPai(0)
	g.ud[g.Zhuang].HandCards[0] = 0
	if g.ud[g.Zhuang].Opt&majiangcom.OptTypeHu != 0 {
		g.ud[g.Zhuang].HuOptCard = int32(g.ud[g.Zhuang].LastAddCard)
	}
	// 广播开始出牌
	g.CurrOperatUser = g.Zhuang
	g.CurrOperation = g.ud[g.Zhuang].Opt

	//闲家最后一张牌置为0
	g.ud[(g.Zhuang+1)%UserNum].LastAddCard = 0

	g.TimerJob, _ = g.table.AddTimer(int64(temp), g.SendUserOptMsg) // 补花动画消息
}

func (g *Game) TestDealCards() {
	g.IsStartGame = true
	g.MaJiang.FlushCards()
	var tmp [UserNum][]int32
	for i := 0; i < UserNum; i++ {
		if i == g.TestUserIdx {
			continue
		}
		g.ud[(g.TestUserIdx+1)%UserNum].HandCards = [majiangcom.MaxCardValue]int{}
		if g.TestUserIdx != g.Zhuang {
			v := g.MaJiang.DealCard()
			g.ud[g.Zhuang].HandCards[v] += 1
			tmp[g.Zhuang] = append(tmp[g.Zhuang], int32(v))
		}
		for m := 0; m < 13; m++ {
			v := g.MaJiang.DealCard()
			g.ud[i].HandCards[v] += 1
			tmp[i] = append(tmp[i], int32(v))
			if i == g.Zhuang {
				g.ud[i].LastAddCard = v
			}
		}

		msg := new(errenmajiang.DealCardMsg)
		for j := 0; j < len(tmp[i]); j++ {
			msg.HandCards = append(msg.HandCards, tmp[i][j])
		}
		//发送发牌消息
		g.ud[i].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_DealCard), msg)
	}

	if g.TestUserIdx != g.Zhuang {
		g.ud[g.TestUserIdx].LastAddCard = 0
	}

	log.Debugf("最后一张牌%v %v", g.ud[g.Zhuang].LastAddCard, g.ud[g.Zhuang].HuOptCard)
	// 记录开始手牌
	next := (g.Zhuang + 1) % UserNum
	g.STARTCARDS[g.Zhuang] = fmt.Sprintf("用户id:%v \r\n 开始作弊率:%v\t\t是否庄家:%v\t手牌:%v \r\n",
		g.GetCheatValue(int32(g.Zhuang)), g.ud[g.Zhuang].User.GetID(), "是", majiangcom.GetHandCardString(g.ud[g.Zhuang].HandCards),
	)

	g.STARTCARDS[next] = fmt.Sprintf("开局手牌 \r\n 开始作弊率:%v\t用户id:%v\t是否庄家:%v\t手牌:%v \r\n",
		g.GetCheatValue(int32(next)), g.ud[next].User.GetID(), "否", majiangcom.GetHandCardString(g.ud[next].HandCards),
	)
	// 游戏轮次
	g.Round = 0
	// 补花
	log.Debugf("最后一张牌%v %v", g.ud[g.Zhuang].LastAddCard, g.ud[g.Zhuang].HuOptCard)
	temp := g.BuHua()
	log.Debugf("最后一张牌%v %v", g.ud[g.Zhuang].LastAddCard, g.ud[g.Zhuang].HuOptCard)
	// 摸牌
	g.ud[g.Zhuang].MoPai(0)
	log.Debugf("最后一张牌%v %v", g.ud[g.Zhuang].LastAddCard, g.ud[g.Zhuang].HuOptCard)
	g.ud[g.Zhuang].HandCards[0] = 0
	if g.ud[g.Zhuang].Opt&majiangcom.OptTypeHu != 0 {
		g.ud[g.Zhuang].HuOptCard = int32(g.ud[g.Zhuang].LastAddCard)
	}
	log.Debugf("最后一张牌%v %v", g.ud[g.Zhuang].LastAddCard, g.ud[g.Zhuang].HuOptCard)
	// 广播开始出牌
	g.CurrOperatUser = g.Zhuang
	g.CurrOperation = g.ud[g.Zhuang].Opt
	if !g.IsStartGame {
		return
	}
	g.TimerJob, _ = g.table.AddTimer(int64(temp), g.SendUserOptMsg) // 补花动画消息
}

func (g *Game) IsTianHu(mask, noMask *int64) {
	if g.CurrOperatUser == g.Zhuang && g.Round == 0 &&
		len(g.ud[g.CurrOperatUser].OptCards) == 0 {
		*mask |= def.TianHu
		*noMask |= def.DanDiao | def.BianZhang | def.KanZhuang | def.ZiMo
	}
}

func (g *Game) IsDiHu(mask, noMask *int64) {
	tempUser := g.ud[g.CurrOperatUser]
	tempNumber := tempUser.GangPaiNumber + tempUser.AnGangPaiNumber + tempUser.ChiPaiNumber + tempUser.PengPaiNumber
	if g.Round < 1 && g.CurrOperatUser != g.Zhuang &&
		tempNumber < 1 && g.ud[g.CurrOperatUser].IsZiMo {
		*mask |= def.DiHu
	}
}

func (g *Game) IsZiMo(mask, noMask *int64) {
	//if int32(g.ud[g.CurrOperatUser].LastAddCard) == g.ud[g.CurrOperatUser].HuOptCard {
	if g.ud[g.CurrOperatUser].IsZiMo {
		log.Debugf("是自摸")
		//g.ud[g.CurrOperatUser].IsZiMo = true
		*mask |= def.ZiMo
		// 是否是妙手回春
		if g.MaJiang.GetLastCardsCount() == 0 {
			*mask |= def.MiaoShouHuiChun
			*noMask |= def.ZiMo
		}
		// 不求人
		tempUserData := g.ud[g.CurrOperatUser]
		//吃碰杠都没有，不记番里面还要有不求人
		if (tempUserData.GangPaiNumber+tempUserData.PengPaiNumber+tempUserData.ChiPaiNumber) == 0 &&
			*noMask&def.BuQiuRen == 0 {
			*mask |= def.BuQiuRen
			*noMask |= def.ZiMo | def.MengQianQing
		}
		// 是否杠上花
		tempOpt := g.ud[g.CurrOperatUser].OptCards
		if len(tempOpt) < 1 {
			return
		}

		for i := 0; i < len(tempOpt); i++ {
			log.Debugf("%v操作%v", g.ud[g.CurrOperatUser].User.GetID(), tempOpt[i])
		}

		if tempOpt[len(tempOpt)-1].Opt&0xE0 != 0 {
			*mask |= def.GangShangKaiHua
			*noMask |= def.ZiMo
		}
	}
}

func (g *Game) IsHaiDiLaoYue(mask, noMask *int64) {
	if g.ud[g.CurrOperatUser].LastAddCard == 0 && g.ud[(g.CurrOperatUser+1)%UserNum].LastAddCard == 0 && g.MaJiang.GetLastCardsCount() == 0 {
		*mask |= def.HaiDiLaoYue
	}
}

func (g *Game) IsHuJueZhang(mask, noMask *int64) {
	if g.AllOutCardMaps[g.ud[g.CurrOperatUser].HuOptCard] == 3 {
		*mask |= def.HeJueZhang
	}
}

func (g *Game) IsQuanQiuRen(mask, noMask *int64) {
	gangNumber := g.ud[g.CurrOperatUser].GangPaiNumber
	pengNumber := g.ud[g.CurrOperatUser].PengPaiNumber
	chiNumber := g.ud[g.CurrOperatUser].ChiPaiNumber
	if (gangNumber > 0 || pengNumber > 0 || chiNumber > 0) &&
		g.ud[g.CurrOperatUser].AnGangPaiNumber == 0 &&
		len(majiangcom.GetHandCards(g.ud[g.CurrOperatUser].HandCards)) == 2 &&
		!g.ud[g.CurrOperatUser].IsZiMo &&
		!g.ud[g.CurrOperatUser].HasMingGang {
		*mask |= def.QuanQiuRen
		*noMask |= def.DanDiao
	}
}

func (g *Game) SendOutCardMessage(index int) {
	msg := new(errenmajiang.BroadOutCardMsg)
	msg.OutCardTime = config.ErRenMaJiang.OutCardTime
	msg.OutUserIndex = int32(index)
	g.table.Broadcast(int32(errenmajiang.ReMsgIDS2C_BroadOutCard), msg)
}

func (g *Game) AutoPalyCard() {
	msg := &errenmajiang.OptRetMsg{}
	msg.Opt = majiangcom.OptTypeOutCard
	tmpCard := g.ud[g.CurrOperatUser].LastAddCard
	if tmpCard == 0 {
		for i := majiangcom.Bai[0]; i >= majiangcom.Wan[0]; i-- {
			if g.ud[g.CurrOperatUser].HandCards[i] > 0 {
				tmpCard = i
				break
			}
		}
	}
	msg.CardValue = int32(tmpCard)
	msg.CardIndex = int32(g.CurrOperatUser)
	//if msg.CardIndex == 0 {
	//	msg.CardIndex = -1
	//}
	g.OnUserOutCard(msg, g.ud[g.CurrOperatUser].User)
}

func (g *Game) OutTimeAutoCheckStatus() {
	g.ud[g.CurrOperatUser].AutoStatus = true
	msg := new(errenmajiang.AutoOutCard)
	msg.Auto = g.ud[g.CurrOperatUser].AutoStatus
	g.ud[g.CurrOperatUser].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_Auto), msg)
}

//补花
func (g *Game) BuHua() int {
	temp := 0
	for i := 0; i < UserNum; i++ {
		msg := new(errenmajiang.BuHuaMsg)
		msg1 := new(errenmajiang.BuHuaMsg)
		msg.ChairId = int32(i)
		msg1.ChairId = int32(i)
		for m := majiangcom.Hua[0]; m < majiangcom.MaxCardValue; m++ {
			if g.ud[i].HandCards[m] == 1 {
				g.ud[i].Hua = append(g.ud[i].Hua, int32(m))
				msg.BuHuaDatas = append(msg.BuHuaDatas, int32(m))
				msg1.BuHuaDatas = append(msg1.BuHuaDatas, int32(m))
				g.ud[i].HandCards[m] = 0
				for {
					v := 0
					if g.ud[i].TingPai {
						v = g.TingMopai(i)
					} else {
						v = g.CtrlWater(g.GetCheatValue(int32(i)), i)
					}
					if v == 0 {
						g.Settle(true)
						return 0
					}
					temp += int(config.ErRenMaJiang.BuHuaMovie)
					//如果还是花继续补花
					if v >= majiangcom.Hua[0] {
						g.ud[i].Hua = append(g.ud[i].Hua, int32(v))
						msg.BuHuaDatas = append(msg.BuHuaDatas, int32(v))
						msg1.BuHuaDatas = append(msg1.BuHuaDatas, int32(v))
						continue
					} else {
						msg.BuHuaDatas = append(msg.BuHuaDatas, int32(v))
					}
					g.ud[i].HandCards[v] += 1
					g.ud[i].LastAddCard = v
					break
				}
			}
		}
		if len(msg.BuHuaDatas) > 0 {
			g.ud[i].Opt = 0
			g.ud[i].MoPai(0)
			if g.ud[i].Opt&majiangcom.OptTypeHu != 0 {
				g.ud[i].HuOptCard = int32(g.ud[i].LastAddCard)
			}
			g.ud[g.Zhuang].HandCards[0] = 0
			g.ud[i].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_BuHua), msg)
			if IsTest {
				g.ud[(i+1)%UserNum].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_BuHua), msg)
			} else {
				g.ud[(i+1)%UserNum].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_BuHua), msg1)
			}

		}
	}
	return temp
}

//玩家摸牌
func (g *Game) MoPai(ChairID int) {
	if g.TestUserIdx == ChairID {
		return
	}
	if g.TimerJob != nil {
		g.table.DeleteJob(g.TimerJob)
	}
	if !g.IsStartGame {
		return
	}
	if g.ud[ChairID].TestFlag == int32(TestStatus1) {
		return
	}
	temp := 0
	if g.ud[ChairID].TingPai {
		temp = g.TingMopai(ChairID)
	} else {
		temp = g.CtrlWater(g.GetCheatValue(int32(ChairID)), ChairID)
	}
	if temp == 0 {
		g.Settle(true)
		return
	}
	if temp < 0 {
		panic(fmt.Sprintln("===============mopaiValues=============", temp))
	}
	g.ud[ChairID].MoPai(temp)
	if g.MaJiang.GetLastCardsCount() == 0 {
		g.ud[ChairID].Opt &= ^(majiangcom.OptTypeAnGang | majiangcom.OptTypeGang | majiangcom.OptTypeMingGang)
	}

	g.ud[ChairID].LastAddCard = temp
	// 发送摸牌消息
	g.SendUserMoPaiMessage(ChairID, temp)
	Temptime := 0
	if temp > majiangcom.Bai[0] {
		Temptime = g.BuHua() // 补花
	}
	g.CurrOperatUser = ChairID
	g.CurrOperation = g.ud[ChairID].Opt
	g.TimerJob, _ = g.table.AddTimer(int64(Temptime), g.SendUserOptMsg)
}

func (g *Game) SendUserMoPaiMessage(currUser, cardValues int) {
	msg := new(errenmajiang.UserMoPaiMessage)
	msg.CardValues = int32(cardValues)
	msg.ChairId = int32(currUser)
	g.ud[currUser].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_MoPai), msg)
	//TODO 测试
	if !IsTest {
		msg.CardValues = -1
	}

	g.ud[(currUser+1)%UserNum].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_MoPai), msg)
}

func (g *Game) OnUserOpt(buffer []byte, user player.PlayerInterface) {
	msg := &errenmajiang.OptRetMsg{}
	proto.Unmarshal(buffer, msg)
	ChairID := user.GetChairID()
	if g.ud[ChairID].Opt == 0 {
		return
	}

	if g.ud[ChairID].Opt&int(msg.Opt) == 0 {
		if g.ud[ChairID].Opt&majiangcom.OptTypeHu != 0 {
			if g.TimerJob != nil {
				g.table.DeleteJob(g.TimerJob)
			}
			if int32(g.ud[ChairID].LastAddCard) == g.ud[ChairID].HuOptCard {
				g.CurrOperation = majiangcom.OptTypeOutCard
				g.ud[ChairID].Opt = majiangcom.OptTypeOutCard
				g.SendUserOptMsg()
			} else if int32(g.LastOutCard) == g.ud[ChairID].HuOptCard {
				//对家打牌
				g.MoPai(ChairID)
			} else {
				g.MoPai(g.CurrOperatUser)
			}
		} else if g.ud[ChairID].Opt&(majiangcom.OptTypeMingGang|majiangcom.OptTypeAnGang) != 0 {
			if g.TimerJob != nil {
				g.table.DeleteJob(g.TimerJob)
			}
			g.CurrOperation = majiangcom.OptTypeOutCard
			g.ud[ChairID].Opt = majiangcom.OptTypeOutCard
			g.SendUserOptMsg()
		} else if g.ud[ChairID].Opt&majiangcom.OptTypePeng != 0 || g.ud[ChairID].Opt&majiangcom.OptTypeGang != 0 ||
			g.ud[ChairID].Opt&(majiangcom.OptTypeZhongChi|majiangcom.OptTypeYouChi|majiangcom.OptTypeZuoChi) != 0 { // 取消碰 杠 吃
			//对家打牌
			if g.TimerJob != nil {
				g.table.DeleteJob(g.TimerJob)
			}
			g.MoPai(ChairID)
		}
	} else {
		if msg.Opt&int32(majiangcom.OptTypeZuoChi|majiangcom.OptTypeZhongChi|majiangcom.OptTypeYouChi) != 0 {
			g.ud[(ChairID+1)%UserNum].SubOutCard()
			g.OnUserChi(int(msg.Opt), user)
			msg.CardValue = int32(g.LastOutCard)
		} else if msg.Opt&int32(majiangcom.OptTypePeng) != 0 {
			g.ud[(ChairID+1)%UserNum].SubOutCard()
			g.OnUserPeng(int(msg.Opt), user)
			msg.CardValue = int32(g.LastOutCard)
		} else if msg.Opt&int32(majiangcom.OptTypeAnGang|majiangcom.OptTypeMingGang|majiangcom.OptTypeGang) != 0 {
			if msg.Opt&int32(majiangcom.OptTypeAnGang) == 0 {
				g.ud[(ChairID+1)%UserNum].SubOutCard()
			}

			g.OnUserGang(int(msg.Opt), int(msg.CardValue), user)
		} else if msg.Opt&int32(majiangcom.OptTypeHu) != 0 {
			g.CurrOperatUser = ChairID
			g.Settle(false)
			return
		} else if msg.Opt&int32(majiangcom.OptTypeOutCard) != 0 {
			g.OnUserOutCard(msg, user)
		}
	}
}

//玩家出牌
func (g *Game) OnUserOutCard(msg *errenmajiang.OptRetMsg, user player.PlayerInterface) {
	if user.GetChairID() != g.CurrOperatUser || g.ud[user.GetChairID()].HandCards[msg.CardValue] <= 0 {
		return
	}
	if msg.CardValue < 1 {
		log.Errorf("%d出牌的值小于1: %v", user.GetID(), msg)
	}
	if g.TimerJob != nil {
		g.table.DeleteJob(g.TimerJob)
	}
	g.LastOutCard = int(msg.CardValue)
	g.ud[user.GetChairID()].LastAddCard = 0
	g.AllOutCardMaps[msg.CardValue] += 1
	tmp := g.ReUpLineOutCardsMaps[g.CurrOperatUser]
	tmp[msg.CardValue] += 1
	g.ReUpLineOutCardsMaps[g.CurrOperatUser] = tmp
	g.ud[user.GetChairID()].HandCards[msg.CardValue] -= 1

	// 游戏轮次
	if g.CurrOperatUser != g.Zhuang {
		g.Round++
	}
	g.ud[g.CurrOperatUser].Opt = 0
	g.CurrOperation = 0
	g.ud[(g.CurrOperatUser+1)%UserNum].GetUserOpt(true, g.LastOutCard)
	if g.MaJiang.GetLastCardsCount() == 0 {
		g.ud[(g.CurrOperatUser+1)%UserNum].Opt &= ^(majiangcom.OptTypeAnGang | majiangcom.OptTypeGang | majiangcom.OptTypeMingGang)
	}
	g.ud[g.CurrOperatUser].SetOutCard(msg.CardValue)
	log.Debugf("用户发的index=%v", msg.CardIndex)
	g.BroadCastUserOpt(majiangcom.OptTypeOutCard, msg.CardValue, msg.CardIndex)
}

func (g *Game) OnUserTingCard(buffer []byte, user player.PlayerInterface) {
	if user.GetChairID() != g.CurrOperatUser {
		return
	}

	msg := &errenmajiang.TingCardMessage{}
	proto.Unmarshal(buffer, msg)
	if msg.TingCardValue != 0 && !g.ud[g.CurrOperatUser].TingPai {
		msg.TingUserIndex = int32(user.GetChairID())
		g.ud[user.GetChairID()].CtrlHuCards = msg.HuArray
		g.ud[g.CurrOperatUser].TingPai = true
		msg.HuArray = []int32{}
		g.table.Broadcast(int32(errenmajiang.ReMsgIDS2C_BroadTing), msg)
		if g.Round == 0 && len(g.ud[g.CurrOperatUser].OptCards) == 0 {
			g.ud[g.CurrOperatUser].TianTing = true
		}
		msg1 := new(errenmajiang.OptRetMsg)
		msg1.Opt = 1
		msg1.CardValue = msg.TingCardValue
		msg1.CardIndex = msg.CardIndex
		g.OnUserOutCard(msg1, user)
		g.ud[user.GetChairID()].GetTingOdds()
	}
}

//玩家吃牌操作
func (g *Game) OnUserChi(Opt int, user player.PlayerInterface) {
	ChairID := user.GetChairID()
	g.ud[ChairID].SetChi(g.LastOutCard, Opt)
	g.ud[ChairID].Opt = 1
	g.CurrOperation = 1

	if Opt == majiangcom.OptTypeZuoChi {
		g.AllOutCardMaps[g.LastOutCard-1] += 1
		g.AllOutCardMaps[g.LastOutCard-2] += 1
	} else if Opt == majiangcom.OptTypeZhongChi {
		g.AllOutCardMaps[g.LastOutCard-1] += 1
		g.AllOutCardMaps[g.LastOutCard+1] += 1
	} else {
		g.AllOutCardMaps[g.LastOutCard+1] += 1
		g.AllOutCardMaps[g.LastOutCard+2] += 1
	}

	tmp := g.ReUpLineOutCardsMaps[user.GetChairID()]
	tmp[g.LastOutCard] -= 1
	g.ReUpLineOutCardsMaps[user.GetChairID()] = tmp
	g.BroadCastUserOpt(Opt, int32(g.LastOutCard), -1)
	g.ud[ChairID].GetHuOdds()
}

//玩家碰牌操作
func (g *Game) OnUserPeng(Opt int, user player.PlayerInterface) {
	ChairID := user.GetChairID()
	g.CurrOperatUser = ChairID
	g.ud[ChairID].SetPeng(g.LastOutCard)
	g.ud[ChairID].Opt = 1
	g.CurrOperation = 1
	g.AllOutCardMaps[g.LastOutCard] += 2
	tmp := g.ReUpLineOutCardsMaps[user.GetChairID()]
	tmp[g.LastOutCard] -= 1
	g.ReUpLineOutCardsMaps[user.GetChairID()] = tmp
	g.BroadCastUserOpt(Opt, int32(g.LastOutCard), -1)
	g.ud[ChairID].GetHuOdds()
}

//玩家杠牌操作
func (g *Game) OnUserGang(Opt int, CardValue int, user player.PlayerInterface) {
	//验证是否能杠
	ChairID := user.GetChairID()
	if g.ud[ChairID].Opt&Opt == 0 {
		return
	}
	g.LastOutCard = CardValue
	if Opt == majiangcom.OptTypeGang {
		g.ud[ChairID].SetGangPai(CardValue)
		g.CurrOperation = 0
		g.AllOutCardMaps[g.LastOutCard] = 4
		tmp := g.ReUpLineOutCardsMaps[user.GetChairID()]
		tmp[g.LastOutCard] -= 1
		g.ReUpLineOutCardsMaps[user.GetChairID()] = tmp
	} else if Opt == majiangcom.OptTypeMingGang {
		if !g.ud[ChairID].SetMingGangPai(CardValue) {
			return
		}
		temp, huCard := majiangcom.CanHu(g.ud[(ChairID+1)%UserNum].HandCards, int32(CardValue)) // 抢杠胡
		if temp != 0 {
			g.ud[(ChairID+1)%UserNum].Opt = temp
			g.CurrOperation = temp
			g.ud[(ChairID+1)%UserNum].HuOptCard = huCard
		} else {
			g.CurrOperation = 0
		}
	} else if Opt == majiangcom.OptTypeAnGang {
		if g.ud[g.CurrOperatUser].HandCards[CardValue] != 4 {
			return
		}
		g.ud[ChairID].SetAnGangPai(CardValue)
		g.CurrOperation = 0
	}
	g.CurrOperatUser = ChairID

	g.ud[g.CurrOperatUser].Opt = 0
	g.BroadCastUserOpt(Opt, int32(CardValue), 0)
}

func (g *Game) SendUserOptMsg() {
	log.Debugf("桌号：%v，玩家ID%v,操作：%v", g.table.GetID(), g.CurrOperatUser, g.CurrOperation)
	if g.CurrOperation == 0 {
		g.MoPai(g.CurrOperatUser)
	} else {

		tempTime := int64(0)
		tempTime = int64(config.ErRenMaJiang.OutCardTime)
		if g.ud[g.CurrOperatUser].Opt == g.CurrOperation {
			g.ud[g.CurrOperatUser].SendOptMsg(tempTime)
			g.ud[(g.CurrOperatUser+1)%UserNum].SendWaitMsg(tempTime)
		} else {
			g.ud[(g.CurrOperatUser+1)%UserNum].SendOptMsg(tempTime)
			g.ud[g.CurrOperatUser].SendWaitMsg(tempTime)
		}

		outCardTime := tempTime
		if g.ud[g.CurrOperatUser].AutoStatus || (g.ud[g.CurrOperatUser].TingPai && g.ud[g.CurrOperatUser].Opt < majiangcom.OptTypeMingGang) {
			outCardTime = rand.Int63n(5000)
		}
		flag := outCardTime == tempTime
		if g.ud[g.CurrOperatUser].Opt&majiangcom.OptTypeHu != 0 {
			if !flag {
				g.Settle(false)
			} else {
				g.TimerJob, _ = g.table.AddTimer(int64(outCardTime), func() {
					g.Settle(false)
				})
			}
		} else if g.ud[g.CurrOperatUser].Opt&(majiangcom.OptTypeAnGang|majiangcom.OptTypeMingGang) != 0 {
			g.TimerJob, _ = g.table.AddTimer(int64(outCardTime), func() {
				if !g.IsStartGame {
					return
				}
				if flag {
					g.OutTimeAutoCheckStatus()
				}
				if g.CheckGangHu(g.ud[g.CurrOperatUser].Opt) {
					g.ud[g.CurrOperatUser].SendOptMsg(int64(config.ErRenMaJiang.OutCardTime))
				} else {
					g.AutoPalyCard()
				}
			})
		} else if g.ud[g.CurrOperatUser].Opt&0xE != 0 ||
			g.ud[g.CurrOperatUser].Opt&majiangcom.OptTypePeng != 0 ||
			g.ud[g.CurrOperatUser].Opt&majiangcom.OptTypeGang != 0 {
			g.TimerJob, _ = g.table.AddTimer(int64(outCardTime), func() {
				if !g.IsStartGame {
					return
				}
				if flag {
					g.OutTimeAutoCheckStatus()
				}
				if g.CheckGangHu(g.ud[g.CurrOperatUser].Opt) {
					g.ud[g.CurrOperatUser].SendOptMsg(int64(config.ErRenMaJiang.OutCardTime))
				} else {
					g.MoPai(g.CurrOperatUser)
				}
			})
		} else if g.ud[g.CurrOperatUser].Opt&majiangcom.OptTypeOutCard != 0 {
			log.Debugf("outCardTime =%v", outCardTime)
			g.TimerJob, _ = g.table.AddTimer(int64(outCardTime), func() {
				if !g.IsStartGame {
					return
				}
				if flag {
					g.OutTimeAutoCheckStatus()
				}

				g.AutoPalyCard()
			})
		}
	}
}

func (g *Game) BroadCastUserOpt(opt int, card int32, index int32) {
	if g.TimerJob != nil {
		g.table.DeleteJob(g.TimerJob)
	}
	msg := new(errenmajiang.BroadOptRetMsg)
	msg.Opt = int32(opt)
	msg.CardValue = card
	msg.OptUserIndex = int32(g.CurrOperatUser)
	msg.CardIndex = index
	OptMsg := new(errenmajiang.OptRetMsg)
	OptMsg.CardValue = card
	OptMsg.CardIndex = index
	OptMsg.Opt = int32(opt)
	g.ud[g.CurrOperatUser].OptCards = append(g.ud[g.CurrOperatUser].OptCards, OptMsg)
	g.table.Broadcast(int32(errenmajiang.ReMsgIDS2C_BroadOptRet), msg)
	//操作动画完成以后发下一步操作
	if opt == 1 {
		g.CurrOperatUser = (g.CurrOperatUser + 1) % UserNum
		g.CurrOperation = g.ud[g.CurrOperatUser].Opt
	}
	g.TimerJob, _ = g.table.AddTimer(600, g.SendUserOptMsg)
}

func (g *Game) CheckGangHu(opt int) bool {
	tempUserData := g.ud[g.CurrOperatUser]
	if opt&majiangcom.OptTypeMingGang != 0 {
		for _, val := range tempUserData.PengCards {
			if val == tempUserData.LastAddCard {
				tempUserData.HandCards[tempUserData.LastAddCard] -= 1
			}
		}
		for _, val := range majiangcom.GetHandCards(tempUserData.HandCards) {
			if val == int32(tempUserData.LastAddCard) && tempUserData.HandCards[val] == 4 {
				tempUserData.HandCards[tempUserData.LastAddCard] -= 4
			}
		}
	} else if opt&majiangcom.OptTypeGang != 0 {
		if tempUserData.HandCards[g.LastOutCard] == 3 {
			tempUserData.HandCards[g.LastOutCard] -= 3
		}
	} else if opt&majiangcom.OptTypeAnGang != 0 {
		for _, val := range tempUserData.HandCards {
			if val == 4 {
				tempUserData.HandCards[val] -= 4
			}
		}
	}
	tempOpt, _ := majiangcom.CanHu(tempUserData.HandCards, 0)
	if tempOpt&majiangcom.OptTypeHu != 0 {
		return true
	} else {
		return false
	}
}

//结算
func (g *Game) Settle(lieu bool) {
	defer func() {
		g.IsStartGame = false
	}()
	if g.TimerJob != nil {
		g.table.DeleteJob(g.TimerJob)
	}
	curr := g.CurrOperatUser
	next := (g.CurrOperatUser + 1) % UserNum

	Fan := ""
	HandCards1 := &errenmajiang.UserHandCards{CardValue: majiangcom.GetHandCards(g.ud[curr].HandCards)}
	HandCards2 := &errenmajiang.UserHandCards{CardValue: majiangcom.GetHandCards(g.ud[next].HandCards)}

	if lieu {
		msg := new(errenmajiang.LiuJuMessage)
		msg.IsLiuJu = lieu
		msg.HandCards = map[int32]*errenmajiang.UserHandCards{int32(curr): HandCards1, int32(next): HandCards2}
		for i := 0; i < UserNum; i++ {
			if g.ud[i].User != nil {
				msg.HeadStr = append(msg.HeadStr, g.ud[i].User.GetHead())
			} else {
				msg.HeadStr = append(msg.HeadStr, "")
			}
		}
		g.table.Broadcast(int32(errenmajiang.ReMsgIDS2C_LiuJuMsg), msg)

		for i := 0; i < UserNum; i++ {
			if g.ud[i].User != nil {
				g.table.KickOut(g.ud[i].User)
				g.ud[i].ResetData()
			}
		}
		g.table.EndGame()
		g.ReSetData()
		return
	}
	if int32(g.ud[g.CurrOperatUser].LastAddCard) != g.ud[g.CurrOperatUser].HuOptCard {
		g.ud[curr].HandCards[g.ud[curr].HuOptCard] += 1
		g.AllOutCardMaps[g.ud[curr].HuOptCard] -= 1
	} else {
		g.ud[g.CurrOperatUser].IsZiMo = true
	}

	if g.ud[g.CurrOperatUser].HuOptCard == 0 {
		g.ud[g.CurrOperatUser].HuOptCard = int32(g.ud[g.CurrOperatUser].LastAddCard)
	}
	// 底分
	diFen := int32(gjson.Parse(g.table.GetAdviceConfig()).Get("Bottom_Pouring").Int())

	// 番名对应番数
	huTypeMap := make(map[string]int32)
	// 总输赢分数
	totalNumber := int32(0)
	// 胡牌番形
	mask, noMask := GetCardsOdds(&g.ud[curr])
	// 判断是否天胡
	g.IsTianHu(&mask, &noMask)
	// 判断是否地胡
	g.IsDiHu(&mask, &noMask)
	// 自摸
	g.IsZiMo(&mask, &noMask)
	if g.ud[curr].GangPaiNumber+g.ud[curr].PengPaiNumber+g.ud[curr].ChiPaiNumber == 0 {
		mask |= def.MengQianQing
	}

	// 点炮
	//g.IsDianPao(&mask, &noMask)
	// 和绝张
	g.IsHuJueZhang(&mask, &noMask)
	// 全求人
	if mask&def.DanDiao != 0 {
		g.IsQuanQiuRen(&mask, &noMask)
	}
	// 海底捞月
	g.IsHaiDiLaoYue(&mask, &noMask)

	//上个玩家操作的是明杠，并且胡的就是上个玩家的牌为抢杠胡
	nextoptlen := len(g.ud[next].OptCards)
	if nextoptlen >= 1 && g.ud[next].OptCards[nextoptlen-1].CardValue == g.ud[curr].HuOptCard &&
		g.ud[next].OptCards[nextoptlen-1].Opt&majiangcom.OptTypeMingGang != 0 {
		mask |= def.QiangGangHu
		noMask |= def.HeJueZhang
	}

	if g.ud[curr].TingPai {
		mask |= def.TingPai
		if g.ud[curr].TianTing {
			mask |= def.TianTing
			noMask |= def.TingPai
		}
	}
	// 没有字一色和清一色就是混一色，字一色和清一色都不记番混一色
	if mask&def.ZiYiSe == 0 && mask&def.QingYiSe == 0 {
		mask |= def.HunYiSe
	}
	HuValues := mask & ^noMask
	// 总番数
	totalDouble := int32(0)
	Fanidex := 0
	for idx, val := range def.FanTypeArray {
		if HuValues&val != 0 {
			Fan += def.FanNameArray[idx] + "、"
			if val != def.SiGuiYi {
				huTypeMap[def.FanNameArray[idx]] = def.FanDoubleArray[idx]
				totalDouble += def.FanDoubleArray[idx]
			} else {
				huTypeMap[def.FanNameArray[idx]] = def.FanDoubleArray[idx] * int32(g.ud[g.CurrOperatUser].SiGuiYiNumber)
				totalDouble += def.FanDoubleArray[idx] * int32(g.ud[g.CurrOperatUser].SiGuiYiNumber)
			}

			if def.FanDoubleArray[idx] >= 64 {
				Fanidex = idx
			}
		}
	}

	if len(g.ud[curr].Hua) > 0 {
		huTypeMap[def.FanNameArray[0]] += int32(len(g.ud[curr].Hua))
		totalDouble += int32(len(g.ud[curr].Hua))
	}
	if totalDouble > config.ErRenMaJiang.MaxCardMult {
		totalDouble = config.ErRenMaJiang.MaxCardMult
	}
	totalNumber = totalDouble * diFen
	//玩家的钱少于要赢的钱时
	if g.ud[curr].User.GetScore() < int64(totalNumber) {
		totalNumber = int32(g.ud[curr].User.GetScore())
	}

	//当要输的钱少于要赢的钱时
	if g.ud[next].User.GetScore() < int64(totalNumber) {
		totalNumber = int32(g.ud[next].User.GetScore())
	}
	profitMoney1 := g.ud[curr].User.SetScore(g.table.GetGameNum(), int64(totalNumber), g.table.GetRoomRate())
	log.Debugf("税后 %v %v", profitMoney1, g.table.GetRoomRate())
	profitMoney2 := g.ud[next].User.SetScore(g.table.GetGameNum(), -int64(totalNumber), g.table.GetRoomRate())

	settleMes1 := new(errenmajiang.Settle)
	settleMes2 := new(errenmajiang.Settle)

	settleMes1.HandCards = map[int32]*errenmajiang.UserHandCards{int32(curr): HandCards1, int32(next): HandCards2}
	settleMes1.HuTypes = huTypeMap
	settleMes1.UserWiner = 1
	settleMes1.UserWinMoney = int32(profitMoney1)
	settleMes1.HuCard = g.ud[curr].HuOptCard
	for i := 0; i < UserNum; i++ {
		settleMes1.UserMoney = append(settleMes1.UserMoney, g.ud[i].User.GetScore())
		settleMes1.HeadStr = append(settleMes1.HeadStr, g.ud[i].User.GetHead())
		settleMes2.UserMoney = append(settleMes2.UserMoney, g.ud[i].User.GetScore())
		settleMes2.HeadStr = append(settleMes2.HeadStr, g.ud[i].User.GetHead())
	}

	g.ud[curr].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_SettleMsg), settleMes1)

	//fmt.Printf("===%v: settleMes: %v===\n", g.ud[curr].User.GetID(), settleMes1)

	settleMes2.HandCards = map[int32]*errenmajiang.UserHandCards{int32(curr): HandCards1, int32(next): HandCards2}
	settleMes2.HuTypes = huTypeMap
	settleMes2.UserWiner = 2
	settleMes2.UserWinMoney = -totalNumber
	settleMes2.HuCard = g.ud[curr].HuOptCard
	g.ud[next].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_SettleMsg), settleMes2)

	//战绩
	var records []*platform.PlayerRecord

	if !g.ud[curr].User.IsRobot() {
		records = append(records, &platform.PlayerRecord{
			PlayerID:     uint32(g.ud[curr].User.GetID()),
			GameNum:      g.table.GetGameNum(),
			ProfitAmount: profitMoney1,
			BetsAmount:   int64(diFen),
			DrawAmount:   int64(totalNumber) - profitMoney1,
			OutputAmount: profitMoney1,
			Balance:      g.ud[curr].User.GetScore(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
	}

	if !g.ud[next].User.IsRobot() {
		records = append(records, &platform.PlayerRecord{
			PlayerID:     uint32(g.ud[next].User.GetID()),
			GameNum:      g.table.GetGameNum(),
			ProfitAmount: profitMoney2,
			BetsAmount:   int64(totalNumber),
			DrawAmount:   0,
			OutputAmount: 0,
			Balance:      g.ud[next].User.GetScore(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
	}

	if len(records) > 0 {
		if _, err := g.table.UploadPlayerRecord(records); err != nil {
			log.Errorf("upload player record error %s", err.Error())
		}
	}

	g.WriteLog(Fan, totalNumber, int32(profitMoney1))

	//跑马灯
	g.PaoMaDeng(profitMoney1, Fanidex, g.ud[curr].User)
	for i := 0; i < UserNum; i++ {
		if g.ud[i].User != nil {
			g.table.KickOut(g.ud[i].User)
			g.ud[i].ResetData()
		}
	}
	g.table.EndGame()
	g.ReSetData()
	return
}

func (g *Game) WriteLog(Fan string, totalNumber int32, taxNumber int32) {
	op := "被点炮"
	next := (g.CurrOperatUser + 1) % UserNum
	if g.ud[g.CurrOperatUser].IsZiMo {
		op = "被自摸"
	}
	card1 := g.ud[g.CurrOperatUser].HandCards
	card2 := g.ud[next].HandCards

	for i := 0; i < g.ud[g.CurrOperatUser].PengPaiNumber; i++ {
		card1[g.ud[g.CurrOperatUser].PengCards[i]] += 3
	}
	for i := 0; i < g.ud[g.CurrOperatUser].GangPaiNumber; i++ {
		card1[g.ud[g.CurrOperatUser].GangPai[i]] += 4
	}

	for i := 0; i < g.ud[g.CurrOperatUser].AnGangPaiNumber; i++ {
		card1[g.ud[g.CurrOperatUser].AnGangPai[i]] += 4
	}

	for i := 0; i < g.ud[g.CurrOperatUser].ChiPaiNumber; i++ {
		card1[g.ud[g.CurrOperatUser].ChiPai[i]]++
	}
	for i := 0; i < g.ud[next].PengPaiNumber; i++ {
		card2[g.ud[next].PengCards[i]] += 3
	}
	for i := 0; i < g.ud[next].GangPaiNumber; i++ {
		card2[g.ud[next].GangPai[i]] += 4
	}

	for i := 0; i < g.ud[next].AnGangPaiNumber; i++ {
		card2[g.ud[next].AnGangPai[i]] += 4
	}

	for i := 0; i < g.ud[next].ChiPaiNumber; i++ {
		card2[g.ud[next].ChiPai[i]]++
	}

	g.SETTELCARDS[g.CurrOperatUser] = fmt.Sprintf("结算作弊率：%v 结算手牌：%v\t胡牌方式：%v\t输赢金额：%v\t 当前金额：%v\r\n", g.GetCheatValue(int32(g.CurrOperatUser)),
		majiangcom.GetHandCardString(card1), Fan, score.GetScoreStr(int64(taxNumber)), score.GetScoreStr(g.ud[g.CurrOperatUser].User.GetScore()))

	g.SETTELCARDS[next] = fmt.Sprintf("结算作弊率：%v 结算手牌：%v\t 胡牌方式：%v\t 输赢金额：%v\t 当前金额：%v\r\n", g.GetCheatValue(int32(next)),
		majiangcom.GetHandCardString(card2), op, score.GetScoreStr(int64(-totalNumber)), score.GetScoreStr(g.ud[next].User.GetScore()))

	diFen := int64(gjson.Parse(g.table.GetAdviceConfig()).Get("Bottom_Pouring").Int())
	diFenText := fmt.Sprintf("底分：%v\r\n", score.GetScoreStr(diFen))
	g.table.WriteLogs(g.ud[next].User.GetID(), diFenText+g.STARTCARDS[next]+g.SETTELCARDS[next])
	g.table.WriteLogs(g.ud[g.CurrOperatUser].User.GetID(), diFenText+g.STARTCARDS[g.CurrOperatUser]+g.SETTELCARDS[g.CurrOperatUser])
	// 打码量设置
	// 底分

	g.ud[g.CurrOperatUser].User.SendChip(diFen)
	g.ud[next].User.SendChip(int64(totalNumber))

	for i := 0; i < UserNum; i++ {
		if g.ud[i].User != nil && !g.ud[i].User.IsRobot() {
			g.table.KickOut(g.ud[i].User)
		}
	}
}

// 获取水位线
func (g *Game) GetCheatValue(userIndex int32) int {
	//先获取用户的
	Prob := g.ud[userIndex].User.GetProb()
	g.ud[userIndex].CheatSrc = "点控"
	if Prob == 0 {
		g.ud[userIndex].CheatSrc = "系统"
		tmp := g.table.GetRoomProb()
		log.Debugf("%v 获取到的系统作弊值为：%v", g.ud[userIndex].User.GetID(), tmp)
		Prob = tmp
	}

	if Prob == 0 {
		g.ud[userIndex].CheatSrc += " 获取到作弊率为0"
		Prob = 1000
	}

	log.Debugf("%v 获取到的系统作弊值为：%v", g.ud[userIndex].User.GetID(), Prob)
	return int(Prob)
}

// 控分放水
func (g *Game) CtrlWater(waterLine, ChairID int) int {
	switch g.MaJiang.GetLastCardsCount() {
	case 0:
		return 0
	case 1:
		return g.MaJiang.DealCard()
	}

	flag := false
	tempWeight := map[int]float64{}
	tempCard := [UserNum]int{}
	pointMap := map[string]float64{}

	if g.ud[ChairID].User.IsRobot() {
		pointMap = config.ErRenMaJiang.CtrlMoMapsRobot
	} else {
		pointMap = config.ErRenMaJiang.CtrlMoMapsUser
	}

	for i := 0; i < UserNum; i++ {
		tempCardMap := g.ud[ChairID].HandCards
		val := g.MaJiang.DealCard()
		tempCard[i] = val
		tempCardMap[val] += 1
		tempWeight[val] = majiangcom.GetCardsWeight(tempCardMap, 0, 0)
	}

	userWeight := pointMap[strconv.Itoa(waterLine)]
	randWeight := rand.Int63n(int64(config.ErRenMaJiang.CtrlMoMapsRobot[strconv.Itoa(waterLine)] + config.ErRenMaJiang.CtrlMoMapsUser[strconv.Itoa(waterLine)]))

	if float64(randWeight) < userWeight {
		flag = true
	}
	max := 0
	min := 0
	resVal := 0
	if tempWeight[tempCard[0]] > tempWeight[tempCard[1]] {
		max = tempCard[0]
		min = tempCard[1]
	} else {
		max = tempCard[1]
		min = tempCard[0]
	}
	if flag {
		resVal = max
	} else {
		resVal = min
	}
	g.MaJiang.ResaultCard(resVal)
	return resVal
}

func (g *Game) TingMopai(ChairID int) int {
	if !g.IsStartGame {
		return 0
	}
	tempPoint := g.MaJiang.GetLastCardsCount()
	if tempPoint == 0 {
		g.Settle(true)
		return 0
	}
	tempCtrlPointMaps := map[string]float64{}
	if g.ud[ChairID].User.IsRobot() {
		tempCtrlPointMaps = config.ErRenMaJiang.CtrlHuMapsRobot
	} else {
		tempCtrlPointMaps = config.ErRenMaJiang.CtrlHuMapsUser
	}
	Point := rand.Intn(tempPoint)
	CtrlPoint := tempCtrlPointMaps[strconv.Itoa(g.GetCheatValue(int32(ChairID)))]
	tempTotalCardArray := majiangcom.InitTestCardsInt(g.MaJiang.GetLastCardArray())
	HuNumber := 0

	tempArray := make([]int, len(g.ud[ChairID].CtrlHuCards))
	var index int
	for _, val := range g.ud[ChairID].CtrlHuCards {
		HuNumber += tempTotalCardArray[val]
		if tempTotalCardArray[val] > 0 {
			tempArray[index] = int(val)
			index++
		}
	}
	if len(g.ud[ChairID].CtrlHuCards) > 0 && float64(Point) <= float64(HuNumber/tempPoint)*CtrlPoint && index > 0 {
		CardValue := tempArray[rand.Intn(index)]
		ret := g.MaJiang.GetCardValue(CardValue)
		if ret != 0 {
			return ret
		}
	}
	return g.CtrlWater(g.GetCheatValue(int32(ChairID)), ChairID)
}

func (g *Game) ReSetData() {
	g.ud = [UserNum]UserData{}
	g.AllOutCardMaps = [majiangcom.MaxCardValue]int{}
	g.ReUpLineOutCardsMaps = map[int][majiangcom.MaxCardValue]int{}
	g.IsStartGame = false
	g.TestUserIdx = -1
	g.CurrOperation = 0
	g.SendSceneIds = [UserNum]*errenmajiang.RepeatedInt{}
	g.UserAvatar = [UserNum]string{}
	g.UserName = [UserNum]string{}
	g.UserMoney = [UserNum]int64{}
	g.UserCity = [UserNum]string{}

	g.table.Close()
}

func (g *Game) Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func (g *Game) HanCardsWeight(Cards [UserNum][]int32) [UserNum][]int32 {
	var tempCard [UserNum]float64
	var tempUser [UserNum]float64
	var resMaps [UserNum][]int32
	tempWeight := map[string]float64{}
	var MaxUserIndex int
	var MaxCardIndex int
	for i := 0; i < UserNum; i++ {
		if g.ud[i].User.IsRobot() {
			tempWeight = config.ErRenMaJiang.CtrlHandMapsRobot
		} else {
			tempWeight = config.ErRenMaJiang.CtrlHandMapsUser
		}
		tempCard[i] = majiangcom.GetCardsWeight(majiangcom.InitTestCards(Cards[i]), 0, 0)
		tempUser[i] = tempWeight[strconv.Itoa(g.GetCheatValue(int32(i)))]
	}
	MaxUserIndex = 0
	randNum := rand.Int63n(int64(tempUser[0] + tempUser[1]))
	RoundPoint := tempUser[0]
	if RoundPoint < tempUser[1] {
		RoundPoint = tempUser[1]
	}
	if float64(randNum+int64(RoundPoint/1000)*100) < tempUser[1] {
		MaxUserIndex = 1
	}
	MaxCardIndex = 0
	if tempCard[0] < tempCard[1] {
		MaxCardIndex = 1
	}
	resMaps[MaxUserIndex] = Cards[MaxCardIndex]
	resMaps[(MaxUserIndex+1)%UserNum] = Cards[(MaxCardIndex+1)%UserNum]
	return resMaps
}

func (g *Game) PaoMaDeng(Gold int64, Type int, user player.PlayerInterface) {
	configs := g.table.GetMarqueeConfig()
	for _, v := range configs {
		if Type != 0 {
			special, _ := strconv.Atoi(v.SpecialCondition)
			if special == 1 && Gold >= v.AmountLimit {
				err := g.table.CreateMarquee(user.GetNike(), Gold, def.FanNameArray[Type], v.RuleId)
				if err != nil {
					log.Debugf("创建跑马灯错误：%v", err)
				}
			}
		} else if Gold >= v.AmountLimit && len(v.SpecialCondition) == 0 {
			err := g.table.CreateMarquee(user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Debugf("创建跑马灯错误：%v", err)
			}
		}
	}
}
