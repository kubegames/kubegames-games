package game

import (
	"common/page"
	"common/rand"
	"fmt"
	"game_buyu/crazy_red/config"
	"game_buyu/crazy_red/data"
	"game_buyu/crazy_red/global"
	"game_buyu/crazy_red/msg"
	"strings"
)

//抢红包
func (game *Game) robRed(user *data.User, red *Red) {
	red.Lock.Lock()
	if game.isUserRobbed(user.User.GetId()) {
		//fmt.Println("玩家已经抢过该红包")
		user.User.SendMsg(global.ERROR_CODE_USER_ROBBED, user.GetUserMsgInfo())
		red.Lock.Unlock()
		return
	}
	var robbedAmount int64
	var mineAmount int64
	roomProb,_ := game.Table.GetRoomProb()
	//fmt.Println("roomProb ::: ",roomProb)
	isMine := isUserMine(user,roomProb,red)
	robbedAmount = red.GetRobAmount(user,isMine)
	if robbedAmount%10 == red.MineNum {
		isMine = true
	} else {
		isMine = false
	}

	if user.User.GetId() == red.sender.User.GetId() {
		red.SelfRobbedAmount += robbedAmount
	}

	red.AddRobbedCount()
	red.SubAmount(robbedAmount)
	user.NotOperateCount = 0
	user.AddAmount(robbedAmount)

	game.curRobbedUserArr = append(game.curRobbedUserArr, user.NewCurRobUser(isMine, mineAmount, robbedAmount))
	if game.MaxRobUidMap[red.Id] == nil {
		game.MaxRobUidMap[red.Id] = &UserRobStruct{User: user, RobbedAmount: robbedAmount}
	}
	if robbedAmount > game.MaxRobUidMap[red.Id].RobbedAmount && !isMine {
		game.MaxRobUidMap[red.Id] = &UserRobStruct{User: user, RobbedAmount: robbedAmount}
	}
	game.Table.Broadcast(global.S2C_ROB_RED, red.NewRobbedRed(isMine, robbedAmount, user))
	game.UserRobbedArr = append(game.UserRobbedArr, &UserRobStruct{
		Red:        red,
		User:          user,
		RobbedAmount: robbedAmount,
		IsMine:       isMine,
		MineAmount:   mineAmount,
		RedRemainAmount:red.Amount,
		RedRobbedCount:red.RobbedCount,
		MineNum:red.MineNum,
	})
	if red.RobbedCount >= red.RedFlood {
		game.DelRedListMap(red)
		game.restart()
	}
	red.Lock.Unlock()

}

func (game *Game) isUserRobbed(uid int64) bool {
	for _, v := range game.UserRobbedArr {
		if v.User.User.GetId() == uid {
			return true
		}
	}
	return false
}

//玩家是否中雷
func isUserMine(user *data.User,prob int32,red *Red) bool {

	if user.User.IsRobot() && red.sender.User.IsRobot() {
		//fmt.Println("机器人抢机器人就5%的概率中雷")
		return rand.RateToExec(5)
	}
	if user.User.GetProb() == 0 {
		user.Cheat = prob
	} else {
		user.Cheat = user.User.GetProb()
	}
	if user.User.IsRobot() {
		user.Cheat = prob
	}
	//发包者
	if red.sender.User.GetProb() == 0 {
		red.sender.Cheat = prob
	} else {
		red.sender.Cheat = red.sender.User.GetProb()
	}
	if red.sender.User.IsRobot() {
		red.sender.Cheat = prob
	}
	if user.Cheat == 0 {
		fmt.Println("玩家作弊率为0，房间作弊率：",prob)
		user.Cheat = 1000
	}

	userMine := config.AiRobConfigMap[user.Cheat].UserMine
	if user.User.IsRobot() {
		userMine = config.AiRobConfigMap[user.Cheat].AiMine
	}

	//fmt.Println("作弊率下中雷概率：",userMine)
	randInt := user.RandInt(1,global.WAN_RATE_TOTAL)
	//if user.User.IsRobot() {
		fmt.Println("玩家id:",user.User.GetId(),"作弊率：",user.Cheat," 产生的随机值：",randInt,"中雷的配置参数：",userMine,"系统作弊率：",prob)
	//}
	if randInt <= userMine {
		return true
	}
	return false


	//return user.RateToExecWithIn(int(userMine), global.WAN_RATE_TOTAL)
}

//获取下一个红包要生成的id
func (game *Game) getNextRedId() (id int64) {
	return game.totalRedCount + 1
}

func (game *Game) GetCurRedList(user *data.User, pageIndex, pageSize int) *msg.S2CCurRedArr {
	totalCurRedList := make([]*msg.S2CCurRed, 0)
	for e := game.redList.Front(); e != nil; e = e.Next() {
		totalCurRedList = append(totalCurRedList, e.Value.(*Red).GetCurRed2C())
	}
	res := new(msg.S2CCurRedArr)
	res.RedArr = make([]*msg.S2CCurRed, 0)
	for i := pageIndex * pageSize; i < pageIndex*pageSize+pageSize; i++ {
		if len(totalCurRedList)-1 < i {
			continue
		} else {
			res.RedArr = append(res.RedArr, totalCurRedList[i])
		}
	}
	pager := page.NewPager(pageIndex, pageSize, len(totalCurRedList))
	res.Size = int32(pager.Size)
	res.Current = int32(pager.Current)
	res.Total = int64(pager.Total)
	res.Pages = int32(pager.Pages)
	res.SelfRedPos = game.GetUserRedPos(user.User.GetId())
	return res
}

func (game *Game) GetLevelStr() string {
	switch game.Table.GetLevel() {
	case 1:
		return "初级场"
	case 2:
		return "中级场"
	case 3:
		return "高级场"
	case 4:
		return "大师场"
	}
	return "初级场"
}

//获取用户的红包当前排序位置
func (game *Game) GetUserRedPos(uid int64) (pos int32) {
	hasSelf := false
	for e := game.redList.Front(); e != nil; e = e.Next() {
		pos++
		red := e.Value.(*Red)
		if red.sender.User.GetId() == uid {
			hasSelf = true
			return
		}
	}
	if !hasSelf {
		return 0
	}
	return
}

//红包抢的最大用户
type UserRobStruct struct {
	Red          *Red
	User         *data.User //抢包用户
	RobbedAmount int64
	IsMine       bool
	MineAmount   int64
	RedRemainAmount int64 //红包剩余金额
	RedRobbedCount int64 //红包已抢个数
	MineNum int64 // 雷号
}

//func SetMaxRobUidMap(redId int64,urs *UserRobStruct)  {
//	maxRobUidLock.Lock()
//	MaxRobUidMap[redId] = urs
//	maxRobUidLock.Unlock()
//}
//
func (game *Game) DelMaxRobUidMap(redId int64) {
	game.maxRobUidLock.Lock()
	delete(game.MaxRobUidMap, redId)
	game.maxRobUidLock.Unlock()
}


//是否触发跑马灯,有特殊条件就是and，没有特殊条件满足触发金额即可
func (game *Game) TriggerHorseLamp(winner *data.User, winAmount int64) {
	for _, v := range game.HoseLampArr {
		if strings.TrimSpace(v.SpecialCondition) == "" {
			if winAmount >= v.AmountLimit && fmt.Sprintf(`%d`,game.Table.GetRoomID()) == v.RoomId {
				//fmt.Println("创建没有特殊条件的跑马灯")
				if err := game.Table.CreateMarquee(winner.User.GetNike(), winAmount, "", v.RuleId); err != nil {
				}
			}else {
				if !winner.User.IsRobot(){
					//fmt.Println("不创建跑马灯：抢到的金额和配置金额：",winAmount,v.AmountLimit)
				}
			}
		}else {
			if !winner.User.IsRobot(){
				//fmt.Println("带有特殊条件 : ",strings.TrimSpace(v.SpecialCondition))
			}
		}
	}
}

// 100 90 盈利金额就是90，drawAmount就是10
//1月8号新加的功能，如果用户离线，抢红包中雷就调一次框架提供的这个方法，以供日志汇总
func (game *Game)mineFrameLogSum(red *Red,profitAmount,drawAmount int64)  {
	//if game.GetUserListMap(red.sender.Id) == nil {
	//	game.Table.AddGameEnd(red.sender.User, profitAmount, 0,
	//		drawAmount, "红包中雷赔付，玩家离开")
	//}
}