package game

import (
	"common/score"
	"fmt"
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/global"
	"game_buyu/rob_red/msg"
	msg2 "game_frame_v2/msg"
	"strings"
	"time"
)

//抢红包
func (game *Game) robRed(user *data.User, red *Red) {
	if user.User.GetScore() < red.OriginAmount {
		fmt.Println("用户金额少于红包金额")
		user.User.SendMsg(global.ERROR_ROB_NOT_ENOUGH, &msg.S2CAxis{})
		return
	}
	if red.RobbedCount >= red.RedFlood {
		fmt.Println("红包已被抢完")
		user.User.SendMsg(global.ERROR_CODE_RED_OVER, &msg.S2CAxis{})
		game.DelRedListMap(red)
		return
	}
	roomProb, _ := game.Table.GetRoomProb()
	var robbedAmount int64
	var mineAmount int64
	isMine := isUserMine(user, red, roomProb)
	robbedAmount = red.GetRobAmount(user, isMine)
	if robbedAmount%10 == red.MineNum {
		isMine = true
	} else {
		isMine = false
	}
	red.AddRobbedCount()
	red.SubAmount(robbedAmount)
	if robbedAmount == 0 {
		if !user.User.IsRobot() {
			//user.GameLogs = append(user.GameLogs,&msg2.GameLog{
			//	UserId:user.User.GetId(),
			//	Content:"抢红包金额："+score.GetScoreStr(robbedAmount)+" 余额："+score.GetScoreStr(user.User.GetScore())+
			//		"用户作弊率："+fmt.Sprintf(`%d`,red.sender.Cheat)+"系统作弊率："+fmt.Sprintf(`%d`,roomProb)+
			//		"红包剩余次数："+fmt.Sprintf(`%d`,red.RedFlood-red.RobbedCount)+"红包剩余金额："+score.GetScoreStr(red.Amount)+
			//		" 发包人id："+fmt.Sprintf(`%d`,red.sender.User.GetId())+"抢包人id："+fmt.Sprintf(`%d`,user.User.GetId())+"雷号："+fmt.Sprintf(`%d`,red.MineNum),
			//})

			//game.Table.WriteLogs(user.User.GetId(), "抢红包金额："+score.GetScoreStr(robbedAmount)+" 余额："+score.GetScoreStr(user.User.GetScore())+
			//	"用户作弊率："+fmt.Sprintf(`%d`,red.sender.Cheat)+"系统作弊率："+fmt.Sprintf(`%d`,roomProb)+
			//	"红包剩余次数："+fmt.Sprintf(`%d`,red.RedFlood-red.RobbedCount)+"红包剩余金额："+score.GetScoreStr(red.Amount)+
			//	" 发包人id："+fmt.Sprintf(`%d`,red.sender.User.GetId())+"抢包人id："+fmt.Sprintf(`%d`,user.User.GetId())+"雷号："+fmt.Sprintf(`%d`,red.MineNum))
		}
		return
	}
	game.SetRobRedUserArr(red.Id, user.User.GetId())
	user.AddAmount(robbedAmount)
	if user.User.GetId() == red.sender.User.GetId() {
		red.SelfRobbedAmount += robbedAmount
	}
	//fmt.Println("11111系统作弊率：",roomProb,"抢包玩家作弊率：",user.Cheat,"是否为机器人：",user.User.IsRobot(),"  发包玩家作弊率：",red.sender.Cheat,"是否为机器人：",red.sender.User.IsRobot())
	if isMine {
		mineAmount = red.OriginAmount * (int64(config.RedConfig.Odds) / 100)
		user.BetsAmount += mineAmount
		taxScore := mineAmount * game.Table.GetRoomRate() / 10000
		recordScore := mineAmount - taxScore
		red.sender.ProfitAmount += recordScore
		user.ProfitAmount -= mineAmount
		red.sender.Output += recordScore

		_, _ = user.User.SetScore(game.Table.GetGameNum(), -mineAmount+robbedAmount, game.Table.GetRoomRate())
		_ = user.User.SendMsg(global.S2C_USER_SCORE, &msg.S2CUserScore{Score: user.User.GetScore()})
		profitAmount, _ := red.sender.User.SetScore(game.Table.GetGameNum(), mineAmount, game.Table.GetRoomRate())
		red.sender.DrawAmount += taxScore
		game.mineFrameLogSum(red, profitAmount, mineAmount-profitAmount)
		red.sender.AddMineAmountToRecord(red.Id, mineAmount)
		user.AddMineRecord(red.NewMineRecord2C(game.GetLevelStr(), robbedAmount, mineAmount))
		_ = red.sender.User.SendMsg(global.S2C_USER_SCORE, &msg.S2CUserScore{Score: red.sender.User.GetScore()})
		//输家打码量
		if user.User.GetId() != red.sender.User.GetId() {
			chip := mineAmount
			user.Chip += chip
			if !user.User.IsRobot() {
				fmt.Println("中雷打码量：", mineAmount, "当前打码量：", user.Chip)
			}
		}
		game.TriggerHorseLamp(red.sender, mineAmount)
		if !user.User.IsRobot() {
			user.GameLogs = append(user.GameLogs, &msg2.GameLog{
				UserId: user.User.GetId(),
				Content: "抢红包中雷，赔付金额：" + score.GetScoreStr(mineAmount-robbedAmount) + " 余额：" + score.GetScoreStr(user.User.GetScore()) +
					"抢包玩家中雷，作弊率：" + fmt.Sprintf(`%d`, user.Cheat) + " 系统作弊率：" + fmt.Sprintf(`%d`, roomProb) +
					"红包剩余次数：" + fmt.Sprintf(`%d`, red.RedFlood-red.RobbedCount) + "红包剩余金额：" + score.GetScoreStr(red.Amount) +
					" 发包人id：" + fmt.Sprintf(`%d`, red.sender.User.GetID()) + "抢包人id：" + fmt.Sprintf(`%d`, user.User.GetID()) + "雷号：" + fmt.Sprintf(`%d`, red.MineNum),
			})
			//game.Table.WriteLogs(user.User.GetId(), "抢红包中雷，赔付金额："+score.GetScoreStr(mineAmount-robbedAmount)+" 余额："+score.GetScoreStr(user.User.GetScore())+
			//	"抢包玩家中雷，作弊率："+fmt.Sprintf(`%d`,user.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`,roomProb)+
			//	"红包剩余次数："+fmt.Sprintf(`%d`,red.RedFlood-red.RobbedCount)+"红包剩余金额："+score.GetScoreStr(red.Amount)+
			//	" 发包人id："+fmt.Sprintf(`%d`,red.sender.User.GetId())+"抢包人id："+fmt.Sprintf(`%d`,user.User.GetId())+"雷号："+fmt.Sprintf(`%d`,red.MineNum))

			//game.Table.WriteLogs(red.sender.User.GetId(), "收到玩家中雷赔付金额："+score.GetScoreStr(mineAmount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore())+
			//	"发包玩家作弊率："+fmt.Sprintf(`%d`,red.sender.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`,roomProb)+
			//	"发包人id："+fmt.Sprintf(`%d`,red.sender.User.GetId())+"抢包人id："+fmt.Sprintf(`%d`,user.User.GetId())+"雷号："+fmt.Sprintf(`%d`,red.MineNum))
		}
		if !red.sender.User.IsRobot() {
			//fmt.Println("发包人id：",red.sender.User.GetId()," 内存地址：",red.sender,&red.sender,"日志长度：",len(red.sender.GameLogs))
			red.sender.GameLogs = append(red.sender.GameLogs, &msg2.GameLog{
				UserId: red.sender.User.GetId(),
				Content: "收到玩家中雷赔付金额：" + score.GetScoreStr(mineAmount) + " 余额：" + score.GetScoreStr(red.sender.User.GetScore()) +
					"发包玩家作弊率：" + fmt.Sprintf(`%d`, red.sender.Cheat) + " 系统作弊率：" + fmt.Sprintf(`%d`, roomProb) +
					"发包人id：" + fmt.Sprintf(`%d`, red.sender.User.GetID()) + "抢包人id：" + fmt.Sprintf(`%d`, user.User.GetID()) + "雷号：" + fmt.Sprintf(`%d`, red.MineNum),
			})
		}
	} else {
		//fmt.Println("房间税率：",game.Table.GetRoomRate(),"用户金额前：",user.User.GetScore()," 抢包金额：",robbedAmount)
		_, _ = user.User.SetScore(game.Table.GetGameNum(), robbedAmount, 0)
		//fmt.Println("房间税率：",game.Table.GetRoomRate(),"用户金额后：",user.User.GetScore())
		//fmt.Println("user.Cheat : ",user.Cheat)
		//赢家打码量 服务费
		if !user.User.IsRobot() {
			user.GameLogs = append(user.GameLogs, &msg2.GameLog{
				UserId: user.User.GetId(),
				Content: "抢红包金额：" + score.GetScoreStr(robbedAmount) + " 余额：" + score.GetScoreStr(user.User.GetScore()) +
					"抢包玩家作弊率：" + fmt.Sprintf(`%d`, user.Cheat) + " 系统作弊率：" + fmt.Sprintf(`%d`, roomProb) +
					"红包剩余次数：" + fmt.Sprintf(`%d`, red.RedFlood-red.RobbedCount) + "红包剩余金额：" + score.GetScoreStr(red.Amount) +
					" 发包人id：" + fmt.Sprintf(`%d`, red.sender.User.GetID()) + "抢包人id：" + fmt.Sprintf(`%d`, user.User.GetID()) + "雷号：" + fmt.Sprintf(`%d`, red.MineNum) +
					"发包人是:" + aiRealStr(red.sender.User.IsRobot()) + " 抢包人是：" + aiRealStr(user.User.IsRobot()) + " 玩家状态：" + userProbStr(user),
			})
			//game.Table.WriteLogs(user.User.GetId(), "抢红包金额："+score.GetScoreStr(robbedAmount)+" 余额："+score.GetScoreStr(user.User.GetScore())+
			//	"抢包玩家作弊率："+fmt.Sprintf(`%d`,user.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`,roomProb)+
			//	"红包剩余次数："+fmt.Sprintf(`%d`,red.RedFlood-red.RobbedCount)+"红包剩余金额："+score.GetScoreStr(red.Amount)+
			//	" 发包人id："+fmt.Sprintf(`%d`,red.sender.User.GetId())+"抢包人id："+fmt.Sprintf(`%d`,user.User.GetId())+"雷号："+fmt.Sprintf(`%d`,red.MineNum)+
			//	"发包人是否为机器人:"+boolStr(red.sender.User.IsRobot())+"抢包人是否机器人："+boolStr(user.User.IsRobot()))
		}
	}
	user.Output += robbedAmount
	user.ProfitAmount += robbedAmount
	//fmt.Println("22222系统作弊率：",roomProb,"抢包玩家作弊率：",user.Cheat,"是否为机器人：",user.User.IsRobot(),"  发包玩家作弊率：",red.sender.Cheat,"是否为机器人：",red.sender.User.IsRobot())
	_ = user.User.SendMsg(global.S2C_ROB_RED, red.NewRobbedRed(isMine, robbedAmount, user.User.GetScore()))
	if red.RobbedCount >= red.RedFlood {
		//red.sender.BetsAmount += red.OriginAmount
		//红包被抢完添加打码量
		red.sender.Chip = red.sender.Chip + red.OriginAmount - red.SelfRobbedAmount
		if !red.sender.User.IsRobot() {
			//fmt.Println("红包被抢完，打码量：",red.OriginAmount," 红包id：",red.Id," 发包者id：",red.sender.User.GetId(),"当前打码量：",red.sender.Chip)
		}
		game.DelRedListMap(red)
		game.DelLockListByRedId(red.Id)
		go red.sender.ChangeUserSendRedStatus(red.Id, global.RED_CUR_STATUS_OVER)
		game.DelRobRedUserArr(red.Id)
		//fmt.Println("------------3-----------",time.Now().Sub(t1))
		game.SendRedFromWaitList()
		//fmt.Println("------------4-----------",time.Now().Sub(t1))
	}
	user.AddRobRedRecord(red.NewUserRobbedRedInfo(robbedAmount, game.GetLevelStr(), isMine))
	//广播策略改为 人数50个以上如果是自己的就一定发，其他人则概率性发送
	game.Table.Broadcast(global.S2C_GET_HALL_RECORD, &msg.S2CHallRecord{
		UserId: user.Id, UserName: user.User.GetNike(), RobbedAmount: robbedAmount, IsMine: isMine, MineAmount: mineAmount, Time: time.Now().Unix(),
		SenderId: red.sender.Id, SenderName: red.sender.Name, RedAmount: red.OriginAmount, MineNum: red.MineNum,
	})
	//fmt.Println("user id : ",user.Id," 抢到的金额：", robbedAmount, " 中雷赔付的金额： ", mineAmount,"红包id：",red.Id, " 红包剩余金额： ", red.Amount,"雷号：",red.MineNum,"剩余血量：",red.RedFlood-red.RobbedCount)
	if red.RobbedCount == 3 || red.RobbedCount == 6 {
		game.Table.Broadcast(global.S2C_RED_ROBBED_COUNT, red.NewRobbedCount2C())
	}
	game.TriggerHorseLamp(user, robbedAmount)
}

//true : 返回 "机器人"
func aiRealStr(flag bool) string {
	if flag {
		return "【机器人】"
	}
	return "【真实玩家】"
}

func userProbStr(user *data.User) string {
	if user.User.GetProb() == 0 {
		return "使用血池"
	}
	return "被点控"
}

//玩家是否中雷
func isUserMine(user *data.User, red *Red, prob int32) bool {
	if user.User.GetId() == red.sender.User.GetId() {
		//自己抢自己的红包一定不中雷
		return false
	}
	//return false
	if user.User.GetProb() == 0 {
		user.Cheat = prob
	} else {
		user.Cheat = user.User.GetProb()
	}
	if user.User.IsRobot() {
		user.Cheat = prob
	}
	if user.Cheat == 0 {
		user.Cheat = 1000
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
	if red.sender.Cheat == 0 {
		red.sender.Cheat = 1000
	}

	userMine := 1
	for _, v := range config.AiConfig.RobotConfig {
		if v.Cheat == user.Cheat {
			if user.User.IsRobot() {
				userMine = v.AiMine
			} else {
				userMine = v.UserMine
			}

			break
		}
	}

	index := user.RandInt(1, global.WAN_RATE_TOTAL)
	if !user.User.IsRobot() {
		fmt.Println("房间作弊率：", prob, "用户作弊率：", user.Cheat, "中雷概率：", userMine, "随机数：", index)
	}
	if index < userMine {
		return true
	}
	return false
	return user.RateToExecWithIn(userMine, global.WAN_RATE_TOTAL)
}

//获取下一个红包要生成的id
func (game *Game) getNextRedId() (id int64) {
	return game.totalRedCount + 1
}

func (game *Game) GetRouteAxis() int32 {
	var i, route int32 = 1, 1
	for i = 1; i <= global.RED_TOTAL_ROUTE; i++ {
		if !game.hasRedRoute(i) {
			route = i
			break
		}
	}
	//fmt.Println("红包路线   >>   :", route)
	return route
}

func (game *Game) hasRedRoute(route int32) bool {
	for e := game.curShowRedList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if red.Route == route {
			return true
		}
	}
	return false
}

//将某个用户的所有红包全部退还
//func (game *Game) BackUserAllRed(uid int64) {
//	for e := game.redList.Front(); e != nil; e = e.Next() {
//		red := e.Value.(*Red)
//		if red.sender.User.GetId() == uid {
//			game.redList.Remove(e)
//			game.DelCurShowRed(red.Id)
//			game.DelWaitRed(red.Id)
//			for e := game.userList.Front(); e != nil; e = e.Next() {
//				user := e.Value.(*data.User)
//				if !user.IsAi {
//					//fmt.Println("红包消失id：", red.Id)
//					if err := user.User.SendMsg(global.S2C_RED_DISAPPEAR, &msg.S2CRedDisappear{
//						RedId: red.Id, Level: red.level, IsRobbed: user.IsRobbedRed(red.Id), IsOtherRobbed: game.IsOtherRobbed(red.Id, user.User.GetId()),
//					}); err != nil {
//					}
//				}
//			}
//			_, _ = red.sender.User.SetScore(game.Table.GetGameNum(), red.Amount, 0)
//			if !red.sender.User.IsRobot() {
//				game.Table.WriteLogs(red.sender.User.GetId(), "红包没抢完，退还："+score.GetScoreStr(red.Amount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore())+
//					"红包剩余次数："+fmt.Sprintf(`%d`,red.RedFlood-red.RobbedCount)+"红包剩余金额："+score.GetScoreStr(red.Amount))
//			}
//		}
//	}
//}

//是否触发跑马灯,有特殊条件就是and，没有特殊条件满足触发金额即可
func (game *Game) TriggerHorseLamp(winner *data.User, winAmount int64) {
	for _, v := range game.HoseLampArr {
		if strings.TrimSpace(v.SpecialCondition) == "" {
			if winAmount >= v.AmountLimit && fmt.Sprintf(`%d`, game.Table.GetRoomID()) == v.RoomId {
				if !winner.User.IsRobot() {
					//fmt.Println("创建没有特殊条件的跑马灯")
				}
				if err := game.Table.CreateMarquee(winner.User.GetNike(), winAmount, "", v.RuleId); err != nil {
				}
			} else {
				if !winner.User.IsRobot() {
					//fmt.Println("不创建跑马灯：抢到的金额和配置金额：",winAmount,v.AmountLimit,game.Table.GetRoomID(),v.RoomId)
				}
			}
		} else {
			if !winner.User.IsRobot() {
				//fmt.Println("带有特殊条件 : ",strings.TrimSpace(v.SpecialCondition),game.Table.GetRoomID(),v.RoomId)
			}
		}
	}
}

// 100 90 盈利金额就是90，drawAmount就是10
//1月8号新加的功能，如果用户离线，抢红包中雷就调一次框架提供的这个方法，以供日志汇总
func (game *Game) mineFrameLogSum(red *Red, profitAmount, drawAmount int64) {
	if game.GetUserList(red.sender.Id) == nil {
		//game.Table.AddGameEnd(red.sender.User, profitAmount, 0,
		//	drawAmount, "红包中雷赔付，玩家离开")
	}
}
