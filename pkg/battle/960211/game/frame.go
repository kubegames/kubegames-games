package game

import (
	"common/log"
	"fmt"
	"game_frame_v2/define"
	"game_poker/pai9/model"
	pai9 "game_poker/pai9/msg"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/player"
)

func (game *Game) UserReady(user player.PlayerInterface) bool {

	userID := user.GetId()
	fmt.Printf("玩家 %d 在房间 %d 准备，游戏状态为 %d\n", userID, game.table.GetId(), game.status)

	//game.UserList[userID].Status = int32(msg.UserStatus_UserNormal)
	// 第一个玩家进入，预加载机器人
	fmt.Println("用户准备接口   game.startJob = ", game.startJob)
	if game.startJob == nil && game.status == pai9.GameStatus_StartStatus {
		fmt.Println("开启匹配定时器")
		// 满桌时间
		fullTableTime := 5
		// 满桌时间权重
		game.startJob, _ = game.table.AddTimer(time.Duration(fullTableTime*1000), game.RobotSitCheck)
	}

	return true
}

func (game *Game) CloseTable() {
	fmt.Println("关闭桌子")
	game.userChairMap = make(map[int]*User, 0)
	game.userIDMap = make(map[int64]*User, 0)
	if game.job != nil {
		game.job.Cancel()
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
	// game.BetMultiList = nil
}

func (game *Game) ResetTable() {
	fmt.Println("重置桌子")
	game.userChairMap = make(map[int]*User, 0)
	game.userIDMap = make(map[int64]*User, 0)
	if game.job != nil {
		game.job.Cancel()
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
	// game.BetMultiList = nil
}

//用户坐下
func (game *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) int {
	game.getUser(user)
	log.Tracef("玩家 %d 进入房间 %d OnActionUserSitDown", user.GetId(), game.table.GetId())

	return define.SIT_DOWN_OK //business.OnActionUserSitDownHandler()
}

func (game *Game) UserExit(user player.PlayerInterface) (iscan bool) {

	defer func() {
		fmt.Printf("用户【%d】退出   ******** 能否离开 %v UserExit\n", user.GetId(), iscan)
	}()

	// 还在匹配过程中或一轮已经结束
	if game.status == pai9.GameStatus_StartStatus { //&& game.startJob != nil {
		_, chair := game.getUser(user)
		delete(game.userChairMap, chair)
		delete(game.userIDMap, user.GetId())
		if game.startJob != nil && len(game.userChairMap) == 0 {
			fmt.Println("将startJob设置为空UserExit")
			game.startJob.Cancel()
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
	delete(game.userIDMap, user.GetId())
	iscan = true
	return true
}

func (game *Game) LeaveGame(user player.PlayerInterface) (iscan bool) {

	defer func() {
		fmt.Printf("用户【%d】退出   ******** 能否离开 %v LeaveGame\n", user.GetId(), iscan)
	}()

	// 还在匹配过程中或一轮已经结束
	if game.status == pai9.GameStatus_StartStatus { //&& game.startJob != nil {
		_, chair := game.getUser(user)
		delete(game.userChairMap, chair)
		delete(game.userIDMap, user.GetId())
		if game.startJob != nil && len(game.userChairMap) == 0 {
			fmt.Println("将startJob设置为空LeaveGame")
			game.startJob.Cancel()
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
	delete(game.userIDMap, user.GetId())
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

func (game *Game) SendScene(user player.PlayerInterface) bool {
	u, chair := game.getUser(user)
	scene := new(pai9.SenceMsg)
	scene.Status = int32(game.status)
	if game.job != nil {
		leftTime := int32(game.job.GetTimeDifference()) + 500
		leftTime = leftTime / 1000 * 1000
		scene.LeftTime = leftTime
		fmt.Println("leftTime = ", leftTime)
	}

	if !user.IsRobot() {
		fmt.Println("scene.LeftTime = ", scene.LeftTime)
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
	scene.RoomID = game.table.GetRoomID()
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

	if !user.IsRobot() {
		fmt.Println("scene  = ", scene)
	}

	user.SendMsg(int32(pai9.SendToClientMessageType_SceneResp), scene)

	if len(game.userChairMap) == TABLE_NUM {
		if game.startJob != nil {
			fmt.Println("将startJob设置为空scene")
			game.startJob.Cancel()
			game.startJob = nil
			// game.isStart = true
		}
		// 发送玩家列表
		msg := new(pai9.UserListRespMsg)
		for chair, u := range game.userChairMap {
			msg.List = append(msg.List, u.GetInfo(int32(chair)))
		}
		game.table.Broadcast(int32(pai9.SendToClientMessageType_UserListResp), msg)
	}

	return true
}

func (game *Game) GameStart(user player.PlayerInterface) bool {
	if len(game.userChairMap) == TABLE_NUM && game.status == pai9.GameStatus_StartStatus {
		// 初始化一副牌
		game.poker = model.Init(model.CardsAllType)
		game.Shuffle()
		return true
	}

	if len(game.userChairMap) == TABLE_NUM && game.status != pai9.GameStatus_StartStatus {
		return true
	}

	return false
}

func (game *Game) RobotSitCheck() {
	// 倒计时结束，申请need个机器人
	need := TABLE_NUM - len(game.userChairMap)
	fmt.Printf("申请 %d 个机器人\n", need)
	if err := game.table.GetRobot(int32(need)); err != nil {
		log.Errorf("申请机器人失败:", err)
	}
}
