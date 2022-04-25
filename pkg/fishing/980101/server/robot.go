package server

import (
	"encoding/json"
	"go-game-sdk/example/game_buyu/980101/config"
	"go-game-sdk/example/game_buyu/980101/msg"
	"go-game-sdk/inter"
	"strconv"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Robot struct {
	BulletNum    int32
	Table        table.TableInterface
	AI           inter.AIUserInter
	TargetFishId int32
	ShootTime    int32
	FixBulletLv  bool
	LastShoot    *msg.Point
	Model        int32
	ShootModel   int32
	FirstCoin    int64
	LastRestTime int64
	BulletLv     int32
	Speed        bool
	Fishes       sync.Map
	FisheIds     sync.Map
	//Lock         *sync.RWMutex
}

func NewRobot(table table.TableInterface) *Robot {
	robot := &Robot{}
	//err := table.GetRobot()
	//if err != nil {
	//	//fmt.Println(err)
	//	return nil
	//}
	robot.Table = table
	//robot.AI = user
	robot.TargetFishId = -1
	robot.ShootTime = 400
	robot.BulletLv = 1
	return robot
}

func (robot *Robot) init() {
	robot.FirstCoin = robot.AI.GetScore()
	robot.LastRestTime = time.Now().UnixNano() / 1e6
	robot.TargetFishId = -1
	robot.Fishes = sync.Map{}
	robot.FisheIds = sync.Map{}
	//robot.Lock = new(sync.RWMutex)
}

func (robot *Robot) enterRoom() {
	req := &msg.EnterRoomReq{
		UserId: robot.AI.GetID(),
	}
	robot.AI.SendMsgToServer(int32(msg.MsgId_INTO_ROOM_Req), req)
}

func (robot *Robot) start() {
	robot.AI.AddTimer(int64(1000), func() {
		robot.changeRotateModel(robot.getChance(50))
		robot.checkRobotBulletLv()
		robot.checkRobotShootModel()
		robot.robotChangeShootPoint()
		t := robot.getRobotRestTime()
		robot.AI.AddTimer(int64(robot.ShootTime+int32(t)), func() {
			robot.startShoot()
		})
	})
	robot.AI.AddTimerRepeat(int64(5000), 0, func() {
		robot.checkRobotBulletLv()
	})

	robot.AI.AddTimer(int64(5000), func() {
		robot.changeModel(3)
		id := robot.getLockFishId()
		robot.changeRobotBehaviour(id)
	})

	robot.AI.AddTimerRepeat(int64(5000), 0, func() {
		robot.checkRobotShootModel()
	})

	robot.AI.AddTimerRepeat(int64(config.GetRobotChangeTime()), 0, func() {
		robot.robotChangeShootPoint()
	})

	robot.AI.AddTimer(int64(config.GetRobotQuitTime()), func() {
		robot.leaveGame()
	})

	robot.AI.AddTimerRepeat(int64(5000), 0, func() {
		robot.robotCheckCoinQuit()
	})

	robot.AI.AddTimerRepeat(int64(config.GetRobotSpeedCheckTime()), 0, func() {
		robot.checkSpeedModel()
	})

	robot.AI.AddTimerRepeat(int64(config.GetRobotLockCheckTime()), 0, func() {
		id := robot.getLockFishId()
		robot.getChangeTarget(id)
	})
}

func (robot *Robot) getRobotRestTime() int {
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

func (robot *Robot) checkRobotShootModel() {
	if robot == nil {
		return
	}
	bullet := int32(0)
	if robot != nil && robot.BulletLv == -1 {
		bullet = robot.BulletLv
		num := robot.AI.GetScore() / int64(config.GetFishBet(strconv.Itoa(int(robot.Table.GetLevel())), bullet))
		limit := config.GetRobotShootLimit()
		if num >= int64(limit) {
			robot.ShootModel = 1
		}
		if num < int64(limit) {
			robot.ShootModel = 2
		}

	}

}

func (robot *Robot) robotChangeShootPoint() {
	if robot == nil {
		return
	}
	point := &msg.Point{}
	if robot.Model == 1 {
		point = robot.getPointInSceneLimit(0, 100, time.Now().UnixNano())
		robot.changeShootPoint(point)
		return
	}
	if robot.Model == 2 {
		point = robot.getLimitPoint(robot.LastShoot, int(config.GetRobotLimit()))
		robot.changeShootPoint(point)
	}
}

func (robot *Robot) getPointInSceneLimit(limitX int, limitY int, i int64) *msg.Point {
	point := &msg.Point{}
	point.X = int32(rand.RandInt(limitX, hight-limitX))
	point.Y = int32(rand.RandInt(limitY, weight-limitY))
	return point
}

func (robot *Robot) getLimitPoint(lastPoint *msg.Point, limit int) *msg.Point {
	if lastPoint == nil {
		return robot.getPointInSceneLimit(0, 100, time.Now().UnixNano())
	}
	point := &msg.Point{}
	seedX := int(lastPoint.X)
	point.X = int32(rand.RandInt(seedX-limit, seedX+limit))
	point.Y = int32(rand.RandInt(100, weight-100))
	return point
}

func (robot *Robot) checkRobotBulletLv() {
	if robot == nil {
		return
	}
	score := robot.AI.GetScore()
	firstCoin := robot.FirstCoin
	if score == firstCoin {
		initNum := config.GetRobotInitShootNum()
		robot.upgradeBulletLv(initNum, initNum, robot.getChance(config.GetRobotIsFixed()))
	}
	winCoin := firstCoin * config.GetRobotWinChangeCoin() / 100
	loseCoin := firstCoin * config.GetRobotLoseChangeCoin() / 100
	if score > firstCoin && score-firstCoin > winCoin {
		limit := config.GetRobotWinChangeLimit()
		low, _ := strconv.Atoi(limit[0].(json.Number).String())
		up, _ := strconv.Atoi(limit[1].(json.Number).String())
		robot.upgradeBulletLv(low, up, robot.getChance(config.GetRobotIsFixed()))
	}
	if score < firstCoin && firstCoin-score > loseCoin {
		limit := config.GetRobotLoseChangeLimit()
		low, _ := strconv.Atoi(limit[0].(json.Number).String())
		up, _ := strconv.Atoi(limit[1].(json.Number).String())
		robot.upgradeBulletLv(low, up, robot.getChance(config.GetRobotIsFixed()))
	}
}

func (robot *Robot) getChance(chance int32) bool {
	rand := int32(rand.RandInt(0, 100))
	if rand < chance {
		return true
	}
	return false
}

func (robot *Robot) getChanceWan(chance int) bool {
	rand := rand.RandInt(0, 10000)
	if rand < chance {
		return true
	}
	return false
}

func (robot *Robot) robotCheckCoinQuit() {
	if robot == nil {
		return
	}
	score := robot.AI.GetScore()
	limit := robot.FirstCoin * int64(config.GetRobotQuitCoin()) / 100
	if score > limit {
		robot.leaveGame()
	}
}

func (robot *Robot) startShoot() {
	if robot.BulletNum >= 15 {
		t := robot.getRobotRestTime()
		robot.AI.AddTimer(int64(robot.ShootTime+int32(t)), func() {
			robot.startShoot()
		})
		return
	}
	req := &msg.ShootReq{
		UserId: robot.AI.GetID(),
		Point:  robot.LastShoot,
	}
	robot.AI.SendMsgToServer(int32(msg.MsgId_SHOOT_Req), req)
	//fmt.Println("shoot =", robot.LastShoot, robot.TargetFishId, robot.Fishes[robot.TargetFishId])
	t := robot.getRobotRestTime()
	robot.AI.AddTimer(int64(robot.ShootTime+int32(t)), func() {
		robot.startShoot()
	})
}

func (robot *Robot) changeShootPoint(point *msg.Point) {
	//fmt.Println("shoot point", point.X, point.Y)
	robot.LastShoot = point
}

func (robot *Robot) getCoinDivideBulletLv(bulletLv int64) int64 {
	return robot.AI.GetScore() / bulletLv
}

func (robot *Robot) getMaxShootNum(bet int64) int64 {
	return robot.AI.GetScore() / bet
}

func (robot *Robot) changeShootModel(shootModel int32) {
	robot.ShootModel = shootModel
}

func (robot *Robot) changeRotateModel(isRandom bool) {
	if isRandom {
		robot.Model = 1
		return
	}
	robot.Model = 2
}

func (robot *Robot) changeModel(model int32) {
	req := &msg.ChangemModelReq{
		UserId:  robot.AI.GetID(),
		ModelId: model,
	}
	robot.AI.SendMsgToServer(int32(msg.MsgId_CHANGEMODEL_Req), req)
}

func (robot *Robot) upgradeBulletLv(low, up int, isFixed bool) {
	lv := robot.getBulletLv(low, up)
	num := robot.getChangeNum(lv)
	for i := 0; i < num; i++ {
		req := &msg.UpgradeReq{
			IsAdd: true,
		}
		robot.AI.SendMsgToServer(int32(msg.MsgId_UPGRADE_Req), req)
		robot.BulletLv++
		if robot.BulletLv > 10 {
			robot.BulletLv = 1
		}
	}
	if isFixed {
		robot.FixBulletLv = true
	}
}

func (robot *Robot) getChangeNum(lv int) int {
	bulletLv := int(robot.BulletLv)
	if lv > bulletLv {
		return lv - bulletLv
	}
	if lv < bulletLv {
		return lv + 10 - bulletLv
	}
	return 0
}

func (robot *Robot) getChangeTarget(fishId int32) {
	req := &msg.ChangemLockFishReq{
		UserId: robot.AI.GetID(),
		FishId: fishId,
	}
	robot.TargetFishId = fishId
	robot.AI.SendMsgToServer(int32(msg.MsgId_CHANGELOCKFISH_Req), req)
}

func (robot *Robot) getBulletLv(low, up int) int {
	//r := rand.RandInt(0, 100)
	//if r < 5 {
	//	return 1
	//}
	//
	//if r < 10 {
	//	return 2
	//}
	//
	//if r < 15 {
	//	return 3
	//}
	//
	//if r < 25 {
	//	return 4
	//}
	//
	//if r < 35 {
	//	return 5
	//}
	//
	//if r < 45 {
	//	return 6
	//}
	//
	//if r < 55 {
	//	return 7
	//}
	//
	//if r < 70 {
	//	return 8
	//}
	//
	//if r < 85 {
	//	return 9
	//}
	//
	//if r < 100 {
	//	return 10
	//}
	score := robot.AI.GetScore()
	lv := make([]int, 0)
	for i := 1; i < 11; i++ {
		bet := config.GetFishBet(robot.Table.GetAdviceConfig(), int32(i))
		if bet == 0 {
			break
		}
		num := int(score / int64(bet))
		if num >= low && num <= up {
			lv = append(lv, i)
		}
		if num > up {
			lv = append([]int{}, i)
		}
	}
	if len(lv) == 0 {
		return 1
	}
	return lv[rand.RandInt(0, len(lv))]
}

func (robot *Robot) leaveGame() {
	//req := &msg.ExistRoomReq{
	//	UserId:  robot.AI.GetID(),
	//}
	//robot.AI.SendMsgToServer(int32(msg.MsgId_EXIST_ROOM_Req), req)
	if robot.AI == nil {
		return
	}
	robot.AI.LeaveRoom()
}

func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case int32(msg.MsgId_SHOOT_Res):
		robot.shoot(buffer)
		break
	case int32(msg.MsgId_HIT_Res):
		robot.hit(buffer)
		break
	case int32(msg.MsgId_SKILLHIT_Res):
		robot.skillHit(buffer)
		break
	case int32(msg.MsgId_DEAD_Res):
		robot.dead(buffer)
		break
	case int32(msg.MsgId_UPGRADE_Res):
		robot.upgrade(buffer)
		break
	case int32(msg.MsgId_CHANGEMODEL_Res):
		robot.modelChange(buffer)
		break
	case int32(msg.MsgId_REFRESHFISH_Req):
		robot.refresh(buffer)
		break
	case int32(msg.MsgId_ZERO):
		robot.init()
		robot.enterRoom()
		robot.start()
		break
	}
}

func (robot *Robot) shoot(buffer []byte) {
	res := &msg.ShootRes{}
	proto.Unmarshal(buffer, res)
	if robot.AI.GetID() == res.GetUserId() {
		robot.BulletNum++
	}
}

func (robot *Robot) skillHit(buffer []byte) {
	res := &msg.SkillHitRes{}
	proto.Unmarshal(buffer, res)
	for _, v := range res.GetFishIds() {
		if v.GetScore() > 0 {
			robot.checkRobotBehaviour(v.GetFishId())
		}
	}
}

func (robot *Robot) dead(buffer []byte) {
	res := &msg.DeadRes{}
	proto.Unmarshal(buffer, res)
	robot.checkRobotBehaviour(res.GetId())
}

func (robot *Robot) hit(buffer []byte) {
	res := &msg.HitRes{}
	proto.Unmarshal(buffer, res)
	if robot.AI.GetID() == res.GetUserId() {
		robot.BulletNum--
		if res.GetFish().GetScore() > 0 {
			robot.checkRobotBehaviour(res.GetFish().GetFishId())
			//robot.checkRobotBulletLv()
		}
	}
}

func (robot *Robot) checkRobotBehaviour(fishId int32) {
	//robot.Lock.Lock()
	//delete(robot.Fishes, fishId)
	//delete(robot.FisheIds, fishId)
	//robot.Lock.Unlock()
	robot.Fishes.Delete(fishId)
	robot.FisheIds.Delete(fishId)
	if robot.TargetFishId == fishId {
		//robot.changeModel(2)
		id := robot.getLockFishId()
		robot.getChangeTarget(id)
	}
}

func (robot *Robot) refresh(buffer []byte) {
	res := &msg.RefreshFishReq{}
	proto.Unmarshal(buffer, res)
	//robot.Lock.Lock()
	for _, v := range res.GetFish() {
		//if config.GetRobotLockFishes(v.GetFishId()) {
		//	robot.changeRobotBehaviour(v.GetId())
		//}
		robot.Fishes.Store(v.GetId(), v.GetBornTime())
		robot.FisheIds.Store(v.GetId(), v.GetFishId())
	}
	//robot.Lock.Unlock()

}

func (robot *Robot) checkSpeedModel() {
	speed := robot.getChanceWan(config.GetRobotSpeedChance())
	//fmt.Println(speed, robot.Speed)
	if (speed && !robot.Speed) || (!speed && robot.Speed) {
		robot.changeModel(2)
	}
}

func (robot *Robot) changeRobotBehaviour(fishId int32) { //&& robot.getChance(int32(config.GetRobotLockChance()))
	//if robot.TargetFishId == -1 {
	//robot.changeModel(2)
	robot.getChangeTarget(fishId)
	//}
}

func (robot *Robot) getLockFishId() int32 {
	id := int32(-1)
	//robot.Lock.Lock()
	//for k, fish := range robot.Fishes {
	//	t := time.Now().UnixNano() / 1e6 - fish
	//	//t := fish.deadTime - time.Now().UnixNano()/1e6
	//	if t >= 8000 && t <= 18000 && config.GetRobotLockFishes(robot.FisheIds[k]) {
	//		robot.Lock.Unlock()
	//		return k
	//	}
	//	id = k
	//}
	//robot.Lock.Unlock()
	robot.Fishes.Range(func(key, value interface{}) bool {
		t := time.Now().UnixNano()/1e6 - value.(int64)
		//t := fish.deadTime - time.Now().UnixNano()/1e6
		id = key.(int32)
		f, _ := robot.FisheIds.Load(key)
		if f == nil {
			return false
		}
		if t >= 8000 && t <= 18000 && config.GetRobotLockFishes(f.(string)) {
			return true
		}
		return false
	})
	return id
}

func (robot *Robot) upgrade(buffer []byte) {

}

func (robot *Robot) modelChange(buffer []byte) {
	res := &msg.ChangemModelRes{}
	proto.Unmarshal(buffer, res)
	if robot.AI.GetID() == res.GetUserId() && res.GetModelId() == 2 {
		if robot.ShootTime == 400 {
			robot.ShootTime /= 2
			robot.Speed = true
		} else if robot.ShootTime != 400 {
			robot.ShootTime *= 2
			robot.Speed = false
		}
	}
}
