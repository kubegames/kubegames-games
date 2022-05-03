package game

import (
	"fmt"

	"github.com/kubegames/kubegames-games/pkg/battle/960211/model"
	pai9 "github.com/kubegames/kubegames-games/pkg/battle/960211/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

func (game *Game) UserReady(user player.PlayerInterface) bool {
	if game.startJob == nil && game.status == pai9.GameStatus_StartStatus {
		// 满桌时间
		fullTableTime := 5
		// 满桌时间权重
		game.startJob, _ = game.table.AddTimer(int64(fullTableTime*1000), game.RobotSitCheck)
	}
	return true
}

func (game *Game) BindRobot(robot player.RobotInterface) player.RobotHandler {
	rb := NewRobot()
	rb.user = robot
	return rb
}

func (game *Game) CloseTable() {
	// game.userChairMap = make(map[int]*User, 0)
	// game.userIDMap = make(map[int64]*User, 0)
	// if game.job != nil {
	// 	game.table.DeleteJob(game.job)
	// 	game.job = nil
	// }
	// if game.showPokerMsg != nil {
	// 	game.showPokerMsg.Reset()
	// }
	// if game.dealPokerMsg != nil {
	// 	game.dealPokerMsg.Reset()
	// }

	// game.status = 0
	// game.zhuangChairID = 0
	// game.setNum = 0
	// game.isStart = false
	// game.qiangNum = 0
	// game.betNum = 0
}

func (game *Game) ResetTable() {
	game.userChairMap = make(map[int]*User, 0)
	game.userIDMap = make(map[int64]*User, 0)
	if game.job != nil {
		game.table.DeleteJob(game.job)
		game.job = nil
	}
	if game.showPokerMsg != nil {
		game.showPokerMsg.Reset()
	}
	if game.dealPokerMsg != nil {
		game.dealPokerMsg.Reset()
	}

	game.status = 0
	game.zhuangChairID = 0
	game.setNum = 0
	game.isStart = false
	game.qiangNum = 0
	game.betNum = 0
}

//用户坐下
func (game *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) table.MatchKind {
	game.getUser(user)
	log.Tracef("玩家 %d 进入房间 %d OnActionUserSitDown", user.GetID(), game.table.GetID())

	return table.SitDownOk
}

func (game *Game) UserOffline(user player.PlayerInterface) (iscan bool) {
	// 还在匹配过程中或一轮已经结束
	if game.status == pai9.GameStatus_StartStatus { //&& game.startJob != nil {
		_, chair := game.getUser(user)
		delete(game.userChairMap, chair)
		delete(game.userIDMap, user.GetID())
		if game.startJob != nil && len(game.userChairMap) == 0 {
			game.table.DeleteJob(game.job)
			game.startJob = nil
		}
		iscan = true
		return true
	}

	if game.status != (pai9.GameStatus_SettleAll) || game.isStart {
		iscan = false
		return false
	}
	_, chair := game.getUser(user)
	delete(game.userChairMap, chair)
	delete(game.userIDMap, user.GetID())
	iscan = true
	return true
}

func (game *Game) UserLeaveGame(user player.PlayerInterface) (iscan bool) {
	// 还在匹配过程中或一轮已经结束
	if game.status == pai9.GameStatus_StartStatus { //&& game.startJob != nil {
		_, chair := game.getUser(user)
		delete(game.userChairMap, chair)
		delete(game.userIDMap, user.GetID())
		if game.startJob != nil && len(game.userChairMap) == 0 {
			game.table.DeleteJob(game.startJob)
			game.startJob = nil
		}
		iscan = true
		return true
	}

	if game.status != (pai9.GameStatus_SettleAll) || game.isStart {
		iscan = false
		return false
	}
	_, chair := game.getUser(user)
	delete(game.userChairMap, chair)
	delete(game.userIDMap, user.GetID())
	iscan = true
	return true
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(pai9.ReceiveMessageType_QiangZhuangReq):
		game.qiangMsg(buffer, user)
	case int32(pai9.ReceiveMessageType_BetMultiReq):
		game.betMsg(buffer, user)
	case int32(pai9.ReceiveMessageType_TestReq):
		// game.handleTest(buffer, user)
	}
}

func (game *Game) SendScene(user player.PlayerInterface) {
	u, chair := game.getUser(user)
	scene := new(pai9.SenceMsg)
	scene.Status = int32(game.status)
	if game.job != nil {
		leftTime := int32(game.job.GetTimeDifference()) + 500
		leftTime = leftTime / 1000 * 1000
		scene.LeftTime = leftTime
	}

	scene.QiangMulti = game.multi.QiangMulti
	scene.BetMulti = u.BetMultiList
	if u == game.userChairMap[game.zhuangChairID] {
		scene.BetMulti = []int32{1, 2, 3, 4, 5}
	}
	scene.UserInfo = append(scene.UserInfo, u.GetInfo(int32(chair)))
	scene.ZhuangChairID = int32(game.zhuangChairID)
	if zhuangUser := game.userChairMap[game.zhuangChairID]; zhuangUser != nil {
		scene.ZhuangVal = zhuangUser.QiangMulti
	}
	scene.RoomID = int64(game.table.GetRoomID())
	scene.SetNum = int32(game.setNum)
	if scene.SetNum == 0 {
		scene.SetNum = 1
	}

	// 发送场景消息
	scene.Poker = game.showPokerMsg

	scene.Bottom = game.getBottom()
	scene.Nums = game.getLeftCards()
	scene.DealFirstChairID = game.dealPokerMsg.DealFirstChairID

	for chairid, v := range game.userChairMap {
		if v.HasQiang {
			scene.QiangInfo = append(scene.QiangInfo, &pai9.QiangZhuangRespMsg{
				ChairID: int32(chairid),
				Val:     v.QiangMulti,
			})
		}
		if v.HasBet {
			scene.BetInfo = append(scene.BetInfo, &pai9.BetMultiRespMsg{
				ChairID: int32(chairid),
				Val:     v.BetMulti,
			})
		}
	}

	scene.IntervalTime = scene.LeftTime

	user.SendMsg(int32(pai9.SendToClientMessageType_SceneResp), scene)
	if len(game.userChairMap) == TABLE_NUM {
		if game.startJob != nil {
			game.table.DeleteJob(game.startJob)
			game.startJob = nil
		}
		// 发送玩家列表
		msg := new(pai9.UserListRespMsg)
		for chair, u := range game.userChairMap {
			msg.List = append(msg.List, u.GetInfo(int32(chair)))
		}
		game.table.Broadcast(int32(pai9.SendToClientMessageType_UserListResp), msg)
	}

	return
}

func (game *Game) GameStart() {
	if len(game.userChairMap) == TABLE_NUM && game.status == pai9.GameStatus_StartStatus {
		// 初始化一副牌
		game.poker = model.Init(model.CardsAllType)
		game.Shuffle()
		return
	}
}

func (game *Game) RobotSitCheck() {
	// 倒计时结束，申请need个机器人
	need := TABLE_NUM - len(game.userChairMap)
	fmt.Printf("申请 %d 个机器人\n", need)
	if err := game.table.GetRobot(uint32(need), game.table.GetConfig().RobotMinBalance, game.table.GetConfig().RobotMaxBalance); err != nil {
		log.Errorf("申请机器人失败:", err)
	}
}
