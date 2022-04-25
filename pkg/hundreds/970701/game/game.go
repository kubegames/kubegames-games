package game

import (
	"encoding/json"
	"fmt"
	"go-game-sdk/define"
	"go-game-sdk/example/game_poker/saima/config"
	"go-game-sdk/example/game_poker/saima/model"
	"go-game-sdk/example/game_poker/saima/msg"
	. "go-game-sdk/inter"
	"go-game-sdk/sdk/global"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
)

var (
	gameStartTime  = 2000
	betTime        = 15000
	betNoticeTime  = 300
	betEndTime     = 3000
	gameTime       = 15000
	countTime      = 5000
	FirstPay       = 7
	FAndSPay       = 27
	FirstFemalePay = 3.8
	FirstMalePay   = 1.27
)

type Game struct {
	Table        TableInter            // table interface
	AllUserList  map[int64]*model.User //所有的玩家列表
	Start        bool
	Status       msg.GameStatus        // 房间状态1 表示
	T            int64                 // 状态切换时间
	GameInfo     []*msg.GameInfo       //历史记录
	BetList      []int64               //下注列表
	BetInfo      map[msg.BetArea]int64 //下注列表
	Result       []int32               //当前结果
	TempGameInfo *msg.GameInfo
	AllBet       int64     //真实玩家总下注
	AllResult    [][]int32 // 所有可能结果
	ControlKey   int32     //本局控制率
}

func (game *Game) Init(table TableInter) {
	game.Table = table
	game.T = time.Now().UnixNano() / 1e6
	gameStartTime = config.GetGameStartTime()
	betTime = config.GetBetTime()
	betNoticeTime = config.GetBetNoticeTime()
	betEndTime = config.GetBetEndTime()
	gameTime = config.GetGameTime()
	countTime = config.GetCountTime()
	FirstPay = config.GetFirstPay()
	FAndSPay = config.GetFAndSPay()
	FirstFemalePay = config.GetFirstFemalePay()
	FirstMalePay = config.GetFirstMalePay()
	game.reset()
}

func (game *Game) reset() {
	game.AllUserList = make(map[int64]*model.User)
	game.Status = msg.GameStatus_game_End
	game.Start = false
	game.GameInfo = make([]*msg.GameInfo, 0)
	game.BetInfo = make(map[msg.BetArea]int64, 38)
	game.AllResult = make([][]int32, 0)
}

func (game *Game) UserReady(user UserInetr) bool {
	return true
}

//用户坐下
func (game *Game) OnActionUserSitDown(user UserInetr, chairId int, config string) int {
	_, ok := game.AllUserList[user.GetID()]
	if !ok && user.GetScore() < game.Table.GetEntranceRestrictions() {
		return define.SIT_DOWN_ERROR_OVER
	}
	if !ok && len(game.AllUserList) >= int(global.GConfig.MaxPeople) {
		return define.SIT_DOWN_ERROR_NORMAL
	}
	if !ok {
		u := &model.User{
			Table: game.Table,
			User:  user,
			UserInfo: &msg.UserInfo{
				UserId:   user.GetID(),
				UserName: user.GetNike(),
				Head:     user.GetHead(),
				Amount:   user.GetScore(),
			},
		}
		u.ResetData()
		game.AllUserList[user.GetID()] = u
	}

	return define.SIT_DOWN_OK
}

func (game *Game) UserExit(user UserInetr) bool {
	return game.userLeave(user)
}

func (game *Game) LeaveGame(user UserInetr) bool {
	return game.userLeave(user)
}

func (game *Game) CloseTable() {
	for _, v := range game.AllUserList {
		game.Table.KickOut(v.User)
	}
}

func (game *Game) ResetTable() {
	game.reset()
}

func (game *Game) GameStart(user UserInetr) bool {
	if !game.Start {
		game.Start = true
		game.BetList = config.GetBetList(int(game.Table.GetLevel()))
		game.start()
		game.Table.AddTimerRepeat(int64(betNoticeTime), 0, func() {
			if game.Status == msg.GameStatus_game_Bet {
				game.gameBetNotice()
			}
		})
	}
	return true
}

func (game *Game) userLeave(user UserInetr) bool {
	u := game.AllUserList[user.GetID()]
	if u != nil && u.AllBet != 0 {
		return false
	}
	delete(game.AllUserList, user.GetID())
	return true
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user UserInetr) {
	switch subCmd {
	case int32(msg.MsgId_UserList_Req):
		game.getUserInfoList(user)
		break
	case int32(msg.MsgId_GameInfo_Req):
		game.getGameInfo(user)
		break
	case int32(msg.MsgId_ReBet_Req):
		game.rebet(buffer, user)
		break
	case int32(msg.MsgId_Bet_Req):
		game.bet(buffer, user)
		break
	case int32(msg.MsgId_Bet_Clear_Req):
		game.betClear(user)
		break
	}
}

func (game *Game) BindRobot(ai AIUserInter) RobotInter {
	robot := new(Robot)
	robot.Init(ai, game.Table)
	return robot
}

//场景消息
func (game *Game) SendScene(user UserInetr) bool {
	log.Tracef("send scene =", user.GetID(), user.IsRobot())

	res := &msg.EnterRoomRes{}
	u := game.AllUserList[user.GetID()]
	tableInfo := &msg.TableInfo{
		GameStatus: game.Status,
		Gameinfo:   game.GameInfo,
		BetInfo:    u.GetBetInfo(),
		BetList:    game.BetList,
		RankIndexs: game.Result,
		AllBetInfo: game.transMapToArray(game.BetInfo),
		T:          int64(game.getWaitTime()) - time.Now().UnixNano()/1e6 + game.T,
	}

	tableInfo.UserInfoArr = game.getUserInfo()
	res.TableInfo = tableInfo
	user.SendMsg(int32(msg.MsgId_INTO_ROOM_Res), res)

	return true
}

//获取玩家信息
func (game *Game) getUserInfo() []*msg.UserInfo {
	userInfoArr := make([]*msg.UserInfo, 0)
	for _, v := range game.AllUserList {
		userInfoArr = append(userInfoArr, v.UserInfo)
	}
	return userInfoArr
}

//获取玩家信息列表
func (game *Game) getUserInfoList(user UserInetr) {
	res := &msg.UserListRes{}
	res.UserInfoArr = game.getUserInfo()
	user.SendMsg(int32(msg.MsgId_UserList_Res), res)
}

//获取历史记录
func (game *Game) getGameInfo(user UserInetr) {
	res := &msg.GameInfoRes{}
	res.Gameinfo = game.GameInfo
	user.SendMsg(int32(msg.MsgId_GameInfo_Res), res)
}

//重复下注
func (game *Game) rebet(buffer []byte, user UserInetr) {
	req := &msg.ReBetReq{}
	proto.Unmarshal(buffer, req)
	u := game.AllUserList[user.GetID()]
	if u == nil || u.IsRebet {
		log.Tracef("用户不存在或者已重复下注")
		game.sendBetFailed(user, "用户不存在或者已重复下注")
		return
	}
	score := int64(0)
	for _, v := range req.GetCountInfo() {
		if v.GetBetArea() < msg.BetArea_champion_1 || v.GetBetArea() > msg.BetArea_champion_Woman ||
			v.GetScore() <= 0 {
			continue
		}
		score += v.GetScore()
		u.BetInfo[v.GetBetArea()] = v.GetScore()
	}
	if score <= 0 || score > user.GetScore() ||
		game.AllBet+score > int64(config.GetAllBetLimit(int(game.Table.GetLevel()))) {
		log.Tracef("用户金币不足或者超过最大下注上限, bet = %v, user score = %v, allBet = %v", score, user.GetScore(), game.AllBet)
		u.BetInfo = make(map[msg.BetArea]int64, 38)
		return
	}
	for _, v := range req.GetCountInfo() {
		game.BetInfo[v.GetBetArea()] += v.GetScore()
	}
	u.IsRebet = true
	u.AllBet += score
	if !user.IsRobot() {
		game.AllBet += score
	}
	u.UserInfo.Amount -= score
	user.SetScore(game.Table.GetGameNum(), -score, game.Table.GetRoomRate())
	user.SendMsg(int32(msg.MsgId_ReBet_Res), &msg.ReBetRes{
		UserId:    req.GetUserId(),
		CountInfo: req.GetCountInfo(),
	})
}

//下注
func (game *Game) bet(buffer []byte, user UserInetr) {
	if game.Status != msg.GameStatus_game_Bet {
		log.Tracef("不在下注阶段, game statut = %v", game.Status)
		game.sendBetFailed(user, "不在下注阶段")
		return
	}
	req := &msg.BetReq{}
	proto.Unmarshal(buffer, req)
	u := game.AllUserList[user.GetID()]
	if u == nil || req.GetIndex() > int32(len(game.BetList)) || req.GetIndex() == 0 || req.GetBetArea() == msg.BetArea_invild {
		log.Tracef("找不到玩家或者下注不合法")
		game.sendBetFailed(user, "找不到玩家或者下注不合法")
		return
	}
	score := game.BetList[req.GetIndex()-1]
	if score > user.GetScore() || u.AllBet+score > int64(config.GetBetLimit(int(game.Table.GetLevel()))) ||
		game.AllBet+score > int64(config.GetAllBetLimit(int(game.Table.GetLevel()))) {
		log.Tracef("用户金币不足或者超过最大下注, bet = %v, user score = %v, allUserBet = %v, allBet = %v", score, user.GetScore(), u.AllBet, game.AllBet)
		game.sendBetFailed(user, "用户金币不足或者超过最大下注")
		return
	}
	u.AllBet += score
	u.IsRebet = true
	if !user.IsRobot() {
		game.AllBet += score
	}
	u.UserInfo.Amount -= score
	u.BetInfo[req.GetBetArea()] += score
	game.BetInfo[req.GetBetArea()] += score
	//u.BetIndexsInfo[req.GetBetArea()] = append(u.BetIndexsInfo[req.GetBetArea()], req.GetIndex())
	user.SetScore(game.Table.GetGameNum(), -score, game.Table.GetRoomRate())
	user.SendMsg(int32(msg.MsgId_Bet_Res), &msg.BetRes{
		UserId:  user.GetID(),
		Index:   req.GetIndex(),
		BetArea: req.GetBetArea(),
	})
}

//下注
func (game *Game) betClear(user UserInetr) {
	u := game.AllUserList[user.GetID()]
	if u == nil {
		log.Tracef("用户不存在")
		return
	}
	for k, v := range u.BetInfo {
		game.BetInfo[k] -= v
		game.AllBet -= v
	}
	user.SetScore(game.Table.GetGameNum(), u.AllBet, 0)
	u.BetClear()
	user.SendMsg(int32(msg.MsgId_Bet_Clear_Res), &msg.BetClearRes{})
}

//下注失败
func (game *Game) sendBetFailed(user UserInetr, reason string) {
	user.SendMsg(int32(msg.MsgId_Bet_Fail_Res), &msg.BetFailedRes{
		UserId: user.GetID(),
		Reason: reason,
	})
}

//开始游戏
func (game *Game) start() {
	log.Tracef("game start")
	game.Table.StartGame()
	game.Status = msg.GameStatus_game_Start
	game.Table.AddTimer(int64(gameStartTime), game.gameBet)
	game.SyncUserInfo()
	game.sendStatusChangeMsg()
}

//同步玩家信息
func (game *Game) SyncUserInfo() {
	res := &msg.SyncUserInfoRes{}
	for _, v := range game.AllUserList {
		if v != nil && v.User != nil {
			res.Amount = v.User.GetScore()
			v.User.SendMsg(int32(msg.MsgId_Sync_Res), res)
		}
	}
}

//发送下注
func (game *Game) gameBetNotice() {
	res := &msg.BetNoticeRes{
		BetInfo: game.transMapToArray(game.BetInfo),
	}
	game.Table.Broadcast(int32(msg.MsgId_Bet_Notice_Res), res)
}

//mapToArray
func (game *Game) transMapToArray(data map[msg.BetArea]int64) []*msg.BetInfo {
	betInfo := make([]*msg.BetInfo, 0)
	for k, v := range data {
		data := &msg.BetInfo{}
		data.BetArea = k
		data.Score = v
		betInfo = append(betInfo, data)
	}
	return betInfo
}

//开始下注
func (game *Game) gameBet() {
	game.Status = msg.GameStatus_game_Bet
	game.Table.AddTimer(int64(betTime), game.betEnd)
	game.sendStatusChangeMsg()
}

//下注结束
func (game *Game) betEnd() {
	game.Status = msg.GameStatus_game_Bet_End
	game.Table.AddTimer(int64(betEndTime), game.gameRunning)
	game.sendStatusChangeMsg()
}

//游戏
func (game *Game) gameRunning() {
	//t := gameTime / 3
	//for i := 0; i < 3; i++ {
	//	game.Table.AddTimer(int64(t * i), game.orderNotice)
	//}
	game.Status = msg.GameStatus_game_Count
	game.orderNotice()
	game.sendStatusChangeMsg()
	game.Table.AddTimer(int64(gameTime), game.count)
}

//发送顺序
func (game *Game) orderNotice() {
	result := make([]int32, 0)
	//for i := 0; i < 3; i++ {
	game.Result = game.getResult()
	result = append(result, game.Result...)
	//}
	res := &msg.OrderNoticeRes{
		RankIndexs: result,
		TimeList:   game.getTimeList(),
	}
	game.Table.Broadcast(int32(msg.MsgId_Order_Notice_Res), res)

}

//结算
func (game *Game) count() {
	//result := game.getResult()
	game.pay(game.Result)
	game.Table.AddTimer(int64(countTime), game.saveScore)
	res := &msg.CountRes{}
	for _, v := range game.AllUserList {
		res.CountInfo = v.GetCountInfo()
		v.User.SendMsg(int32(msg.MsgId_Count_Res), res)
	}
}

//获取结果
func (game *Game) getResult() []int32 {
	result := game.getControlResult()
	gameInfo := &msg.GameInfo{}
	gameInfo.FirstIndex = result[0]
	gameInfo.SecondIndex = result[1]
	game.TempGameInfo = gameInfo
	log.Tracef("result =", result)
	return result
}

//获取随机结果
func (game *Game) GetRandResult() []int32 {
	result := []int32{1, 2, 3, 4, 5, 6, 7, 8}
	rand.Shuffle(8, func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}

//获取结果
func (game *Game) getControlResult() []int32 {
	result := make([]int32, 0)
	all := make([][]int32, 0)
	first := int32(0)
	second := int32(0)
	min, controlLossRation := game.getControlLossRation()
	log.Tracef("controlLossRation = ", controlLossRation, min)
	if len(game.AllResult) == 0 {
		game.AllResult = game.GetAllResult()
	}
	for _, v := range game.AllResult {
		game.pay(v)
		lossRation := game.getLossRation()
		if lossRation >= min && lossRation < controlLossRation {
			all = append(all, v)
			log.Tracef("lossRation =", lossRation)
			//break
		}
	}
	length := len(all)
	if length > 0 {
		v := all[rand.Intn(length)]
		first = v[0]
		second = v[1]
		result = append(result, v...)
	}
	randResult := game.GetRandResult()
	for _, v := range randResult {
		if v != first && v != second {
			result = append(result, v)
		}
	}
	return result
}

//获取赔付率
func (game *Game) getControlLossRation() (int64, int64) {
	controlKey := game.Table.GetRoomProb()
	game.ControlKey = controlKey
	if controlKey == 3000 {
		return 0, 10000
	}
	return config.GetXueChiChance(controlKey)
}

//获取赔付率
func (game *Game) getLossRation() int64 {
	if game.AllBet == 0 {
		return 0
	}
	return game.getAllPay() * 10000 / game.AllBet
}

//获取所有赔付
func (game *Game) getAllPay() int64 {
	allPay := int64(0)
	for _, v := range game.AllUserList {
		if v.User.IsRobot() {
			continue
		}
		allPay += v.GetWin()
	}
	return allPay
}

//获取时间列表
func (game *Game) getTimeList() []*msg.TimeList {
	timeList := make([]*msg.TimeList, 8)
	sum := make([]int, 8)
	tem := make(map[int]*msg.TimeList, 8)
	list := config.GetTimeList()
	for k, _ := range list {
		times, s := game.getOneTimeList(list[k].([]interface{}))
		tem[s] = times
		sum[k] = s
	}
	sort.Slice(sum, func(i, j int) bool {
		return sum[i] < sum[j]
	})
	for k, v := range game.Result {
		timeList[v-1] = tem[sum[k]]
		fmt.Println("v = ", v, sum[k])
	}
	return timeList
}

//获取单一时间列表
func (game *Game) getOneTimeList(list []interface{}) (*msg.TimeList, int) {
	times := make([]int32, len(list))
	sum := 0
	for k, _ := range list {
		value, _ := strconv.Atoi(list[k].(json.Number).String())
		times[k] = int32(value)
		sum += value
	}
	return &msg.TimeList{
		Times: times,
	}, sum
}

//计算赔付
func (game *Game) pay(result []int32) {
	payArea := game.GetPayArea(result)
	for _, v := range game.AllUserList {
		v.CountInfo = make(map[msg.BetArea]int64, 38)
		for k, v1 := range v.BetInfo {
			if payArea[k] == 1 {
				if k < msg.BetArea_first_second_12 {
					v.CountInfo[k] = v1 * int64(FirstPay)
				} else if k < msg.BetArea_champion_Man {
					v.CountInfo[k] = v1 * int64(FAndSPay)
				} else if k == msg.BetArea_champion_Man {
					v.CountInfo[k] = int64(float64(v1) * FirstMalePay)
				} else if k == msg.BetArea_champion_Woman {
					v.CountInfo[k] = int64(float64(v1) * FirstFemalePay)
				}
			}
		}
	}
}

//获取所有可能结果
func (game *Game) GetAllResult() [][]int32 {
	allResult := make([][]int32, 0)
	for i := 1; i < 9; i++ {
		for j := 1; j < 9; j++ {
			if j != i {
				result := make([]int32, 2)
				result[0] = int32(i)
				result[1] = int32(j)
				allResult = append(allResult, result)
			}
		}
	}
	return allResult
}

//获取赔付区域
func (game *Game) GetPayArea(result []int32) map[msg.BetArea]int {
	first := result[0]
	second := result[1]
	pay := make(map[msg.BetArea]int)
	switch first {
	case 1:
		switch second {
		case 2:
			pay[msg.BetArea_first_second_12] = 1
			break
		case 3:
			pay[msg.BetArea_first_second_13] = 1
			break
		case 4:
			pay[msg.BetArea_first_second_14] = 1
			break
		case 5:
			pay[msg.BetArea_first_second_15] = 1
			break
		case 6:
			pay[msg.BetArea_first_second_16] = 1
			break
		case 7:
			pay[msg.BetArea_first_second_17] = 1
			break
		case 8:
			pay[msg.BetArea_first_second_18] = 1
			break
		}
		pay[msg.BetArea_champion_1] = 1
		pay[msg.BetArea_champion_Woman] = 1
		break
	case 2:
		switch second {
		case 1:
			pay[msg.BetArea_first_second_12] = 1
			break
		case 3:
			pay[msg.BetArea_first_second_23] = 1
			break
		case 4:
			pay[msg.BetArea_first_second_24] = 1
			break
		case 5:
			pay[msg.BetArea_first_second_25] = 1
			break
		case 6:
			pay[msg.BetArea_first_second_26] = 1
			break
		case 7:
			pay[msg.BetArea_first_second_27] = 1
			break
		case 8:
			pay[msg.BetArea_first_second_28] = 1
			break
		}
		pay[msg.BetArea_champion_2] = 1
		pay[msg.BetArea_champion_Woman] = 1
		break
	case 3:
		switch second {
		case 2:
			pay[msg.BetArea_first_second_23] = 1
			break
		case 1:
			pay[msg.BetArea_first_second_13] = 1
			break
		case 4:
			pay[msg.BetArea_first_second_34] = 1
			break
		case 5:
			pay[msg.BetArea_first_second_35] = 1
			break
		case 6:
			pay[msg.BetArea_first_second_36] = 1
			break
		case 7:
			pay[msg.BetArea_first_second_37] = 1
			break
		case 8:
			pay[msg.BetArea_first_second_38] = 1
			break
		}
		pay[msg.BetArea_champion_3] = 1
		pay[msg.BetArea_champion_Man] = 1
		break
	case 4:
		switch second {
		case 2:
			pay[msg.BetArea_first_second_24] = 1
			break
		case 3:
			pay[msg.BetArea_first_second_34] = 1
			break
		case 1:
			pay[msg.BetArea_first_second_14] = 1
			break
		case 5:
			pay[msg.BetArea_first_second_35] = 1
			break
		case 6:
			pay[msg.BetArea_first_second_36] = 1
			break
		case 7:
			pay[msg.BetArea_first_second_37] = 1
			break
		case 8:
			pay[msg.BetArea_first_second_38] = 1
			break
		}
		pay[msg.BetArea_champion_4] = 1
		pay[msg.BetArea_champion_Man] = 1
		break
	case 5:
		switch second {
		case 2:
			pay[msg.BetArea_first_second_25] = 1
			break
		case 3:
			pay[msg.BetArea_first_second_35] = 1
			break
		case 4:
			pay[msg.BetArea_first_second_45] = 1
			break
		case 1:
			pay[msg.BetArea_first_second_15] = 1
			break
		case 6:
			pay[msg.BetArea_first_second_56] = 1
			break
		case 7:
			pay[msg.BetArea_first_second_57] = 1
			break
		case 8:
			pay[msg.BetArea_first_second_58] = 1
			break
		}
		pay[msg.BetArea_champion_5] = 1
		pay[msg.BetArea_champion_Man] = 1
		break
	case 6:
		switch second {
		case 2:
			pay[msg.BetArea_first_second_26] = 1
			break
		case 3:
			pay[msg.BetArea_first_second_36] = 1
			break
		case 4:
			pay[msg.BetArea_first_second_46] = 1
			break
		case 5:
			pay[msg.BetArea_first_second_56] = 1
			break
		case 1:
			pay[msg.BetArea_first_second_16] = 1
			break
		case 7:
			pay[msg.BetArea_first_second_67] = 1
			break
		case 8:
			pay[msg.BetArea_first_second_68] = 1
			break
		}
		pay[msg.BetArea_champion_6] = 1
		pay[msg.BetArea_champion_Man] = 1
		break
	case 7:
		switch second {
		case 2:
			pay[msg.BetArea_first_second_27] = 1
			break
		case 3:
			pay[msg.BetArea_first_second_37] = 1
			break
		case 4:
			pay[msg.BetArea_first_second_47] = 1
			break
		case 5:
			pay[msg.BetArea_first_second_57] = 1
			break
		case 6:
			pay[msg.BetArea_first_second_67] = 1
			break
		case 1:
			pay[msg.BetArea_first_second_17] = 1
			break
		case 8:
			pay[msg.BetArea_first_second_78] = 1
			break
		}
		pay[msg.BetArea_champion_7] = 1
		pay[msg.BetArea_champion_Man] = 1
		break
	case 8:
		switch second {
		case 2:
			pay[msg.BetArea_first_second_28] = 1
			break
		case 3:
			pay[msg.BetArea_first_second_38] = 1
			break
		case 4:
			pay[msg.BetArea_first_second_48] = 1
			break
		case 5:
			pay[msg.BetArea_first_second_58] = 1
			break
		case 6:
			pay[msg.BetArea_first_second_68] = 1
			break
		case 7:
			pay[msg.BetArea_first_second_78] = 1
			break
		case 1:
			pay[msg.BetArea_first_second_18] = 1
			break
		}
		pay[msg.BetArea_champion_8] = 1
		pay[msg.BetArea_champion_Man] = 1
		break
	}
	return pay
}

//结算
func (game *Game) saveScore() {
	game.Table.AddTimer(int64(countTime), game.start)
	fmt.Println("save score")
	if len(game.GameInfo) >= 10 {
		game.GameInfo = append(game.GameInfo[:0], game.GameInfo[1:]...)
	}
	game.GameInfo = append(game.GameInfo, game.TempGameInfo)
	for _, v := range game.AllUserList {
		if v.AllBet == 0 {
			v.NoBetCount++
		}
		game.checkMarquee(v.Win, v.User.GetNike())
		score, _ := v.User.SetScore(game.Table.GetGameNum(), v.Win, game.Table.GetRoomRate())
		v.UserInfo.Amount += score
		if !v.User.IsRobot() {
			game.createOperationLog(v, game.TempGameInfo)
			v.User.SendChip(v.GetChip())
			v.User.SendRecord(game.Table.GetGameNum(), score-v.AllBet, v.AllBet, v.Win-score, score, "")
		}
	}
	game.resetData()
	game.checkKickOutUser()
	game.Table.EndGame()
	game.Status = msg.GameStatus_game_End
}

//跑马灯
func (game *Game) checkMarquee(win int64, nickName string) {
	orderRules := game.Table.GetMarqueeConfig()
	length := len(orderRules)
	for i := 0; i < length; i++ {
		v := orderRules[i]
		//SpecialCondition
		//special, _ := strconv.ParseInt(v.GetSpecialCondition(), 10, 64)
		if v.GetAmountLimit() < 0 || win < v.GetAmountLimit() {
			continue
		}
		game.Table.CreateMarquee(nickName, win, "", v.GetRuleId())
		break
	}
}

//操作日志
func (game *Game) createOperationLog(user *model.User, gameInfo *msg.GameInfo) {
	content := "当前作弊率: " + strconv.Itoa(int(game.ControlKey)) +
		" 开奖结果: " + " 第一名编号: " + strconv.Itoa(int(gameInfo.GetFirstIndex())) + "号马 " +
		" 第二名编号: " + strconv.Itoa(int(gameInfo.GetSecondIndex())) + "号马 " + user.GetOperationLog()
	log.Tracef("content = ", content)
	game.Table.WriteLogs(user.User.GetID(), content)
}

//踢出长时间不下注的玩家
func (game *Game) checkKickOutUser() {
	for _, v := range game.AllUserList {
		if v.NoBetCount >= 5 {
			v.User.SendMsg(int32(msg.MsgId_Bet_Fail_Res), &msg.BetFailedRes{
				UserId: v.User.GetID(),
				Reason: "长时间未下注",
			})
			game.userLeave(v.User)
			game.Table.KickOut(v.User)
		}
	}
}

//重置数据
func (game *Game) resetData() {
	for _, v := range game.AllUserList {
		v.ResetData()
	}
	game.AllBet = 0
	game.Result = make([]int32, 0)
	game.BetInfo = make(map[msg.BetArea]int64, 38)
}

func (game *Game) sendStatusChangeMsg() {
	game.T = time.Now().UnixNano() / 1e6
	res := &msg.GameStatusChangeRes{}
	res.GameStatus = game.Status
	res.WaitTime = game.getWaitTime()
	game.Table.Broadcast(int32(msg.MsgId_Status_Change_Res), res)
}

//获取阶段时间
func (game *Game) getWaitTime() int32 {
	switch game.Status {
	case msg.GameStatus_game_Start:
		return int32(gameStartTime)
	case msg.GameStatus_game_Bet:
		return int32(betTime)
	case msg.GameStatus_game_Bet_End:
		return int32(betEndTime)
	case msg.GameStatus_game_Count:
		return int32(gameTime + countTime)
	default:
		return 0
	}
}

//按照概率排名
func (game *Game) getOrderIndex(chances []int32) []int32 {
	log.Tracef("chances =", chances)
	indexs := make([]int32, 0)
	length := len(chances)
	c := make([]int32, 0)
	c = append(c, chances...)
	for i := 0; i < length; i++ {
		chances1 := game.increment(c)
		index := game.getAChance(chances1)
		if index > -1 {
			c[index] = 0
			indexs = append(indexs, index)
		}
	}
	log.Tracef("indexs = ", indexs)
	return indexs
}

//补足10000概率
func (game *Game) increment(chances []int32) []int32 {
	tem := make([]int32, 0)
	tem = append(tem, chances...)
	chance := game.sum(tem)
	length := len(tem)
	increment := int32(0)
	if chance < 10000 {
		increment = (10000 - chance) / int32(length)
	}
	if increment > 0 {
		for j := 0; j < length; j++ {
			if tem[j] != 0 {
				tem[j] += increment
			}
		}
	}
	return tem
}

//确定一个名次
func (game *Game) getAChance(chances []int32) int32 {
	chance := game.sum(chances)
	if chance <= 0 {
		log.Tracef("chances =", chances)
		return -1
	}
	r := int32(rand.Intn(int(chance)))
	index := int32(-1)
	length := len(chances)
	count := int32(0)
	for i := 0; i < length; i++ {
		count += chances[i]
		if r < count {
			index = int32(i)
			break
		}
	}
	return index
}

//求和
func (game *Game) sum(chances []int32) int32 {
	length := len(chances)
	chance := int32(0)
	for j := 0; j < length; j++ {
		chance += chances[j]
	}
	return chance
}
