package server

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/fishing/980401/config"
	"github.com/kubegames/kubegames-games/pkg/fishing/980401/msg"
	"github.com/kubegames/kubegames-games/pkg/fishing/980401/tools"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Robot struct {
	BulletNum    int32
	Table        table.TableInterface
	AI           player.RobotInterface
	TargetFishId int32
	ShootTime    int32
	FixBulletLv  bool
	LastShoot    *msg.Point
	Model        int32
	ShootModel   int32
	FirstCoin    int64
	LastRestTime int64
	BulletLv     int32
	SkillNum     map[int32]int
	BossTime     bool
	CanShoot     bool
	Fishes       sync.Map
	FisheIds     sync.Map
	//Lock         *sync.RWMutex
	IsLock bool
}

func NewRobot(table table.TableInterface) *Robot {
	robot := &Robot{}
	//err := table.GetRobot()
	//if err != nil {
	//	log.Traceln(err)
	//	return nil
	//}
	robot.Table = table
	//robot.AI = user
	robot.TargetFishId = -1
	robot.ShootTime = 200
	return robot
}

func (robot *Robot) init() {
	robot.FirstCoin = robot.AI.GetScore()
	robot.LastRestTime = time.Now().UnixNano() / 1e6
	robot.TargetFishId = -1
	robot.BulletLv = 1
	robot.CanShoot = true
	robot.SkillNum = make(map[int32]int, 0)
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

func (robot *Robot) startShoot() {
	if !robot.CanShoot || robot.BulletNum >= 15 {
		return
	}
	req := &msg.ShootReq{
		UserId: robot.AI.GetID(),
		Point:  robot.LastShoot,
	}
	req.BulletType = robot.checkBulletType()
	robot.AI.SendMsgToServer(int32(msg.MsgId_SHOOT_Req), req)
}

func (robot *Robot) checkBulletType() int32 {
	if robot.SkillNum[5] > 0 {
		robot.SkillNum[5]--
		return 1
	}
	return 0
}

func (robot *Robot) changeShootPoint(point *msg.Point) {
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

func (robot *Robot) refresh(buffer []byte) {
	res := &msg.RefreshFishReq{}
	proto.Unmarshal(buffer, res)
	//robot.Lock.Lock()
	for _, v := range res.GetFish() {
		//if config.GetRobotLockFishes(v.GetFishId()) {
		//	robot.changeRobotBehaviour(v.GetID())
		//}
		robot.Fishes.Store(v.GetId(), v.GetBornTime())
		robot.FisheIds.Store(v.GetId(), v.GetFishId())
	}
	//robot.Lock.Unlock()
}

func (robot *Robot) dead(buffer []byte) {
	res := &msg.DeadRes{}
	proto.Unmarshal(buffer, res)
	robot.checkRobotBehaviour(res.GetId())
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

func (robot *Robot) checkRobotBehaviour(fishId int32) {
	//robot.Lock.Lock()
	//delete(robot.Fishes, fishId)
	//delete(robot.FisheIds, fishId)
	//robot.Lock.Unlock()
	robot.Fishes.Delete(fishId)
	robot.FisheIds.Delete(fishId)
	if robot.TargetFishId == fishId {
		//robot.changeModel(2)
		robot.getChangeTarget(-1)
		robot.checkLockTarget()
	}
}

func (robot *Robot) changeRobotBehaviour(fishId int32) {
	if robot.TargetFishId == -1 && GetChance(int32(config.GetRobotLockChance()), 0) {
		//robot.changeModel(2)
		robot.getChangeTarget(fishId)
	}
}

func (robot *Robot) getLockFishId() int32 {
	id := int32(-1)
	//robot.Lock.Lock()
	//for k, fish := range robot.Fishes {
	//
	//	t := time.Now().UnixNano() / 1e6 - fish
	//	//t := fish.deadTime - time.Now().UnixNano()/1e6
	//	if t >= 8000 && t <= 18000 && config.GetRobotLockFishes(robot.FisheIds[k]) {
	//		//robot.Lock.Unlock()
	//		return k
	//	}
	//	id = k
	//}
	robot.Fishes.Range(func(key, value interface{}) bool {
		t := time.Now().UnixNano()/1e6 - value.(int64)
		id = key.(int32)
		//t := fish.deadTime - time.Now().UnixNano()/1e6
		fishId, _ := robot.FisheIds.Load(key)
		if fishId == nil {
			return false
		}
		if t >= 8000 && t <= 18000 && config.GetRobotLockFishes(fishId.(string)) {
			//robot.Lock.Unlock()
			return true
		}
		return false
	})
	//robot.Lock.Unlock()
	return id
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

func (robot *Robot) isLeave() {
	robot.AI.AddTimer(int64(1000), robot.isLeave)
}

func (robot *Robot) checkIsLeave() {
	t := 0
	if robot.BossTime {
		t = config.GetBossAddTime()
	}
	if robot.SkillNum[5] > 0 || robot.SkillNum[6] > 0 {
		t = config.GetSkillAddTime()
	}
	if t > 0 {
		robot.AI.AddTimer(int64(t), func() {
			robot.leaveGame()
		})
		return
	}
	robot.leaveGame()
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
	case int32(msg.MsgId_REFRESHFISH_Req):
		robot.refresh(buffer)
		break
	case int32(msg.MsgId_SKILLEND_Req):
		robot.skillEnd(buffer)
		break
	case int32(msg.MsgId_CHANGEMODEL_Res):
		robot.modelChange(buffer)
		break
	case int32(msg.MsgId_BOSSFORECAST_Req):
		robot.bossForecast(buffer)
		break
	case int32(msg.MsgId_FISHTIDEFORECAST_Req):
		robot.fishTideForecast(buffer)
		break
	case int32(msg.MsgId_ZERO):
		robot.init()
		robot.enterRoom()
		robot.start()
		break

	}
}

func (robot *Robot) start() {
	robot.AI.AddTimer(int64(1000), func() {
		robot.changeRotateModel(GetChance(50, 0))
		CheckRobotBulletLv(robot)
		CheckRobotShootModel(robot)
		RobotChangeShootPoint(robot)
		//t := robot.getRobotRestTime()
		robot.AI.AddTimerRepeat(int64(robot.ShootTime), 0, func() {
			robot.startShoot()
		})
	})
	robot.AI.AddTimerRepeat(int64(5000), 0, func() {
		CheckRobotBulletLv(robot)
	})

	robot.AI.AddTimerRepeat(int64(5000), 0, func() {
		CheckRobotShootModel(robot)
	})

	robot.AI.AddTimerRepeat(int64(config.GetRobotChangeTime()), 0, func() {
		RobotChangeShootPoint(robot)
	})

	robot.AI.AddTimer(int64(config.GetRobotQuitTime()), func() {
		robot.checkIsLeave()
	})

	robot.AI.AddTimerRepeat(int64(5000), 0, func() {
		robot.robotCheckCoinQuit()
	})

	robot.AI.AddTimerRepeat(int64(config.GetRobotModelCheckTime()), 0, func() {
		robot.checkModel()
	})

	robot.AI.AddTimerRepeat(int64(config.GetRobotLockCheckTime()), 0, func() {
		robot.checkLockTarget()
	})

	robot.AI.AddTimerRepeat(int64(config.GetRobotSkillFrozenCheckTime()), 0, func() {
		robot.checkSkillFrozen()
	})

	robot.AI.AddTimerRepeat(int64(config.GetRobotSkillSummonCheckTime()), 0, func() {
		robot.checkSkillSummon()
	})
}

func (robot *Robot) checkSkillSummon() {
	if robot.getChanceWan(config.GetRobotSkillSummonCheckChance(), 0) {
		req := &msg.SkillReq{}
		req.SkillId = 1
		robot.AI.SendMsgToServer(int32(msg.MsgId_SKILL_Req), req)
	}
}

func (robot *Robot) checkSkillFrozen() {
	if robot.getChanceWan(config.GetRobotSkillFrozenChance(), 0) {
		req := &msg.SkillReq{}
		fishes := make([]int32, 0)
		//robot.Lock.Lock()
		//for k:= range robot.Fishes {
		//	fishes = append(fishes, k)
		//}
		//robot.Lock.Unlock()
		robot.Fishes.Range(func(key, value interface{}) bool {
			fishes = append(fishes, key.(int32))
			return true
		})
		req.Fishes = fishes
		req.SkillId = 2
		robot.AI.SendMsgToServer(int32(msg.MsgId_SKILL_Req), req)
	}
}

func (robot *Robot) checkLockTarget() {
	if robot.IsLock {
		id := robot.getLockFishId()
		robot.getChangeTarget(id)
	}
}

func (robot *Robot) checkModel() {
	lock := false
	rand := tools.RandInt(0, 10000, 0)
	lockChance := config.GetRobotLockModelChance()
	if rand < lockChance {
		lock = true
	}

	if !lock && robot.IsLock {
		robot.changeModel(3)
	}
	if lock && !robot.IsLock {
		robot.changeModel(3)
	}

}

func (robot *Robot) getChanceWan(chance int, seed int64) bool {
	rand := tools.RandInt(0, 10000, seed)
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

func (robot *Robot) shoot(buffer []byte) {
	res := &msg.ShootRes{}
	proto.Unmarshal(buffer, res)
	if robot.AI.GetID() == res.GetUserId() {
		robot.BulletNum++
	}
}

func (robot *Robot) hit(buffer []byte) {
	res := &msg.HitRes{}
	proto.Unmarshal(buffer, res)
	if robot.AI.GetID() == res.GetUserId() {
		robot.BulletNum--
		if res.GetFish().GetScore() > 0 {
			robot.checkSkillNum(res.GetKey())
		}
	}
}

func (robot *Robot) checkSkillNum(fishId string) {
	skillId := config.GetSkillId(fishId)
	if skillId > 4 {
		robot.SkillNum[skillId]++
	}
	if skillId == 5 {
		robot.changeShootStaut()
	}
}

func (robot *Robot) changeShootStaut() {
	robot.CanShoot = false
	robot.AI.AddTimer(int64(2500), func() {
		robot.CanShoot = true
	})
}

func (robot *Robot) upgrade(buffer []byte) {

}

func (robot *Robot) fishTideForecast(buffer []byte) {
	robot.BossTime = false
}

func (robot *Robot) bossForecast(buffer []byte) {
	robot.BossTime = true
}

func (robot *Robot) skillEnd(buffer []byte) {
	res := &msg.SkillEndReq{}
	proto.Unmarshal(buffer, res)
	skillId := res.GetSkillId()
	if res.GetUserId() == robot.AI.GetID() && robot.SkillNum[skillId] > 0 {
		robot.SkillNum[skillId]--
	}
}

func (robot *Robot) modelChange(buffer []byte) {
	res := &msg.ChangemModelReq{}
	proto.Unmarshal(buffer, res)
	if robot.AI.GetID() == res.GetUserId() && res.GetModelId() == 3 {
		robot.IsLock = !robot.IsLock
		if robot.TargetFishId == -1 {
			robot.checkLockTarget()
		}
	}
}
