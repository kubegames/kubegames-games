package data

import (
	"common/page"
	"fmt"
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/global"
	"game_buyu/rob_red/msg"
	msg2 "game_frame_v2/msg"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	Id           int64 // 框架那边proto定义的int32
	Name         string
	ChairId      uint  //椅子id
	Status       int32 //玩家当前状态，0：正在游戏，1：比牌输了，2：弃牌
	IsAi         bool  //是否为机器人
	RobbedAmount int64 //抢到的金额
	Table        table.TableInterface
	User         player.PlayerInterface
	*config.AiRobConfig
	SendRedList       []*msg.S2CSendRedRecord //发过的红包列表
	totalSentAmount   int64                   //发送过的总金额
	RobbedList        []*msg.S2CRobbedRedInfo //抢过的红包记录
	totalRobbedAmount int64                   //抢过的总金额
	MineList          []*msg.S2CMinRecord
	totalMineAmount   int64 //中雷过的总金额
	LastRobTime       time.Time

	IsSendWaitMsg bool
	Cheat         int32
	Chip          int64 //打码量
	BetsAmount    int64 //下注金额（投入） 投入只有发包和中雷两个地方
	Output        int64 //产出 ： 抢+退回+收到中雷赔付（扣税)
	DrawAmount    int64
	//BetsAmount int64 //战绩 抢包金额+用户赔付
	ProfitAmount int64
	GameLogs     []*msg2.GameLog
	GameNum      string
}

func NewUser(uid int64, name string, isAi bool, table table.TableInterface) *User {
	user := &User{
		Id: uid, Name: name, IsAi: isAi, Table: table, SendRedList: make([]*msg.S2CSendRedRecord, 0),
		RobbedList: make([]*msg.S2CRobbedRedInfo, 0), MineList: make([]*msg.S2CMinRecord, 0),
		LastRobTime: time.Now(), GameLogs: make([]*msg2.GameLog, 0),
	}
	return user
}

//用户资金增加
func (user *User) AddAmount(amount int64) {
	user.RobbedAmount += amount
	if !user.User.IsRobot() {
		user.LastRobTime = time.Now()
	}
}

func (user *User) GetUserMsgInfo() *msg.S2CUserInfo {
	return &msg.S2CUserInfo{
		Name: user.Name, Uid: user.Id, Head: user.User.GetHead(), Amount: user.User.GetScore(), ChairId: int32(user.ChairId), Status: user.Status,
		RobbedAmount: user.RobbedAmount,
	}
}

func (user *User) RandInt(min, max int) int {
	//if min >= max || min == 0 || max == 0 {
	//	return max.
	//}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min) + min
}

//传入指定概率，然后返回是否执行  比如 rate：90 表示有90%的概率要执行
func (user *User) RateToExec(rate int) bool {
	r := user.RandInt(1, 100)
	//fmt.Println("随机数r : ",r)
	if r < rate {
		return true
	}
	return false
}

//从max中随机去一个数，看是否小于rate
func (user *User) RateToExecWithIn(rate, max int) bool {
	r := user.RandInt(1, max)
	//fmt.Println("随机数r : ", r)
	if r < rate {
		return true
	}
	return false
}

func (user *User) ResetUser() {
	user.Status = global.USER_CUR_STATUS_FINISH_GAME
}

//添加发包记录
func (user *User) AddSendRedRecord(record *msg.S2CSendRedRecord) {
	if user.IsAi {
		return
	}
	user.totalSentAmount += record.RedAmount
	user.SendRedList = append(user.SendRedList, record)
}

//改变发送过的红包状态
func (user *User) ChangeUserSendRedStatus(redId int64, status int32) {
	for _, v := range user.SendRedList {
		if v.RedId == redId {
			//fmt.Println("更改红包status")
			v.Status = status
			break
		}
	}
}

func (user *User) AddRobRedRecord(record *msg.S2CRobbedRedInfo) {
	if user.IsAi {
		return
	}
	user.totalRobbedAmount += record.RobbedAmount
	user.RobbedList = append(user.RobbedList, record)
}

func (user *User) AddMineRecord(record *msg.S2CMinRecord) {
	if user.IsAi {
		return
	}
	user.totalMineAmount += record.MineAmount
	user.MineList = append(user.MineList, record)
}

func (user *User) GetSendRedRecord(pageIndex, pageSize int) *msg.S2CSendRedRecordArr {
	//fmt.Println("user : ", user.Id, " 发包 index :", pageIndex, "size : ", pageSize, " total : ", len(user.SendRedList))
	res := new(msg.S2CSendRedRecordArr)
	res.RedArr = make([]*msg.S2CSendRedRecord, 0)
	pager := page.NewPager(pageIndex, pageSize, len(user.SendRedList))
	for i := pageIndex * pageSize; i < pageIndex*pageSize+pageSize; i++ {
		if len(user.SendRedList)-1 < i {
			continue
		} else {
			res.RedArr = append(res.RedArr, user.SendRedList[i])
		}
	}
	res.Size = int32(pager.Size)
	res.Current = int32(pager.Current)
	res.Total = int64(pager.Total)
	res.Pages = int32(pager.Pages)
	for _, v := range user.SendRedList {
		//fmt.Println("红包id：",v.RedId," origin amount: ",v.OriginAmount)
		res.TotalAmount += v.RedAmount
	}
	fmt.Println("res total : ", res.TotalAmount)
	res.TotalCount = int64(len(user.SendRedList))
	//fmt.Println("res ::: ", fmt.Sprintf(`%+v`, res))
	return res
}

//获取
func (user *User) IsRobbedRed(redId int64) bool {
	for _, v := range user.RobbedList {
		if v.RedId == redId {
			return true
		}
	}
	return false
}

func (user *User) GetRobRedRecord(pageIndex, pageSize int) *msg.S2CRobbedRedInfoArr {
	//fmt.Println("user : ", user.Id, "抢包：index :", pageIndex, "size : ", pageSize, " total : ", len(user.RobbedList))
	res := new(msg.S2CRobbedRedInfoArr)
	res.RobbedArr = make([]*msg.S2CRobbedRedInfo, 0)
	for i := pageIndex * pageSize; i < pageIndex*pageSize+pageSize; i++ {
		if len(user.RobbedList)-1 < i {
			continue
		} else {
			res.RobbedArr = append(res.RobbedArr, user.RobbedList[i])
		}
	}
	pager := page.NewPager(pageIndex, pageSize, len(user.RobbedList))
	res.Size = int32(pager.Size)
	res.Current = int32(pager.Current)
	res.Total = int64(pager.Total)
	res.Pages = int32(pager.Pages)
	res.TotalAmount = user.totalRobbedAmount
	res.TotalCount = int64(len(user.RobbedList))
	//fmt.Println("res ::: ", fmt.Sprintf(`%+v`, res))
	return res
}

func (user *User) GetMineRecord(pageIndex, pageSize int) *msg.S2CMinRecordArr {
	//fmt.Println("user : ", user.Id, "抢包：index :", pageIndex, "size : ", pageSize, " total : ", len(user.MineList))
	res := new(msg.S2CMinRecordArr)
	res.MineArr = make([]*msg.S2CMinRecord, 0)
	for i := pageIndex * pageSize; i < pageIndex*pageSize+pageSize; i++ {
		if len(user.MineList)-1 < i {
			continue
		} else {
			res.MineArr = append(res.MineArr, user.MineList[i])
		}
	}
	pager := page.NewPager(pageIndex, pageSize, len(user.MineList))
	res.Size = int32(pager.Size)
	res.Current = int32(pager.Current)
	res.Total = int64(pager.Total)
	res.Pages = int32(pager.Pages)
	res.TotalAmount = user.totalMineAmount
	res.TotalCount = int64(len(user.MineList))
	//fmt.Println("res ::: ", fmt.Sprintf(`%+v`, res))
	return res
}

//给发送的红包添加中雷获得金额
func (user *User) AddMineAmountToRecord(redId, mineAmount int64) {
	for _, v := range user.SendRedList {
		if v.RedId == redId {
			v.TotalMineAmount += mineAmount
		}
	}
}
