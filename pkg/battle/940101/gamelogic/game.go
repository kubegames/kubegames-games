package gamelogic

import (
	"math/rand"

	"github.com/golang/protobuf/proto"
	majiangcom "github.com/kubegames/kubegames-games/internal/pkg/majiang"
	"github.com/kubegames/kubegames-games/pkg/battle/940101/config"
	errenmajiang "github.com/kubegames/kubegames-games/pkg/battle/940101/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
	"github.com/tidwall/gjson"
)

type LogicInterFace struct {
}

//初始化桌子
func (inter *LogicInterFace) InitTable(table table.TableInterface) {
	g := new(Game)
	g.InitTable(table)
	table.Start(g, nil, nil)
}

func (g *Game) InitTable(table table.TableInterface) {
	g.table = table

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
}

//用户坐下
func (g *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) table.MatchKind {
	g.ud[chairId].User = user
	return table.SitDownOk
}

func (g *Game) UserOffline(user player.PlayerInterface) bool {
	if g.IsStartGame {
		return false
	}
	return true
}

func (g *Game) UserLeaveGame(user player.PlayerInterface) bool {
	if g.IsStartGame {
		return false
	}
	return true
}

//游戏消息
func (g *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(errenmajiang.MsgIDC2S_OptRet):
		g.OnUserOpt(buffer, user)
	case int32(errenmajiang.MsgIDC2S_TingCard):
		g.OnUserTingCard(buffer, user)
	case int32(errenmajiang.MsgIDC2S_AutoMsg):
		g.Auto(buffer, user)
	case int32(errenmajiang.MsgIDC2S_TestCard):
		//g.TestCard(buffer, user)
	}
}

func (g *Game) Auto(buffer []byte, user player.PlayerInterface) {
	msg := new(errenmajiang.AutoOutCard)
	proto.Unmarshal(buffer, msg)
	g.ud[user.GetChairID()].AutoStatus = msg.Auto
}

func (g *Game) TestCard(buff []byte, user player.PlayerInterface) {
	msg := new(errenmajiang.TestCardMsg)
	if err := proto.Unmarshal(buff, msg); err != nil {
		log.Errorf("errenmajiang proto.Unmarshal TestCardMsg error: %v", err.Error())
	}

	g.ud[user.GetChairID()].TestFlag = msg.Flag
	g.TestUserIdx = user.GetChairID()
	if msg.Flag == int32(TestStatus0) {
		g.ud[user.GetChairID()].HandCards = majiangcom.InitTestCards(msg.CardValues)
		log.Tracef("%v 测试发牌 %v %v", g.ud[user.GetChairID()].User.GetID(), majiangcom.GetHandCardString(g.ud[user.GetChairID()].HandCards), msg.CardValues)
		log.Tracef("测试发牌 %v", majiangcom.GetHandCards(g.ud[user.GetChairID()].HandCards))
		msg1 := new(errenmajiang.DealCardMsg)
		msg1.HandCards = msg.CardValues
		count := len(msg.CardValues)
		if count != 0 {
			g.ud[user.GetChairID()].LastAddCard = int(msg.CardValues[count-1])
		}
		//发送发牌消息
		g.ud[g.TestUserIdx].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_DealCard), msg1)
		if g.TimerJob != nil {
			g.table.DeleteJob(g.TimerJob)
			g.TimerJob = nil
		}

		g.GameStart()
	} else if msg.Flag == int32(TestStatus1) {
		if g.CurrOperatUser != user.GetChairID() {
			return
		}
		msg1 := new(errenmajiang.UserMoPaiMessage)
		//msg1.OldCard = int32(g.ud[user.GetChairID()].LastAddCard)

		//g.ud[user.GetChairID()].HandCards[g.ud[user.GetChairID()].LastAddCard] -= 1
		g.MaJiang.DealCard()
		g.ud[user.GetChairID()].MoPai(int(msg.CardVaalue))
		g.ud[user.GetChairID()].LastAddCard = int(msg.CardVaalue)
		// 发送摸牌消息
		//g.SendUserMoPaiMessage(user.GetChairID(), int(msg.CardVaalue))
		msg1.CardValues = msg.CardVaalue
		g.SendUserMoPaiMessage(user.GetChairID(), int(msg.CardVaalue))
		if int(msg.CardVaalue) > majiangcom.Bai[0] {
			g.BuHua() // 补花
		}
		g.CurrOperation = g.ud[user.GetChairID()].Opt
		g.SendUserOptMsg()
	}
}

func (g *Game) UserReady(user player.PlayerInterface) bool {
	return true
}

//BindRobot 绑定机器人
func (g *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	robot := new(Robot)
	robot.Init(ai, g)
	return robot
}

//场景消息
func (g *Game) SendScene(user player.PlayerInterface) {
	g.ud[user.GetChairID()].User = user

	index := user.GetChairID()
	g.SendSceneIds[index] = &errenmajiang.RepeatedInt{UserInfo: []int64{int64(user.GetChairID()), user.GetID()}}
	g.UserName[index] = user.GetNike()
	g.UserMoney[index] = user.GetScore()
	g.UserAvatar[index] = user.GetHead()
	g.UserCity[index] = user.GetCity()

	UserCount := 0
	for i := 0; i < UserNum; i++ {
		if g.ud[i].User != nil {
			UserCount++
		}
	}
	if UserCount < UserNum {
		//多久申请机器人
		applyTime := rand.Int63n(config.ErRenMaJiang.RobotTime[1]-config.ErRenMaJiang.RobotTime[0]) + config.ErRenMaJiang.RobotTime[0]
		g.TimerJob, _ = g.table.AddTimer(int64(applyTime), g.GetRobot)
	}
	// 发送场景消息
	g.SceneMessageBiu(user)
	// 发送用户上线消息
	if g.IsStartGame {
		g.SendUpLineMessageBiu(user)
		if g.ud[index].TingPaiMsg != nil && len(g.ud[index].TingPaiMsg.Cards) > 0 {
			g.ud[index].User.SendMsg(int32(errenmajiang.ReMsgIDS2C_NoticeTing), g.ud[index].TingPaiMsg)
		}
	}
	return
}

func (g *Game) CheckReLine(user player.PlayerInterface) bool {
	tmpUser := g.ud[user.GetChairID()].User
	if tmpUser != nil && tmpUser.GetID() == user.GetID() {
		return true
	}
	return false
}

func (g *Game) SceneMessageBiu(user player.PlayerInterface) {
	SceneMsg := new(errenmajiang.SceneMessage)
	SceneMsg.UserInfos = g.SendSceneIds[:]
	SceneMsg.UserAvatar = g.UserAvatar[:]
	SceneMsg.UserName = g.UserName[:]
	SceneMsg.UserMoney = g.UserMoney[:]
	SceneMsg.UserCity = g.UserCity[:]
	SceneMsg.RoomId = int64(g.table.GetRoomID())
	SceneMsg.IsTest = IsTest
	SceneMsg.EntranceRestrictions = g.table.GetEntranceRestrictions()
	g.table.Broadcast(int32(errenmajiang.ReMsgIDS2C_SenceID), SceneMsg)
}

func (g *Game) SendUpLineMessageBiu(user player.PlayerInterface) {
	ChairId := user.GetChairID()
	g.ud[ChairId].AutoStatus = false
	g.ud[ChairId].ReLineFlag = true
	UpLineMsg := new(errenmajiang.StartLineMessage)
	UpLineMsg.OutCards = map[int32]*errenmajiang.UserOutCard{}
	UpLineMsg.OptCards = map[int32]*errenmajiang.OptCard{}
	UpLineMsg.HuaCards = map[int32]*errenmajiang.HuaCards{}

	UpLineMsg.LastOutCard = int32(g.LastOutCard)
	UpLineMsg.HandCards = majiangcom.GetHandCards(g.ud[ChairId].HandCards)
	UpLineMsg.ZhuangIndex = int32(g.Zhuang)
	UpLineMsg.HomeCardNum = int32(len(majiangcom.GetHandCards(g.ud[(ChairId+1)%UserNum].HandCards)))
	UpLineMsg.LeftOverCardsNum = int32(g.MaJiang.GetLastCardsCount())
	UpLineMsg.TouZi = g.TouZi
	UpLineMsg.TingFlag = g.ud[user.GetChairID()].TingPai
	UpLineMsg.RoomBaseBet = int32(gjson.Parse(g.table.GetAdviceConfig()).Get("Bottom_Pouring").Int())
	UpLineMsg.RoomBaseName = g.table.GetLevel()

	for i := 0; i < UserNum; i++ {
		UpLineMsg.OutCards[int32(i)] = &errenmajiang.UserOutCard{OutCards: g.ud[i].OutCards[0:g.ud[i].OutCardsIndex]}
	}

	UpLineMsg.HuaCards = map[int32]*errenmajiang.HuaCards{int32(0): {Cards: g.ud[0].Hua}, int32(1): {Cards: g.ud[1].Hua}}
	UpLineMsg.OptCards = map[int32]*errenmajiang.OptCard{int32(0): {Opts: g.ud[0].OptCards}, int32(1): {Opts: g.ud[1].OptCards}}

	user.SendMsg(int32(errenmajiang.ReMsgIDS2C_UpLineMsg), UpLineMsg)

	t := g.TimerJob.GetTimeDifference()
	if g.CurrOperation != 0 {
		if g.CurrOperatUser == user.GetChairID() {
			g.ud[g.CurrOperatUser].SendOptMsg(t)
		} else {
			g.ud[user.GetChairID()].SendWaitMsg(t)
		}
	}
}

func (g *Game) GameStart() {
	UserCount := 0
	for i := 0; i < UserNum; i++ {
		if g.ud[i].User != nil {
			if g.ud[i].ReLineFlag {
				g.ud[g.CurrOperatUser].ReLineFlag = false
				return
			}
			UserCount++
		}
	}
	if UserCount == UserNum {
		//发骰子
		g.table.StartGame()
		g.SetZhuang()
	}
	return
}

func (g *Game) GetRobot() {
	if g.IsStartGame {
		return
	}

	if err := g.table.GetRobot(1, g.table.GetConfig().RobotMinBalance, g.table.GetConfig().RobotMaxBalance); err != nil {
		log.Errorf("生成机器人失败：%v", err)
	}
}

func (g *Game) ResetTable() {
	g.InitTable(g.table)
}
