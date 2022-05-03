package server

import (
	"encoding/json"
	"go-game-sdk/define"
	"go-game-sdk/example/game_buyu/980101/config"
	"go-game-sdk/example/game_buyu/980101/data"
	"go-game-sdk/example/game_buyu/980101/msg"
	"go-game-sdk/inter"
	"go-game-sdk/lib/clock"
	"go-game-sdk/sdk/global"
	"math"
	rand2 "math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	frameMsg "github.com/kubegames/kubegames-sdk/app/message"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/golang/protobuf/proto"
)

var (
	hight                = 1334
	weight               = 750
	hightMax             = 1550
	weightMax            = 1000
	fishWeight           = 300
	fishTideTime         = 60000
	fishTideForecastTime = 540000
	shotOffTime          = 300000
	yuchaochixushijian   = 0
	lastTime             = int64(0)
	diff                 = int64(0)
)

type TableLogic struct {
	Users           map[int64]*data.User
	Pool            int32 //资金池
	Table           table.TableInterface
	Fishes          map[int32]*Fish
	FishNum         map[msg.Type]int
	FishCap         map[msg.Type]int
	Robots          map[int64]*Robot
	IsDismiss       bool
	SceneId         int
	MaxSceneNum     int
	IslandId        string
	Frozen          bool
	FishTide        bool
	TimeFish        map[string]*TimeFish
	M               sync.Mutex
	LastDismissTime int64
	timer           map[string]*clock.Job
	start           bool
	//AddedChance  int32
	uniqueId int32
	skillNum map[int32]int
}

type Fish struct {
	info     *msg.Fish
	score    int32
	protect  int32
	deadTime int64
	lineName string
}

type TimeFish struct {
	id        string
	speed     int32
	startTime int
	timeSpace int
	num       int
	totalNum  int
	table     *TableLogic
}

func (timeFish *TimeFish) start() {

	if timeFish.table.IsDismiss {
		return
	}
	//if timeFish.totalNum > 0 && timeFish.num >= timeFish.totalNum {
	//	return
	//}
	if !timeFish.table.FishTide && !timeFish.table.Frozen {
		timeFish.table.refreshFish(timeFish.id, false)
	}
	//timeFish.num++
}

func (timeFish *TimeFish) startTimer() {
	job, _ := timeFish.table.Table.AddTimerRepeat(int64(timeFish.timeSpace), uint64(timeFish.totalNum), timeFish.start)
	timeFish.table.timer["timefish"+timeFish.id+strconv.Itoa(timeFish.timeSpace)] = job
}

func (table *TableLogic) init(table2 table.TableInterface) {
	table.Table = table2
	table.reset()
	//go func() {
	//	http.ListenAndServe(":8888", nil)
	//}()
	//go func() {
	//	runtime.GC()
	//	time.Sleep(1000)
	//}()
}

func (table *TableLogic) reset() {
	table.Users = make(map[int64]*data.User, 0)
	table.Robots = make(map[int64]*Robot, 0)
	table.timer = make(map[string]*clock.Job, 0)
	table.skillNum = make(map[int32]int, 0)
	table.Pool = 10000000
	table.Frozen = false
	table.FishTide = false
	table.IsDismiss = true
	table.uniqueId = 0
	fishWeight = config.GetFishWeight()
	fishTideTime = config.GetFishTideTime()
	fishTideForecastTime = config.GetFishTideForecastTime()
	shotOffTime = config.GetShotoffTime()
	table.resetFish()
}

func (table *TableLogic) resetFish() {
	table.FishNum = make(map[msg.Type]int, 0)
	table.FishNum[msg.Type_SMALL] = 0
	table.FishNum[msg.Type_MIDDLE] = 0
	table.FishNum[msg.Type_BIG] = 0
	table.FishNum[msg.Type_BOSS] = 0
	table.FishNum[msg.Type_PILE] = 0
	table.FishNum[msg.Type_KING] = 0
	table.Fishes = make(map[int32]*Fish, 0)
	table.TimeFish = table.createTimeFish()
}

func (table *TableLogic) updateTableInfo() {
	table.MaxSceneNum = config.GetSceneNum(table.IslandId)
	table.changeSceneId()
	table.FishCap = config.GetSceneFishCap(table.IslandId, strconv.Itoa(table.SceneId))
}

func (table *TableLogic) createTimeFish() map[string]*TimeFish {
	timeFish := make(map[string]*TimeFish, 0)
	seed := int64(0)
	for _, v := range config.GetTimeFish() {
		fishId, start, space, num := config.GetTimeFishInfo(v, table.randInt, seed)
		timeFish[fishId+string(space)] = &TimeFish{
			id:        fishId,
			timeSpace: space,
			startTime: start,
			totalNum:  num,
			table:     table,
		}
		seed += 1000
	}
	return timeFish
}

//用户坐下
func (table *TableLogic) OnActionUserSitDown(user player.PlayerInterface, chairId int, cfg string) int {
	//if len(table.Users) > 0 {
	//	return false
	//}
	if config.GetFishBet(cfg, 1) <= 0 {
		return define.SIT_DOWN_ERROR_OVER
	}
	if len(table.Users) == 0 && user.IsRobot() {
		return define.SIT_DOWN_ERROR_NORMAL
	}
	if len(table.Users) > 3 {
		return define.SIT_DOWN_ERROR_NORMAL
	}
	now := time.Now().UnixNano() / 1e6
	if now-table.LastDismissTime < 10000 {
		return define.SIT_DOWN_ERROR_NORMAL
	}
	return define.SIT_DOWN_OK
}

func (table *TableLogic) ResetTable() {
	table.saveScore()
	table.dismiss()
}

func (table *TableLogic) CloseTable() {

}

func (table *TableLogic) UserExit(user player.PlayerInterface) bool {
	table.userLeave(user)
	return true
}

func (table *TableLogic) LeaveGame(user player.PlayerInterface) bool {
	table.userLeave(user)
	return true
}

func (table *TableLogic) SendScene(user player.PlayerInterface) bool {
	user.SendMsg(int32(msg.MsgId_ZERO), &msg.Fish{})
	return true
}

func (table *TableLogic) GameStart(user player.PlayerInterface) bool {
	if user.IsRobot() {
		user.SendMsg(int32(msg.MsgId_ZERO), &msg.Fish{})
	}
	return true
}

func (table *TableLogic) userLeave(user player.PlayerInterface) {
	table.Table.Broadcast(int32(msg.MsgId_EXIST_ROOM_Res), &msg.ExistRoomRes{UserId: user.GetID()})
	u := table.Users[user.GetID()]
	//gameNum := user.GetRoomNum()
	if u != nil {
		table.saveUserScore(u)
		u.WriteLog()
		gameNum := u.GameNum
		user.SendRecord(gameNum, u.OutputAmount+u.Bet, -u.Bet, u.Win-u.OutputAmount, u.OutputAmount, "")
	}
	delete(table.Users, user.GetID())

	if !user.IsRobot() && !table.IsDismiss {
		table.changeCalculateRobot()
	}
	table.checkDismissTable()
	table.checkRobot()

}

//游戏消息
func (table *TableLogic) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(msg.MsgId_SHOOT_Req):
		table.shoot(buffer, user)
		break
	case int32(msg.MsgId_HIT_Req):
		table.hit(buffer, user)
		break
	case int32(msg.MsgId_UPGRADE_Req):
		table.upgrade(buffer, user)
		break
	case int32(msg.MsgId_DEAD_Req):
		table.dead(buffer, user)
		break
	case int32(msg.MsgId_EXIST_ROOM_Req):
		table.UserExit(user)
		table.Table.KickOut(user)
		break
	case int32(msg.MsgId_CHANGEMODEL_Req):
		table.changeModel(buffer, user)
		break
	case int32(msg.MsgId_INTO_ROOM_Req):
		table.enterRoom(buffer, user)
		break
	case int32(msg.MsgId_CHANGELOCKFISH_Req):
		table.changeLockFish(buffer, user)
		break
	case int32(msg.MsgId_SKILLHIT_Req):
		table.skillHit(buffer, user)
		break
	case int32(msg.MsgId_TEST_Req):
		table.test(buffer, user)
		break

	}
}

func (table *TableLogic) UserReady(user player.PlayerInterface) bool {
	return true
}

func (table *TableLogic) changeSceneId() {
	sceneId := table.SceneId
	if sceneId != 0 {
		sceneId++
	}
	if sceneId > table.MaxSceneNum {
		sceneId = 1
	}
	if sceneId == 0 {
		sceneId = rand.RandInt(0, table.MaxSceneNum) + 1
	}
	table.SceneId = sceneId
}

func (table *TableLogic) sendChangeScene() {
	req := &msg.ChangemSceneReq{
		SceneId: int32(table.SceneId),
	}
	table.Table.Broadcast(int32(msg.MsgId_CHANGESCENE_Req), req)
}

func (table *TableLogic) sendFishTideEnd() {
	req := &msg.FishtideEndReq{}
	table.Table.Broadcast(int32(msg.MsgId_FISHTIDEEND_Req), req)
}

func (table *TableLogic) checkDismissTable() {
	for _, u := range table.Users {
		if !u.IsRobot {
			return
		}
	}
	if !table.IsDismiss {
		//table.Table.AddTimer(int64(2000), table.dismiss)
		//table.IsDismiss = true
		//table.LastDismissTime = time.Now().UnixNano() / 1e6
		//table.dismiss()
		//table.ResetTable()
		//table.Table.EndGame()
		table.IsDismiss = true
	}
}

func (table *TableLogic) kickOutAllUser() {
	if !table.IsDismiss {
		return
	}
	for _, u := range table.Users {
		if !u.IsRobot {
			return
		}
	}
	table.LastDismissTime = time.Now().UnixNano() / 1e6
	for _, r := range table.Users {
		table.Table.KickOut(r.InnerUser)
	}
	table.dismiss()
	table.Table.EndGame()
}

func (table *TableLogic) dismiss() {
	//for _, r := range table.Robots {
	table.stopTimer()
	table.reset()
}

func (table *TableLogic) stopTimer() {
	for _, v := range table.timer {
		if v != nil {
			v.Cancel()
		}
	}
}

func (table *TableLogic) enterRoom(buffer []byte, user player.PlayerInterface) {
	req := &msg.EnterRoomReq{}
	proto.Unmarshal(buffer, req)
	user2 := data.NewUser(table.Table)
	user2.UserInfo = &msg.UserInfo{
		UserId:     user.GetID(),
		UserName:   user.GetNike(),
		Head:       user.GetHead(),
		Amount:     user.GetScore(),
		ChairId:    int32(user.GetChairID()),
		BulletLv:   1,
		LockFishId: -1,
		IsRobot:    user.IsRobot(),
	}
	user2.IsRobot = user.IsRobot()
	user2.GameNum = user.GetRoomNum()
	user2.Log = make([]*frameMsg.GameLog, 0)
	if user.IsRobot() {
		user2.UserInfo.Lock = true
	}
	if table.IsDismiss {
		table.changeIslandId()
		table.updateTableInfo()
	}
	user2.InnerUser = user
	user2.LastShootTime = time.Now().UnixNano() / 1e6
	table.Users[user.GetID()] = user2
	res := &msg.EnterRoomRes{}
	fishes := make([]*Fish, 0)
	fisheInfos := make([]*msg.Fish, 0)
	for _, v := range table.Fishes {
		fish := v
		fishes = append(fishes, fish)
		fisheInfos = append(fisheInfos, fish.info)
	}
	infos := make([]*msg.UserInfo, 0)
	for _, v := range table.Users {
		info := v
		infos = append(infos, info.UserInfo)
	}
	tableInfo := &msg.TableInfo{
		TableId:     int32(table.Table.GetID()),
		Fishes:      fisheInfos,
		UserInfoArr: infos,
		SceneId:     int32(table.SceneId),
		FishTide:    table.FishTide,
		ServerTime:  time.Now().UnixNano() / 1e6,
	}
	res.TableInfo = tableInfo
	//table.Table.Broadcast(int32(msg.MsgId_INTO_ROOM_Res), res)
	user.SendMsg(int32(msg.MsgId_INTO_ROOM_Res), res)
	table.Table.Broadcast(int32(msg.MsgId_SOMEONEENTER_Req), &msg.SomeoneEnter_Req{
		UserInfoArr: infos,
	})
	if table.IsDismiss {
		table.IsDismiss = false
		table.Table.StartGame()
		table.tick()
		table.changeCalculateRobot()
	}
	table.checkRobot()
}

func (table *TableLogic) changeIslandId() {
	table.IslandId = strconv.Itoa(int(table.Table.GetLevel()))
}

func (table *TableLogic) changeCalculateRobot() {
	for _, u := range table.Users {
		if !u.IsRobot {
			u.UserInfo.CalculateRobot = true
			req := &msg.ChangeCalculateRobot_Req{
				UserId: u.UserInfo.GetUserId(),
			}
			u.Table.Broadcast(int32(msg.MsgId_CHANGECALCULATEROBOT_Req), req)
			break
		}
	}
}

func (table *TableLogic) shoot(buffer []byte, user player.PlayerInterface) {
	req := &msg.ShootReq{}
	proto.Unmarshal(buffer, req)
	user2 := table.Users[req.GetUserId()]
	if user2 == nil {
		return
	}
	bet := config.GetFishBet(table.Table.GetAdviceConfig(), user2.UserInfo.GetBulletLv())
	if user2.InnerUser.GetScore()+user2.TaxedScore+user2.SubScore < int64(bet) {
		if user2.IsRobot { //
			table.robotQuit(user2.UserInfo.GetUserId())
		}
		return
	}
	if user2.IsRobot {
		//table.updateScore(user2, -int64(bet))
		table.count(user2, -int64(bet))
	}
	table.Users[user.GetID()].LastShootTime = time.Now().UnixNano() / 1e6
	res := &msg.ShootRes{
		UserId: req.GetUserId(),
		Point:  req.GetPoint(),
	}
	table.Table.Broadcast(int32(msg.MsgId_SHOOT_Res), res)
}

func (table *TableLogic) hit(buffer []byte, user player.PlayerInterface) {
	req := &msg.HitReq{}
	proto.Unmarshal(buffer, req)
	user2 := table.Users[req.GetUserId()]
	bulletLv := req.GetBulletLv()
	bet := config.GetFishBet(table.Table.GetAdviceConfig(), bulletLv)
	if user2 == nil || user2.InnerUser.GetScore()+user2.TaxedScore+user2.SubScore < int64(bet) {
		return
	}
	user2.BulletNum++
	if user2.BulletNum > 10 {
		return
	}
	if bulletLv < 0 || bulletLv > 10 {
		table.shotOffGame(user2)
		table.createOperationLog(user2, 0, 1, 0, 0, "", true)
	}
	coinChange, dead, key := table.hitFish(user2, req.GetFishId(), bulletLv)
	//table.updateScore(user2, int64(coinChange))
	table.count(user2, int64(coinChange))
	table.count(user2, int64(dead.GetScore()))
	dead.Score = float64(table.tax(int64(dead.GetScore())))
	res := &msg.HitRes{
		UserId: req.GetUserId(),
		Fish:   dead,
		Key:    key,
	}
	res.CoinChange = bet
	table.Table.Broadcast(int32(msg.MsgId_HIT_Res), res)
	if key != "" && dead.GetScore() > 0 {
		fishName := config.GetFishName(key)
		fishScore := config.GetFishScore(key)
		table.checkMarquee(user2.InnerUser.GetNike(), fishName, int64(fishScore*bet), int64(fishScore))
	}
	if key != "" && !user2.InnerUser.IsRobot() {
		table.createOperationLog(user2, int64(dead.GetScore()), 1, user2.InnerUser.GetScore()+user2.SubScore+user2.TaxedScore, bet, key, false)
	}
	//if dead.GetScore() > 0 {
	//table.checkLockFishId(req.GetFishId(), user2.InnerUser)
	//table.checkRobotBehaviour(req.GetFishId())
	//table.checkRobotBulletLv(table.getRobot(req.GetUserId()))
	//}
}

func (table *TableLogic) tax(coinChange int64) int64 {
	if coinChange > 0 {
		coin := coinChange * (10000 - table.Table.GetRoomRate()) / 10000
		return coin
	}
	return coinChange
}

func (table *TableLogic) count(user *data.User, coinChange int64) {
	if coinChange > 0 {
		user.AddScore += coinChange
		user.Win += coinChange
		user.OutputAmount += table.tax(coinChange)
		user.TaxedScore += table.tax(coinChange)
		user.UserInfo.Amount += table.tax(coinChange)
		return
	}
	user.Bet += coinChange
	user.SubScore += coinChange
	user.UserInfo.Amount += coinChange
}

func (table *TableLogic) hitFish(user *data.User, fishId, bulletLv int32) (int32, *msg.DeadFish, string) {
	bet := config.GetFishBet(table.Table.GetAdviceConfig(), bulletLv)
	coinChange := int32(0)
	fish := table.Fishes[fishId]
	dead := &msg.DeadFish{
		FishId: fishId,
	}
	key := ""
	if fish != nil {
		coinChange = -bet
		if user.IsRobot {
			coinChange = 0
		}
		key = fish.info.GetFishId()
		skillId := fish.info.GetSkillId()
		winCoin := table.getWinCoin(user, fish.info.GetFishId(), bet, fish.score, 0)
		if winCoin > 0 {
			table.checkFishNum(fish)
			delete(table.Fishes, fishId)
			//coinChange += winCoin
			dead.Score = float64(winCoin)
			table.skillNum[skillId]++
			if skillId == 3 {
				table.Frozen = true
			}
		}
	}
	table.Pool += -coinChange
	return coinChange, dead, key
}

func (table *TableLogic) getWinCoin(user *data.User, fishId string, bet int32, score int32, seed int64) int32 {
	chance := config.GetFishHitChance(fishId, user.IsRobot)
	//if user.IsRobot {
	//	chance += table.AddedChance
	//}
	key := table.Table.GetRoomProb()
	PointKey := user.InnerUser.GetProb()
	if PointKey == 0 {
		chance += config.GetXueChiChance(fishId, strconv.Itoa(int(table.Table.GetLevel())), key)
	}
	if PointKey != 0 {
		chance += config.GetXueChiChance(fishId, strconv.Itoa(int(table.Table.GetLevel())), PointKey)
	}
	if table.getChanceWan(chance, seed) {
		return bet * score
	}
	return 0
}

func (table *TableLogic) getControlKey(user *data.User) (int32, string) {
	key := table.Table.GetRoomProb()
	PointKey := user.InnerUser.GetProb()
	if key == 0 {
		key = 1000
	}
	if PointKey == 0 {
		return key, "否"
	}
	return PointKey, "是"
}

func (table *TableLogic) skillHit(buffer []byte, user player.PlayerInterface) {
	req := &msg.SkillHitReq{}
	proto.Unmarshal(buffer, req)
	user2 := table.Users[req.GetUserId()]
	if user2 == nil {
		return
	}
	coinChange := int32(0)
	kills := make([]int32, 0)
	bet := config.GetFishBet(table.Table.GetAdviceConfig(), user2.UserInfo.GetBulletLv())
	fishes := make([]*msg.DeadFish, 0)
	skillId := req.GetSkillId()
	fishId := req.GetFishId()
	if table.skillNum[skillId] < 1 {
		return
	}
	table.skillNum[skillId]--
	if user2 != nil && skillId != 3 {
		num := config.GetSkillHitNum(fishId, skillId)
		killNum := config.GetSkillFishNum(fishId, skillId)
		for _, fishId := range req.GetFishIds() {
			hitFish := table.Fishes[fishId]
			if hitFish == nil {
				continue
			}
			for i := 0; i < num; i++ {
				winCoin := table.getWinCoin(user2, hitFish.info.GetFishId(), bet, hitFish.score, int64(i))
				if winCoin > 0 {
					kills = append(kills, fishId)
					coinChange += winCoin
					winCoin = int32(table.tax(int64(winCoin)))
					fishes = append(fishes, &msg.DeadFish{
						FishId: fishId,
						Score:  float64(winCoin),
					})
					table.checkFishNum(hitFish)
					delete(table.Fishes, fishId)
					//table.checkLockFishId(fishId, user)
					//table.checkRobotBehaviour(fishId)
					break
				}
			}
			if skillId != 4 && len(fishes) > killNum {
				break
			}
		}
	}
	if skillId == 3 {
		table.skillFrozen(fishId)
	}
	//if cionChange > 0 {
	res := &msg.SkillHitRes{
		FishIds: fishes,
		Fishes:  req.GetFishIds(),
		SkillId: skillId,
		UserId:  req.GetUserId(),
		Point:   req.GetPoint(),
	}
	//table.updateScore(user2, int64(coinChange))
	table.count(user2, int64(coinChange))
	table.Table.Broadcast(int32(msg.MsgId_SKILLHIT_Res), res)
	table.checkMarquee(user2.InnerUser.GetNike(), config.GetFishName(fishId), int64(coinChange), int64(coinChange))
	if !user2.InnerUser.IsRobot() {
		table.createOperationLog(user2, int64(coinChange), 2, user2.InnerUser.GetScore()+user2.SubScore+user2.TaxedScore, 0, "", false)
	}
	//}

}

func (table *TableLogic) updateScore(user *data.User, coinChange int64) {
	log.Tracef("user update score :", coinChange)
	//bussType := int32(100701)
	//betAmount := coinChange
	//if coinChange > 0 {
	//	bussType = 200701
	//	betAmount = 0
	//}
	user.InnerUser.SetScore(table.Table.GetGameNum(), coinChange, table.Table.GetRoomRate())
	user.UserInfo.Amount = user.InnerUser.GetScore()
}

func (table *TableLogic) checkMarquee(nickName, fishName string, coin, fishScore int64) {
	orderRules := table.orderMarqueeRules(table.Table.GetMarqueeConfig())
	//for _, v := range orderRules {
	length := len(orderRules)
	for i := 0; i < length; i++ {
		v := orderRules[i]
		//SpecialCondition
		special, _ := strconv.ParseInt(v.GetSpecialCondition(), 10, 64)
		if v.GetAmountLimit() < 0 || coin < v.GetAmountLimit() || (special > 0 && fishScore <= special) {
			continue
		}
		table.createMarquee(nickName, fishName, coin, v.GetRuleId())
		break
	}
}

func (table *TableLogic) orderMarqueeRules(rules []*frameMsg.MarqueeConfig) []*frameMsg.MarqueeConfig {
	orderRules := make([]*frameMsg.MarqueeConfig, 0)
	orderRules = append(orderRules, rules...)
	length := len(orderRules)
	for i := 0; i < length; i++ {
		for j := i + 1; j < length; j++ {
			change := false
			special, _ := strconv.ParseInt(orderRules[i].GetSpecialCondition(), 10, 64)
			special1, _ := strconv.ParseInt(orderRules[j].GetSpecialCondition(), 10, 64)
			if special != 0 {
				if special1 != 0 && special1 > special {
					change = true
				}
			}
			if special == 0 && (orderRules[i].GetAmountLimit() < orderRules[j].GetAmountLimit() ||
				special1 != 0) {
				change = true
			}
			if change {
				tem := orderRules[i]
				orderRules[i] = orderRules[j]
				orderRules[j] = tem
			}
		}
	}
	return orderRules
}

func (table *TableLogic) createMarquee(nickName, special string, coin, ruleId int64) {
	table.Table.CreateMarquee(nickName, coin, special, ruleId)
}

func (table *TableLogic) createOperationLog(user *data.User, coinChange, operationType, userScore int64, bet int32, fishId string, isAbnormal bool) {
	operation := table.getOperationExplain(operationType)
	userId := user.InnerUser.GetID()
	//content := operation + " 赔付金额:" + strconv.FormatInt(coinChange, 10) +
	//	" 炮弹倍数: " + strconv.Itoa(int(bet)) + " 击中鱼ID : " + fishId +
	//	" 击中鱼时间: " + time.Now().Format("2006-01-02 15:04:05") +
	//	" 用户剩余金额:" + strconv.FormatInt(userScore, 10)
	key, isPoint := table.getControlKey(user)
	content := "用户ID:" + strconv.FormatInt(userId, 10) + operation + " 赔付金额:" + score.GetScoreStr(coinChange) +
		" 炮弹倍数: " + score.GetScoreStr(int64(bet)) + " 击中鱼ID : " + fishId +
		" 击中鱼时间: " + time.Now().Format("2006-01-02 15:04:05") +
		" 用户剩余金额:" + score.GetScoreStr(userScore) +
		" 作弊值: " + strconv.Itoa(int(key)) +
		" 是否点控: " + isPoint
	if isAbnormal {
		content += " 用户数据异常: 是"
	}
	user.Log = append(user.Log, &frameMsg.GameLog{
		UserId:  userId,
		Content: content,
	})
	//table.Table.WriteLogs(userId, content)
}

func (table *TableLogic) getOperationExplain(operationType int64) string {
	switch operationType {
	case 1:
		return "发炮结算:"
	case 2:
		return "技能结算:"
	default:
		return ""

	}
}

func (table *TableLogic) skillFrozen(fishId string) {
	table.Frozen = true
	t := config.GetSkillHitNum(fishId, 3) * 1000
	if t == 0 {
		t = 10000
	}
	table.addFishTime(int64(t))
	table.Table.AddTimer(int64(t), func() {
		table.Frozen = false
	})
}

func (table *TableLogic) upgrade(buffer []byte, user player.PlayerInterface) {
	req := &msg.UpgradeReq{}
	proto.Unmarshal(buffer, req)
	user2 := table.Users[user.GetID()]
	if user2 == nil {
		return
	}
	lv := user2.UserInfo.BulletLv
	if req.GetIsAdd() {
		lv++
	} else {
		lv--
	}
	if lv > config.MaxBulletLv {
		lv = 1
	}
	if lv < 1 {
		lv = config.MaxBulletLv
	}
	table.Users[user.GetID()].UserInfo.BulletLv = lv
	res := &msg.UpgradeRes{
		UserId: user.GetID(),
		Lv:     lv,
	}
	table.Table.Broadcast(int32(msg.MsgId_UPGRADE_Res), res)
}

func (table *TableLogic) refreshFish(fishId string, isCount bool) {
	if fishId == "" || table.Frozen {
		return
	}
	fishType := config.GetFishType(fishId)
	score := int(config.GetFishScore(fishId))
	num := config.GetFishNum(fishId)
	if num > 0 {
		score *= int(num)
	}
	if isCount && table.FishNum[fishType]+score > table.FishCap[fishType] {
		return
	}
	//id := table.getUniqueId()
	line, speeds, t, name, variant, isRandom := table.getLine(fishId)
	if len(speeds) == 0 {
		speed := config.GetFishSpeed(fishId)
		speeds = append(speeds, speed)
	}
	skillId := config.GetSkillId(fishId)
	res := table.createFish(fishId, line, speeds, num, skillId, nil, t, name, variant)
	res.IsRandom = isRandom
	if isCount {
		table.FishNum[fishType] += score
	}
	table.Table.Broadcast(int32(msg.MsgId_REFRESHFISH_Req), res)
	if skillId > 0 {
		table.getSkillNum()
	}
	//if config.GetRobotLockFishes(fishId) {
	//	table.changeRobotBehaviour(table.uniqueId)
	//}
}

func (table *TableLogic) createFormation() {
	key := config.GetAFormationKey()
	info := config.GetFormation(key)
	yuchaochixushijian = config.GetFishTideSustainTime(key)
	for k, _ := range info {
		//if key != "3" {
		table.refreshSomeFishes(key, k)
		//}
		//if key == "3" {
		//	table.refreshSomeCircleFishes(key, k)
		//}
	}
}

func (table *TableLogic) refreshSomeFishes(formationKey, key string) {
	fishId, speed, t, num, lines := config.GetFormationFishInfo(formationKey, key)
	line, _, _ := table.getConfLine(lines, make([]interface{}, 0), json.Number(0))
	speeds := make([]int32, 0)
	speeds = append(speeds, speed)
	skillId := config.GetSkillId(fishId)
	if fishId != "" {
		for i := 0; i < num; i++ {
			if i == 0 {
				res := table.createFish(fishId, line, speeds, 1, skillId, nil, 0, "", []int32{})
				table.Table.Broadcast(int32(msg.MsgId_REFRESHFISH_Req), res)
				continue
			}
			job, _ := table.Table.AddTimer(int64(t*i), func() {
				if table.IsDismiss {
					return
				}
				res := table.createFish(fishId, line, speeds, 1, skillId, nil, 0, "", []int32{})
				table.Table.Broadcast(int32(msg.MsgId_REFRESHFISH_Req), res)
			})
			table.timer[fishId+string(i)] = job
		}
	}
}

func (table *TableLogic) refreshSomeCircleFishes(formationKey, key string) { //survival
	fishId, speed, t, num, lines, radius, overlying, angle, _ := config.GetCircleFormationFishInfo(formationKey, key)
	line, _, _ := table.getConfLine(lines, make([]interface{}, 0), json.Number(0))
	speeds := make([]int32, 0)
	speeds = append(speeds, speed)
	skillId := config.GetSkillId(fishId)
	formationInfo := &msg.FormationInfo{
		Radius:    float32(radius),
		Overlying: float32(overlying),
		Angle:     float32(angle),
		Point:     line[0],
	}
	if fishId != "" {
		for i := 0; i <= num; i++ {
			if i == 0 {
				res := table.createFish(fishId, line, speeds, 1, skillId, formationInfo, 20000, "", []int32{})
				table.Table.Broadcast(int32(msg.MsgId_REFRESHFISH_Req), res)
				continue
			}
			job, _ := table.Table.AddTimer(int64(t*i), func() {
				if table.IsDismiss {
					return
				}
				res := table.createFish(fishId, line, speeds, 1, skillId, formationInfo, 20000, "", []int32{})
				table.Table.Broadcast(int32(msg.MsgId_REFRESHFISH_Req), res)
			})
			table.timer[fishId+string(i)] = job
		}
	}
}

func (table *TableLogic) createFish(fishId string, line []*msg.Point, speed []int32, num int32, skillId int32, formationInfo *msg.FormationInfo, t int32, lineName string, variant []int32) *msg.RefreshFishReq {
	fishes := make([]*msg.Fish, 0)
	for i := int32(0); i < num; i++ {
		id := table.getUniqueId()
		if t == 0 {
			t = table.getBezierLineTime(speed[0], line)
		}
		if len(variant) == 0 {
			variant = []int32{0}
		}
		fish := &msg.Fish{
			Id:            id,
			TypeId:        config.GetFishType(fishId),
			BornTime:      time.Now().UnixNano() / 1e6,
			Line:          line,
			BornPoint:     line[0],
			FishId:        fishId,
			Speed:         speed,
			Num:           num,
			SkillId:       skillId,
			FormationInfo: formationInfo,
			Variant:       variant,
		}
		table.Fishes[id] = &Fish{
			info:     fish,
			score:    config.GetFishScore(fishId),
			deadTime: time.Now().UnixNano()/1e6 + int64(t),
			lineName: lineName,
		}
		fishes = append(fishes, fish)
	}
	offset := table.getPoint(config.GetFishOffset(fishId))
	if num > 1 && len(offset) == 0 {
		offset = []*msg.Point{
			&msg.Point{X: 5, Y: 5},
			&msg.Point{X: 105, Y: 105},
			&msg.Point{X: -105, Y: 205},
			&msg.Point{X: 5, Y: 325},
		}
	}
	res := &msg.RefreshFishReq{
		Fish:   fishes,
		Offset: offset,
	}
	return res
}

func (table *TableLogic) checkFish() {
	if table.IsDismiss {
		return
	}
	dead := make([]int32, 0)
	t := time.Now().UnixNano() / 1e6
	for _, v := range table.Fishes {
		if t > v.deadTime {
			dead = append(dead, v.info.GetId())
		}
	}
	if len(dead) > 0 {
		for _, v := range dead {
			table.checkFishNum(table.Fishes[v])
			//table.checkRobotBehaviour(v)
			delete(table.Fishes, v)
			res := &msg.DeadRes{
				Id: v,
			}
			table.Table.Broadcast(int32(msg.MsgId_DEAD_Res), res)
		}
	}
	//table.Table.AddTimer(int64(1000), table.checkFish)
}

func (table *TableLogic) checkFishNum(fish *Fish) {
	if fish == nil {
		return
	}
	t := fish.info.GetTypeId()
	table.FishNum[t] -= int(fish.score)
	if table.FishNum[t] < 0 {
		table.FishNum[t] = 0
	}
}

func (table *TableLogic) addFishTime(t int64) {
	for _, v := range table.Fishes {
		v.deadTime += t
	}
}

func (table *TableLogic) getSkillNum() {
	num := 0
	for _, v := range table.Fishes {
		if v.info.GetSkillId() > 0 {
			num++
		}
	}
}

func (table *TableLogic) test(buffer []byte, user player.PlayerInterface) {
	if global.GConfig.Runmode != "dev" {
		return
	}
	req := &msg.TestReq{}
	proto.Unmarshal(buffer, req)
	if req.GetFunc() == 1 && !table.FishTide {
		table.timer["fishTide"].Cancel()
		table.fishTideForecast()
	}
}

func (table *TableLogic) randInt(limit int, i int64) int {
	now := time.Now().UnixNano()
	if now == lastTime {
		diff++
		i += diff
	} else {
		diff = 0
	}
	lastTime = now
	r := rand2.New(rand2.NewSource(now + int64(i)*1000))
	return r.Intn(limit)
}

func (table *TableLogic) getLine(fishId string) ([]*msg.Point, []int32, int32, string, []int32, bool) {
	chance := config.GetFishChance(fishId)
	if !table.getChance(chance) {
		for i := 0; i < 100; i++ {
			line, speed, t, name, variant := config.GetConfLine(fishId, table.randInt, int64(i))
			if line != nil && name != nil && len(line) > 0 && speed != nil && len(speed) > 0 {
				if table.checkFishLine(name.(json.Number).String()) {
					continue
				}
				l, s, t := table.getConfLine(line, speed, t)
				v := table.getArray(variant)
				return l, s, t, name.(json.Number).String(), v, false
			}
		}

	}
	line := make([]*msg.Point, 0)
	len := 0
	if table.getChance(50) {
		len = 3
	} else {
		len = 5
	}
	directionX := false
	if table.getChance(100) {
		directionX = true
	}
	startAndEnd := table.getStartAndEnd(directionX)
	other := table.getOther(len-2, directionX)
	line = append(line, startAndEnd[0])
	line = append(line, other...)
	line = append(line, startAndEnd[1])
	if table.getChance(20) {
		line = table.reverse(line)
	}
	if 0 < line[0].X && line[0].X < int32(hight) && 0 > line[0].Y && table.getChance(70) {
		line = table.reverse(line)
	}
	return line, make([]int32, 0), 0, "", make([]int32, 0), true
}

func (table *TableLogic) checkFishLine(name string) bool {
	for _, v := range table.Fishes {
		if v.lineName == name {
			return true
		}
	}
	return false
}

func (table *TableLogic) getConfLine(lines []interface{}, speeds []interface{}, time interface{}) ([]*msg.Point, []int32, int32) {
	line := table.getPoint(lines)
	speed := table.getArray(speeds)

	t, _ := strconv.Atoi(time.(json.Number).String())
	return line, speed, int32(t)
}

func (table *TableLogic) getArray(speeds []interface{}) []int32 {
	speed := make([]int32, 0)
	for _, v := range speeds {
		s, _ := strconv.Atoi(v.(json.Number).String())
		speed = append(speed, int32(s))
	}
	return speed
}

func (table *TableLogic) getPoint(lines []interface{}) []*msg.Point {
	line := make([]*msg.Point, 0)
	for _, v := range lines {
		s := v.([]interface{})
		x, _ := strconv.Atoi(s[0].(json.Number).String())
		y, _ := strconv.Atoi(s[1].(json.Number).String())
		line = append(line, &msg.Point{
			X: int32(x),
			Y: int32(y),
		})
	}
	return line
}

func (table *TableLogic) getStartAndEnd(directionX bool) []*msg.Point {
	startAndEnd := make([]*msg.Point, 0)
	if directionX {
		startAndEnd = table.getPointOutSceneLR()
	} else {
		startAndEnd = table.getPointOutSceneUD()
	}

	return startAndEnd
}

func (table *TableLogic) getOther(num int, directionX bool) []*msg.Point {
	startAndEnd := make([]*msg.Point, 0)
	for i := 0; i < num; i++ {
		startAndEnd = append(startAndEnd, table.getPointInScene(int64(i)))
	}
	table.sort(startAndEnd, directionX)

	return startAndEnd
}

func (table *TableLogic) getPointOutSceneLR() []*msg.Point {
	points := make([]*msg.Point, 0)
	randY := rand.RandInt(0, weight)
	rand := rand.RandInt(0, (hightMax-hight)/2)
	x1 := -rand - fishWeight
	x2 := rand + hight + fishWeight
	y1 := randY
	change := weight / 2
	y2 := 0
	if randY > change {
		y2 = randY - change
	} else {
		y2 = randY + change
	}
	points = append(points, &msg.Point{X: int32(x1), Y: int32(y1)})
	points = append(points, &msg.Point{X: int32(x2), Y: int32(y2)})
	return points
}

func (table *TableLogic) getPointOutSceneUD() []*msg.Point {
	points := make([]*msg.Point, 0)
	randY := rand.RandInt(0, hight)
	rand := rand.RandInt(0, (weightMax-weight)/2)
	y1 := -rand - fishWeight
	y2 := rand + weight + fishWeight
	x1 := randY
	change := hight / 2
	x2 := 0
	if randY > change {
		x2 = randY - change
	} else {
		x2 = randY + change
	}
	points = append(points, &msg.Point{X: int32(x1), Y: int32(y1)})
	points = append(points, &msg.Point{X: int32(x2), Y: int32(y2)})
	return points
}

func (table *TableLogic) getPointInSceneLimit(limitX int, limitY int, i int64) *msg.Point {
	point := &msg.Point{}
	point.X = int32(rand.RandInt(limitX, hight-limitX))
	point.Y = int32(rand.RandInt(limitY, weight-limitY))
	return point
}

func (table *TableLogic) getPointInScene(i int64) *msg.Point {
	point := &msg.Point{}
	point.X = int32(table.randInt(hight, i))
	point.Y = int32(table.randInt(weight, i))
	return point
}

func (table *TableLogic) sort(points []*msg.Point, directionX bool) []*msg.Point {
	tem := make([]*msg.Point, 0)
	tem = append(tem, points...)
	if directionX {
		sort.Slice(tem, func(i, j int) bool {
			return tem[i].X < tem[j].X
		})
	} else {
		sort.Slice(tem, func(i, j int) bool {
			return tem[i].Y < tem[j].Y
		})
	}
	return tem
}

func (table *TableLogic) reverse(points []*msg.Point) []*msg.Point {
	tem := make([]*msg.Point, 0)
	tem = append(tem, points...)
	index := len(tem) / 2
	lastIndex := len(tem) - 1
	for i := 0; i < index; i++ {
		point := tem[i]
		tem[i] = tem[lastIndex-i]
		tem[lastIndex-i] = point
	}
	return tem
}

func (table *TableLogic) getChance(chance int32) bool {
	rand := int32(rand.RandInt(0, 100))
	if rand < chance {
		return true
	}
	return false
}

func (table *TableLogic) getChanceWan(chance int32, seed int64) bool {
	rand := int32(table.randInt(10000, seed))
	if rand < chance {
		return true
	}
	return false
}

func (table *TableLogic) getUniqueId() int32 {
	table.M.Lock()
	defer table.M.Unlock()
	//id := int32(0)
	//for i := 0; i < 10000; i++ {
	//	rand := int32(table.randInt( 10000, int64(i)))
	//	if table.Fishes[rand] == nil {
	//		id = rand
	//		break
	//	}
	//}
	table.uniqueId++
	return table.uniqueId
}

func (table *TableLogic) tick() {
	job, _ := table.Table.AddTimer(int64(fishTideForecastTime), func() {
		table.fishTideForecast()
	})
	table.timer["fishTide"] = job
	table.Table.AddTimer(int64(1000), func() {
		//table.refresh()
		table.Table.AddTimerRepeat(int64(config.GetRefreshTime()), 0, func() {
			if !table.IsDismiss {
				table.refresh()
			}
		})
		table.timeFishStart()
	})

	table.Table.AddTimerRepeat(int64(1000), 0, func() {
		if !table.IsDismiss {
			table.shotOff()
			table.checkFish()
		}
	})

	table.Table.AddTimerRepeat(int64(config.GetRoundTime()), 0, func() {
		table.Table.EndGame()
		table.Table.StartGame()
		table.SendLog()
	})

	table.Table.AddTimerRepeat(int64(config.GetSaveScoreTime()), 0, func() {
		table.saveScore()
	})

	table.Table.AddTimerRepeat(int64(1000), 0, table.kickOutAllUser)

	table.Table.AddTimerRepeat(int64(1000), 0, table.resetBulletNum)
}

func (table *TableLogic) resetBulletNum() {
	for _, v := range table.Users {
		v.BulletNum = 0
	}
}

//发送日志
func (table *TableLogic) SendLog() {
	for _, v := range table.Users {
		if !v.IsRobot {
			v.WriteLog()
		}
	}
}

func (table *TableLogic) saveScore() {
	for _, v := range table.Users {
		table.saveUserScore(v)
	}
}

func (table *TableLogic) saveUserScore(user *data.User) {
	if user.SubScore != 0 {
		table.updateScore(user, user.SubScore)
		user.InnerUser.SendChip(-user.SubScore)
		user.SubScore = 0
	}
	if user.AddScore != 0 {
		table.updateScore(user, user.AddScore)
		user.AddScore = 0
		user.TaxedScore = 0
	}
}

func (table *TableLogic) shotOff() {
	now := time.Now().UnixNano() / 1e6
	for _, u := range table.Users {
		if now-u.LastShootTime > int64(shotOffTime) { // !u.IsRobot &&
			table.shotOffGame(u)
		}
	}

}

func (table *TableLogic) shotOffGame(user *data.User) {
	//table.Table.Broadcast(int32(msg.MsgId_EXIST_ROOM_Res), &msg.ExistRoomRes{UserId:user.UserInfo.GetUserId(),})
	//delete(table.Users, user.UserInfo.GetUserId())

	table.userLeave(user.InnerUser)
	table.Table.KickOut(user.InnerUser)
	if user.IsRobot {
		//for i, robot := range table.Robots {
		//	if robot.AI.GetID() == user.UserInfo.GetUserId() {
		//		table.Robots = append(table.Robots[:i], table.Robots[i + 1:]...)
		//		break
		//	}
		//}
		delete(table.Robots, user.UserInfo.GetUserId())
	}
	if !user.IsRobot && !table.IsDismiss {
		table.changeCalculateRobot()
	}
	table.checkDismissTable()
	table.checkRobot()
}

func (table *TableLogic) refresh() {
	for k, v := range table.FishNum {
		if v < table.FishCap[k] && !table.Frozen && !table.FishTide {
			table.refreshFish(config.GetFishByType(k), true)
			break
		}
	}
	//num := len(table.Fishes)
	// table.Table.AddTimer(int64(config.GetRefreshTime()), func() {
	// 	if !table.IsDismiss {
	// 		table.refresh()
	// 	}
	// })
}

func (table *TableLogic) timeFishStart() {
	for _, v := range table.TimeFish {
		v.startTimer()
	}
}

func (table *TableLogic) fishTide() {
	fishTide := &msg.FishTideReq{
		Formation: 1,
		Line:      1,
	}
	table.Table.Broadcast(int32(msg.MsgId_FISHTIDE_Req), fishTide)
	table.createFormation()
	job, _ := table.Table.AddTimer(int64(yuchaochixushijian), func() {
		if table.IsDismiss {
			return
		}
		table.FishTide = false
		//table.allFishDead()
		table.sendFishTideEnd()
		//table.resetFish()
		table.timeFishStart()
		job, _ := table.Table.AddTimer(int64(fishTideForecastTime), func() {
			table.fishTideForecast()
		})
		table.timer["fishTide"] = job
	})
	table.timer["fishTide"] = job
}

func (table *TableLogic) allFishDead() {
	for _, v := range table.Fishes {
		res := &msg.DeadRes{
			Id: v.info.GetId(),
		}
		table.Table.Broadcast(int32(msg.MsgId_DEAD_Res), res)
	}
}

func (table *TableLogic) fishTideForecast() {
	table.FishTide = true
	table.Table.Broadcast(int32(msg.MsgId_FISHTIDEFORECAST_Req), &msg.FishTideForecastReq{})
	table.updateTableInfo()
	table.sendChangeScene()
	job, _ := table.Table.AddTimer(int64(fishTideTime), func() {
		if !table.IsDismiss {
			//table.allFishDead()
			//table.resetFish()
			table.fishTide()
		}
	})
	table.timer["fishTide"] = job
}

func (table *TableLogic) dead(buffer []byte, user player.PlayerInterface) {
	req := &msg.DeadReq{}
	proto.Unmarshal(buffer, req)
	fish := table.Fishes[req.GetId()]
	if fish == nil {
		return
	}
	table.checkFishNum(fish)
	delete(table.Fishes, req.GetId())
	res := &msg.DeadRes{
		Id: req.GetId(),
	}
	table.Table.Broadcast(int32(msg.MsgId_DEAD_Res), res)
	//table.checkLockFishId(req.GetId(), user)
}

func (table *TableLogic) changeModel(buffer []byte, user player.PlayerInterface) {
	req := &msg.ChangemModelReq{}
	proto.Unmarshal(buffer, req)
	user2 := table.Users[user.GetID()]
	if user2 == nil {
		return
	}
	modelId := req.GetModelId()
	if modelId == 1 {
		user2.UserInfo.Auto = !user2.UserInfo.GetAuto()
	}
	if modelId == 2 {
		user2.UserInfo.Speed = !user2.UserInfo.GetSpeed()
	}
	if modelId == 3 {
		user2.UserInfo.Lock = !user2.UserInfo.GetLock()
	}
	res := &msg.ChangemModelRes{
		UserId:  user.GetID(),
		ModelId: req.GetModelId(),
	}
	table.Table.Broadcast(int32(msg.MsgId_CHANGEMODEL_Res), res)
}

func (table *TableLogic) checkLockFishId(fishId int32, user player.PlayerInterface) {
	for _, u := range table.Users {
		if u.IsRobot && u.UserInfo.GetLockFishId() == fishId {
			u.UserInfo.LockFishId = table.getRobotTarget()
			res := &msg.ChangemLockFishRes{
				UserId: u.UserInfo.GetUserId(),
				FishId: u.UserInfo.GetLockFishId(),
			}
			table.Table.Broadcast(int32(msg.MsgId_CHANGELOCKFISH_Res), res)
			//break
		}
	}
}

func (table *TableLogic) changeLockFish(buffer []byte, user player.PlayerInterface) {
	req := &msg.ChangemLockFishReq{}
	proto.Unmarshal(buffer, req)
	user2 := table.Users[req.GetUserId()]
	if user2 == nil {
		return
	}
	fish := table.Fishes[req.GetFishId()]
	if fish != nil {
		y := fish.info.GetBornPoint().GetY()
		if user2.IsRobot && (y < 0 || y > int32(weight)) {
			return
		}
		user2.UserInfo.LockFishId = req.GetFishId()
	}
	res := &msg.ChangemLockFishRes{
		UserId: req.GetUserId(),
		FishId: user2.UserInfo.GetLockFishId(),
	}
	table.Table.Broadcast(int32(msg.MsgId_CHANGELOCKFISH_Res), res)
}

func (table *TableLogic) checkRobot() {
	if table.IsDismiss {
		return
	}
	length := len(table.Users)
	if length == 2 && table.getRealUserNum() == 2 { //
		r := rand.RandInt(0, 5000)
		table.Table.AddTimer(int64(r), func() {
			table.addRobot()
		})

	}
	if length == 1 && table.getRealUserNum() == 1 {
		num := 1
		if table.getChance(50) {
			num = 2
		}
		for i := 0; i < num; i++ {
			table.Table.AddTimer(int64(table.randInt(5000, int64(i))), table.addRobot)
		}
	}
	if length == 4 && table.getRealUserNum() == 2 {
		table.Table.AddTimer(int64(table.randInt(5000, int64(0))), table.subRobot)
	}
}

func (table *TableLogic) getRealUserNum() int {
	num := 0
	for _, u := range table.Users {
		if !u.IsRobot {
			num++
		}
	}
	return num
}

func (table *TableLogic) addRobot() {
	if table.IsDismiss {
		return
	}
	if len(table.Users) > 3 || len(table.Robots) > 1 {
		return
	}
	err := table.Table.GetRobot(1)
	if err != nil {
		//log.Traceln("GET robot err", err)
	}
}

func (table *TableLogic) checkRobotBehaviour(fishId int32) {
	for _, v := range table.Robots {
		if v.TargetFishId == fishId {
			v.changeModel(2)
			v.getChangeTarget(-1)
		}
	}
}

func (table *TableLogic) changeRobotBehaviour(fishId int32) {
	for _, v := range table.Robots {
		if v.TargetFishId == -1 && table.getChance(int32(config.GetRobotLockChance())) {
			v.changeModel(2)
			v.getChangeTarget(fishId)
		}

	}
}

func (table *TableLogic) BindRobot(user inter.AIUserInter) player.RobotHandler {
	robot := NewRobot(table.Table)
	robot.AI = user
	return robot
	//table.Robots[robot.AI.GetID()] = robot
	//user.SendMsg(int32(msg.MsgId_ZERO), &msg.Fish{})
	//robot.init()
	//robot.enterRoom()
	//table.getLimitPoint(robot.LastShoot, int(config.GetRobotLimit()))
	//table.Table.AddTimer(int64(1000), func() {
	//	robot.changeRotateModel(table.getChance(50))
	//	table.checkRobotBulletLv(robot)
	//	table.robotChangeShootPoint(robot)
	//	table.checkRobotShootModel(robot)
	//	table.robotShoot(robot)
	//})

	//table.Table.AddTimer(int64(config.GetRobotQuitTime()), func() {
	//	table.robotQuit(robot.AI.GetID())
	//})
	//
	//table.Table.AddTimer(int64(5000), func() {
	//	table.robotCheckCoinQuit(robot)
	//})
}

func (table *TableLogic) robotCheckCoinQuit(robot *Robot) {
	if robot == nil {
		return
	}
	score := robot.AI.GetScore()
	limit := robot.FirstCoin * int64(config.GetRobotQuitCoin()) / 100
	if score > limit {
		table.robotQuit(robot.AI.GetID())
		return
	}
	table.Table.AddTimer(int64(5000), func() {
		table.robotCheckCoinQuit(robot)
	})
}

func (table *TableLogic) robotQuit(userId int64) {
	//robot := table.getRobot(userId)
	//if robot != nil {
	//	robot.leaveGame()
	//	table.Robots[userId] = nil
	//	delete(table.Robots, userId)
	//}
	user := table.Users[userId]
	if user != nil {
		table.shotOffGame(user)
	}
}

func (table *TableLogic) getRobot(userId int64) *Robot {
	for _, robot := range table.Robots {
		if robot.AI.GetID() == userId {
			return robot
		}
	}

	return nil
}

func (table *TableLogic) checkRobotLockFish(robot *Robot) {
	fishId := table.getRobotTarget()
	if fishId != -1 {
		robot.getChangeTarget(fishId)
	}
	table.Table.AddTimer(int64(config.GetRobotLockTime()), func() {
		table.checkRobotLockFish(robot)
	})
}

func (table *TableLogic) checkRobotBulletLv(robot *Robot) {
	if robot == nil {
		return
	}
	score := robot.AI.GetScore()
	firstCoin := robot.FirstCoin
	if score == firstCoin {
		initNum := config.GetRobotInitShootNum()
		robot.upgradeBulletLv(initNum, initNum, table.getChance(config.GetRobotIsFixed()))
	}
	winCoin := firstCoin * config.GetRobotWinChangeCoin() / 100
	loseCoin := firstCoin * config.GetRobotLoseChangeCoin() / 100
	if score > firstCoin && score-firstCoin > winCoin {
		limit := config.GetRobotWinChangeLimit()
		low, _ := strconv.Atoi(limit[0].(json.Number).String())
		up, _ := strconv.Atoi(limit[1].(json.Number).String())
		robot.upgradeBulletLv(low, up, table.getChance(config.GetRobotIsFixed()))
	}
	if score < firstCoin && firstCoin-score > loseCoin {
		limit := config.GetRobotLoseChangeLimit()
		low, _ := strconv.Atoi(limit[0].(json.Number).String())
		up, _ := strconv.Atoi(limit[1].(json.Number).String())
		robot.upgradeBulletLv(low, up, table.getChance(config.GetRobotIsFixed()))
	}
}

func (table *TableLogic) getRobotTarget() int32 {
	targetType := msg.Type_SMALL
	r := rand.RandInt(0, 100)
	if r < 15 {
		return table.getFishIdByType(targetType)
	}
	if r < 35 {
		targetType = msg.Type_MIDDLE
		return table.getFishIdByType(targetType)
	}
	if r < 60 {
		targetType = msg.Type_BIG
		return table.getFishIdByType(targetType)
	}
	if r < 100 {
		targetType = msg.Type_BOSS
		return table.getFishIdByType(targetType)
	}
	return -1
}

func (table *TableLogic) getFishIdById(fishId string) int32 {
	for k, fish := range table.Fishes {
		//t := time.Now().UnixNano() / 1e6 - fish.info.BornTime
		t := fish.deadTime - time.Now().UnixNano()/1e6
		if fish.info.GetFishId() == fishId && t >= 5000 && t <= 18000 {
			return k
		}
	}
	return -1
}

func (table *TableLogic) getFishIdByType(targetType msg.Type) int32 {
	for k, fish := range table.Fishes {
		//t := time.Now().UnixNano() / 1e6 - fish.info.BornTime
		t := fish.deadTime - time.Now().UnixNano()/1e6
		if fish.info.GetTypeId() == targetType && t >= 5000 && t <= 18000 {
			return k
		}
	}
	return -1
}

func (table *TableLogic) robotShoot(robot *Robot) {
	if robot == nil {
		return
	}
	user := table.Users[robot.AI.GetID()]
	if user == nil { //&& user.UserInfo.LockFishId == -1
		return
	}
	robot.startShoot()
	//r := rand.RandInt(0,1000) + 500
	t := table.getRobotRestTime(robot)
	table.Table.AddTimer(int64(robot.ShootTime+int32(t)), func() {
		//if user.UserInfo.LockFishId == -1 {
		//	fishId := table.getRobotTarget()
		//	if fishId != -1 {
		//		robot.getChangeTarget(fishId)
		//	}
		//}
		//if robot.TargetFishId == -1 {
		table.robotShoot(robot)
		//}
	})
}

func (table *TableLogic) getRobotRestTime(robot *Robot) int {
	if robot.ShootModel == 1 || robot.TargetFishId == -1 {
		return 0
	}
	now := time.Now().UnixNano() / 1e6
	lastRestTime := robot.LastRestTime
	t := int64(config.GetRobotRestSpace())
	if now-lastRestTime >= t {
		return config.GetRobotRestTime()
	}
	return 0
}

func (table *TableLogic) checkRobotShootModel(robot *Robot) {
	if robot == nil {
		return
	}
	user := table.Users[robot.AI.GetID()]
	if user == nil {
		return
	}
	bullet := int32(0)
	if user != nil && user.UserInfo.GetLockFishId() == -1 {
		bullet = user.UserInfo.GetBulletLv()
		num := robot.AI.GetScore() / int64(config.GetFishBet(table.Table.GetAdviceConfig(), bullet))
		limit := config.GetRobotShootLimit()
		if num >= int64(limit) {
			robot.ShootModel = 1
		}
		if num < int64(limit) {
			robot.ShootModel = 2
		}

	}
	table.Table.AddTimer(int64(5000), func() {
		table.checkRobotShootModel(robot)
	})

}

func (table *TableLogic) robotChangeShootPoint(robot *Robot) {
	if robot == nil {
		return
	}
	point := &msg.Point{}
	if robot.Model == 1 {
		point = table.getPointInSceneLimit(0, 100, time.Now().UnixNano())
		robot.changeShootPoint(point)
		table.Table.AddTimer(int64(config.GetRobotChangeTime()), func() {
			table.robotChangeShootPoint(robot)
		})
		return
	}
	if robot.Model == 2 {
		point = table.getLimitPoint(robot.LastShoot, int(config.GetRobotLimit()))
		robot.changeShootPoint(point)
		table.Table.AddTimer(int64(config.GetRobotRandChangeTime()), func() {
			table.robotChangeShootPoint(robot)
		})
	}
}

func (table *TableLogic) getLimitPoint(lastPoint *msg.Point, limit int) *msg.Point {
	if lastPoint == nil {
		return table.getPointInSceneLimit(0, 100, time.Now().UnixNano())
	}
	point := &msg.Point{}
	seedX := int(lastPoint.X)
	point.X = int32(rand.RandInt(seedX-limit, seedX+limit))
	point.Y = int32(rand.RandInt(0, weight))
	return point
}

func (table *TableLogic) subRobot() {
	//if len(table.Robots) > 0 {
	//	r := rand.RandInt(0, 5000)
	//	table.Table.AddTimer(int64(r), func() {
	//		for k, v := range table.Robots {
	//			v.leaveGame()
	//			v = nil
	//			delete(table.Robots, k)
	//			break
	//		}
	//
	//	})
	//}
	for _, v := range table.Users {
		if v.IsRobot {
			table.shotOffGame(v)
			return
		}
	}
}

func (table *TableLogic) checkIsAllRobot() bool {
	for _, user := range table.Users {
		if !user.IsRobot {
			return false
		}
	}
	return true
}

func (table *TableLogic) getBezierLineTime(speed int32, line []*msg.Point) int32 {
	return table.getBezierLineLength(table.getBezierLine(100, line))*1000/speed + 5000
}

func (table *TableLogic) getBezierLinePartLength(start int, end int, line []*msg.Point) int32 {
	l := float64(0)
	//length := end - 1
	if start > end || end > len(line) {
		return 0
	}
	for i := start; i < end; i++ {
		l += table.getDistance(line[i], line[i+1])
	}
	return int32(l)
}

func (table *TableLogic) getBezierLineLength(line []*msg.Point) int32 {
	l := float64(0)
	length := len(line) - 1
	for i := 0; i < length; i++ {
		l += table.getDistance(line[i], line[i+1])
	}
	return int32(l)
}

func (table *TableLogic) getBezierLine(num int, controlPoints []*msg.Point) []*msg.Point {
	line := make([]*msg.Point, 0)
	t := float64(0)
	diff := 1 / float64(num)
	for i := 0; i < num; i++ {
		line = append(line, table.createBezierPoint(t, controlPoints))
		t += diff
	}
	return line
}

func (table *TableLogic) createBezierPoint(t float64, controlPoints []*msg.Point) *msg.Point {
	x := float64(0)
	y := float64(0)
	tem := make([]*msg.Point, 0)
	tem = append(tem, controlPoints...)
	length := len(tem) - 1
	for i := 0; i < len(tem); i++ {
		p := tem[i]
		m := table.factorial(length) / table.factorial(i) / table.factorial(length-i) *
			math.Pow(1-t, float64(length-i)) * math.Pow(t, float64(i))
		x += float64(p.X) * m
		y += float64(p.Y) * m
	}
	return &msg.Point{
		X: int32(x),
		Y: int32(y),
	}
}

func (table *TableLogic) factorial(num int) float64 {
	if num <= 1 {
		return 1
	}
	return float64(num) * table.factorial(num-1)
}

func (table *TableLogic) isCollide(points []*msg.Point, point *msg.Point, r float64) bool {

	for _, p := range points {
		if table.getDistance(p, point) <= r {
			//return true
		}
	}
	mp, sides, length, width := table.getMiddlePoint(points)
	side := table.getIntersectSide(sides, mp, point)
	cl := table.getCompareLength(side, length, width)
	collide := false
	if table.getDistance(mp, point) <= float64(cl/table.getCos(side, mp, point))+r {
		collide = true
	}
	return collide
}

func (table *TableLogic) getDistance(a *msg.Point, b *msg.Point) float64 {
	return math.Sqrt(math.Pow(float64(a.X-b.X), 2) + math.Pow(float64(a.Y-b.Y), 2))
}

//获取向量
func (table *TableLogic) getVector(a *msg.Point, b *msg.Point) *msg.Point {
	return &msg.Point{
		X: a.X - b.X,
		Y: a.Y - b.Y,
	}
}

//获取点乘
func (table *TableLogic) getProduct(a *msg.Point, b *msg.Point) int32 {
	return a.X*b.X + a.Y*b.Y
}

//获取中点
func (table *TableLogic) getMiddlePoint(points []*msg.Point) (*msg.Point, [][]*msg.Point, int32, int32) {
	tem := make([]*msg.Point, 0)
	tem = append(tem, points...)
	point := &msg.Point{}
	side := make([][]*msg.Point, 0)
	length := int32(0)
	width := int32(0)
	sure := table.getProduct(table.getVector(tem[0], tem[1]), table.getVector(tem[0], tem[2])) == 0
	if sure {
		point = &msg.Point{
			X: (tem[1].X + tem[2].X) / 2,
			Y: (tem[1].Y + tem[2].Y) / 2,
		}
		side = append(side, table.getASide(tem[0], tem[1]))
		side = append(side, table.getASide(tem[0], tem[2]))
		side = append(side, table.getASide(tem[1], tem[3]))
		side = append(side, table.getASide(tem[2], tem[3]))
		length = int32(table.getDistance(tem[0], tem[1]))
		width = int32(table.getDistance(tem[0], tem[2]))
	}
	if !sure {
		if table.getDistance(tem[0], tem[1]) > table.getDistance(tem[0], tem[2]) {
			point = &msg.Point{
				X: (tem[0].X + tem[1].X) / 2,
				Y: (tem[0].Y + tem[1].Y) / 2,
			}
			side = append(side, table.getASide(tem[0], tem[2]))
			side = append(side, table.getASide(tem[0], tem[3]))
			side = append(side, table.getASide(tem[1], tem[2]))
			side = append(side, table.getASide(tem[1], tem[3]))
			length = int32(table.getDistance(tem[0], tem[3]))
			width = int32(table.getDistance(tem[0], tem[2]))
		} else {
			point = &msg.Point{
				X: (tem[0].X + tem[2].X) / 2,
				Y: (tem[0].Y + tem[2].Y) / 2,
			}
			side = append(side, table.getASide(tem[0], tem[1]))
			side = append(side, table.getASide(tem[0], tem[3]))
			side = append(side, table.getASide(tem[1], tem[2]))
			side = append(side, table.getASide(tem[2], tem[3]))
			length = int32(table.getDistance(tem[0], tem[1]))
			width = int32(table.getDistance(tem[0], tem[3]))
		}

	}
	return point, side, length, width
}

//判断线段ab与cd是否相交
func (table *TableLogic) isIntersect(a *msg.Point, b *msg.Point, c *msg.Point, d *msg.Point) bool {
	ac := table.getVector(c, a)
	ad := table.getVector(d, a)
	ab := table.getVector(b, a)
	cad := float64(table.getProduct(ac, ad)) / (table.getDistance(c, a) * table.getDistance(d, a))
	cab := float64(table.getProduct(ac, ab)) / (table.getDistance(c, a) * table.getDistance(b, a))
	cac := float64(table.getProduct(ad, ab)) / (table.getDistance(b, a) * table.getDistance(d, a))
	return cac > cad && cab > cad
}

func (table *TableLogic) getASide(a *msg.Point, b *msg.Point) []*msg.Point {
	side := make([]*msg.Point, 0)
	side = append(side, a)
	side = append(side, b)
	return side
}

func (table *TableLogic) getIntersectSide(sides [][]*msg.Point, m1 *msg.Point, m2 *msg.Point) []*msg.Point {
	tem := make([][]*msg.Point, 0)
	tem = append(tem, sides...)
	side := make([]*msg.Point, 0)
	for _, v := range tem {
		if table.isIntersect(v[0], v[1], m1, m2) {
			side = append(side, v...)
		}
	}
	return side
}

func (table *TableLogic) getCompareLength(side []*msg.Point, length int32, width int32) int32 {
	len := int32(table.getDistance(side[0], side[1]))
	cl := length / 2
	if len == length {
		cl = width / 2
	}
	return cl
}

func (table *TableLogic) getCos(side []*msg.Point, m1 *msg.Point, m2 *msg.Point) int32 {
	cos := table.getProduct(table.getVector(side[0], side[1]), table.getVector(m1, m2))
	if cos > 0 {
		return 2 * cos
	}
	return 2 * table.getProduct(table.getVector(side[1], side[0]), table.getVector(m1, m2))
}
