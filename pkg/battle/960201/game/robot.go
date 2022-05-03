package game

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

type Robot struct {
	game   *Game
	AiUser player.RobotInterface
	user   *data.User
}

func NewRobot(game *Game, user *data.User) *Robot {
	return &Robot{
		game: game, user: user,
	}
}

// OnGameMessage 机器人收到消息
func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	////log.Traceln("机器人接收消息：",subCmd)
	switch subCmd {
	case global.S2C_CUR_ACTION_USER:
		s2cMsg := &msg.S2CUserInfo{}
		if err := proto.Unmarshal(buffer, s2cMsg); err != nil {
			//log.Traceln("unmarshal err : ", err)
			return
		}
		//log.Traceln("该机器人：",s2cMsg.UserId," 发言"," 自己的id：",robot.user.Id)
		if s2cMsg.UserId != robot.user.Id {
			return
		}
		//thinkTime := rand2.RandInt(1, 4)
		time.AfterFunc(2*time.Second, func() {
			robot.AiActionNew()
		})
	case global.ERROR_CODE_OVER_USER_MIN_AMOUNT:
		robot.straightFollow(buffer)
	case global.ERROR_CODE_COMPARE_NOT_ENOUGH:
		robot.straightFollow(buffer)
	}
}

//直接跟注
func (robot *Robot) straightFollow(buffer []byte) {
	log.Traceln("收到 直接跟注")
	s2cMsg := &msg.S2CUserInfo{}
	if err := proto.Unmarshal(buffer, s2cMsg); err != nil {
		log.Traceln("unmarshal err : ", err)
		return
	}
	////log.Traceln("该机器人：",s2cMsg.UserId," 发言"," 自己的id：",robot.user.Id)
	if s2cMsg.UserId != robot.user.Id {
		return
	}
	if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, 0, global.USER_OPTION_FOLLOW)); err != nil {
	}
}

//机器人发言
func (robot *Robot) AiActionNew() {
	if robot.game.IsAllIn {
		robot.aiAllIn()
		return
	}

	if robot.user.IsSawCards {
		//log.Traceln("aiSawAction：：：",robot.AiUser.GetID())
		robot.aiSawAction()
	} else {
		//log.Traceln("aiSawAction else ：：：",robot.AiUser.GetID())
		robot.aiNotSee()
	}
}

func NewC2sUserAction(userId, amount int64, option string) *msg.C2SUserAction {
	return &msg.C2SUserAction{
		UserId: userId, Option: option, Amount: amount,
	}
}

//ai没看牌要做的操作
func (robot *Robot) aiNotSee() {

	//如果看牌则执行看牌操作
	if robot.aiIsSeeCard() {
		if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, 0, global.USER_OPTION_SEE_CARDS)); err != nil {
		}
	} else {
		robot.aiNotSeeAction()
	}
	//如果不看牌则执行不看牌操作
}

//ai决定是否要看牌
func (robot *Robot) aiIsSeeCard() bool {
	if robot.game.Round < 2 {
		return false
	}

	if robot.user.AiCharacter == nil {
		return false
	}
	rate := robot.user.AiCharacter.SeeCardsBase
	//当前自己的牌序对应概率
	switch robot.user.CardIndexInTable {
	case 1:
		rate += robot.user.AiCharacter.SeeCardsCardsIndex1
	case 2:
		rate += robot.user.AiCharacter.SeeCardsCardsIndex2
	case 3:
		rate += robot.user.AiCharacter.SeeCardsCardsIndex3
	case 4:
		rate += robot.user.AiCharacter.SeeCardsCardsIndex4
	case 5:
		rate += robot.user.AiCharacter.SeeCardsCardsIndex5
	}
	//看牌人数概率
	sawCount := robot.game.getSawCount()
	switch sawCount {
	case 1:
		rate += robot.user.AiCharacter.SeeCardsCount1
	case 2:
		rate += robot.user.AiCharacter.SeeCardsCount2
	case 3:
		rate += robot.user.AiCharacter.SeeCardsCount3
	case 4:
		rate += robot.user.AiCharacter.SeeCardsCount4
	}

	//看牌后加注人数
	sawRaiseCount := robot.game.getSawRaiseCount()
	switch sawRaiseCount {
	case 1:
		rate += robot.user.AiCharacter.SeeCardsRaise1
	case 2:
		rate += robot.user.AiCharacter.SeeCardsRaise2
	case 3:
		rate += robot.user.AiCharacter.SeeCardsRaise3
	case 4:
		rate += robot.user.AiCharacter.SeeCardsRaise4
	}

	//游戏小于当前轮数
	if robot.game.Round <= 2 {
		rate += robot.user.AiCharacter.SeeCardsRound3
	} else if robot.game.Round <= 5 {
		rate += robot.user.AiCharacter.SeeCardsRound5
	} else if robot.game.Round <= 10 {
		rate += robot.user.AiCharacter.SeeCardsRound10
	} else {
		rate += robot.user.AiCharacter.SeeCardsRound15
	}

	return robot.user.RateToExec(rate)
}

//ai没看牌要做的跟注、加注等操作
func (robot *Robot) aiNotSeeAction() {

	if robot.user.AiCharacter == nil {
		return
	}
	sawCount := robot.game.getSawCount() //看了牌都用户数量
	userList := robot.game.GetStatusUserList(global.USER_CUR_STATUS_ING)
	rateFollow := robot.user.AiCharacter.NotSeeBaseFollow
	rateRaise := robot.user.AiCharacter.NotSeeBaseRaise
	rateCompare := robot.user.AiCharacter.NotSeeBaseCompare

	//将各项加起来再做权重运算

	//跟注
	rateFollow += robot.user.AigetNotSeeFollowSawWeight(sawCount)
	rateFollow += robot.user.AigetNotSeeFollowCardsIndexWeight()
	rateFollow += robot.user.AigetNotSeeRemainFollowWeight(len(userList))
	rateFollow += robot.user.AigetNotSeeOtherCompareWeight(global.USER_OPTION_FOLLOW, robot.game.CompareCount)
	rateFollow += robot.user.AigetNotSeeSelfCompareWeight(global.USER_OPTION_FOLLOW)
	rateFollow += robot.user.AigetNotSeeRaiseAmountWeight(*robot.game.GameConfig, global.USER_OPTION_FOLLOW, robot.game.MinAction)
	rateFollow += robot.user.AigetNotSeeRoundWeight(global.USER_OPTION_FOLLOW, robot.game.Round)
	//加注
	////log.Traceln("not see 加注 base ::: ",rateRaise)
	rateRaise += robot.user.AigetNotSeeRaiseSawWeight(sawCount)
	rateRaise += robot.user.AigetNotSeeRaiseCardsIndexWeight()
	////log.Traceln("AigetNotSeeRaiseSawWeight  ",rateRaise)
	rateRaise += robot.user.AigetNotSeeRemainRaiseWeight(len(userList))
	////log.Traceln("AigetNotSeeRemainRaiseWeight  ",rateRaise)
	rateRaise += robot.user.AigetNotSeeOtherCompareWeight(global.USER_OPTION_RAISE, robot.game.CompareCount)
	////log.Traceln("AigetNotSeeOtherCompareWeight  ",rateRaise)
	rateRaise += robot.user.AigetNotSeeSelfCompareWeight(global.USER_OPTION_RAISE)
	////log.Traceln("AigetNotSeeSelfCompareWeight  ",rateRaise)
	rateRaise += robot.user.AigetNotSeeRaiseAmountWeight(*robot.game.GameConfig, global.USER_OPTION_RAISE, robot.game.MinAction)
	////log.Traceln("AigetNotSeeRaiseAmountWeight  ",rateRaise)
	rateRaise += robot.user.AigetNotSeeRoundWeight(global.USER_OPTION_RAISE, robot.game.Round)
	////log.Traceln("AigetNotSeeRoundWeight  ",rateRaise)
	//比牌
	rateCompare += robot.user.AigetNotSeeCompareSawWeight(sawCount)
	rateCompare += robot.user.AigetNotSeeCompareCardsIndexWeight()
	rateCompare += robot.user.AigetNotSeeRemainCompareWeight(len(userList))
	rateCompare += robot.user.AigetNotSeeOtherCompareWeight(global.USER_OPTION_COMPARE, robot.game.CompareCount)
	rateCompare += robot.user.AigetNotSeeSelfCompareWeight(global.USER_OPTION_COMPARE)
	rateCompare += robot.user.AigetNotSeeRaiseAmountWeight(*robot.game.GameConfig, global.USER_OPTION_COMPARE, robot.game.MinAction)
	rateCompare += robot.user.AigetNotSeeRoundWeight(global.USER_OPTION_COMPARE, robot.game.Round)
	////log.Traceln("ai未看牌，follow: ", rateFollow, " raise: ", rateRaise, " compare: ", rateCompare)
	robot.chooseOneToAction(rateFollow, rateRaise, rateCompare, 0)
}

//ai看过牌之后要做的跟注、加注等操作
func (robot *Robot) aiSawAction() {

	if robot.user.AiCharacter == nil {
		return
	}
	sawCount := robot.game.getSawCount() //看了牌都用户数量
	userList := robot.game.GetStatusUserList(global.USER_CUR_STATUS_ING)
	rateFollow := robot.user.AiCharacter.SawBaseFollow
	rateRaise := robot.user.AiCharacter.SawBaseRaise
	rateCompare := robot.user.AiCharacter.SawBaseCompare
	//log.Traceln("rateCompare11111 : ",rateCompare)
	rateGiveUp := robot.user.AiCharacter.SawBaseGiveUp

	//将各项加起来再做权重运算

	//跟注
	rateFollow += robot.user.AigetSawFollowSawWeight(sawCount)
	rateFollow += robot.user.AigetSawFollowCardsIndexWeight()
	rateFollow += robot.user.AigetSawRemainFollowWeight(len(userList))
	rateFollow += robot.user.AigetSawOtherCompareWeight(global.USER_OPTION_FOLLOW, robot.game.CompareCount)
	rateFollow += robot.user.AigetSawSelfCompareWeight(global.USER_OPTION_FOLLOW)
	rateFollow += robot.user.AigetSawRaiseAmountWeight(*robot.game.GameConfig, global.USER_OPTION_FOLLOW, robot.game.MinAction)
	rateFollow += robot.user.AigetSawRoundWeight(global.USER_OPTION_FOLLOW, robot.game.Round)
	rateFollow += robot.user.AigetSawCardType(global.USER_OPTION_FOLLOW)
	//加注
	rateRaise += robot.user.AigetSawRaiseSawWeight(sawCount)
	rateRaise += robot.user.AigetSawRaiseCardsIndexWeight()
	rateRaise += robot.user.AigetSawRemainRaiseWeight(len(userList))
	rateRaise += robot.user.AigetSawOtherCompareWeight(global.USER_OPTION_RAISE, robot.game.CompareCount)
	rateRaise += robot.user.AigetSawSelfCompareWeight(global.USER_OPTION_RAISE)
	rateRaise += robot.user.AigetSawRaiseAmountWeight(*robot.game.GameConfig, global.USER_OPTION_RAISE, robot.game.MinAction)
	rateRaise += robot.user.AigetSawRoundWeight(global.USER_OPTION_RAISE, robot.game.Round)
	rateRaise += robot.user.AigetSawCardType(global.USER_OPTION_RAISE)
	//比牌
	rateCompare += robot.user.AigetSawCompareSawWeight(sawCount)
	//log.Traceln("rateCompare2222: ",rateCompare)
	rateCompare += robot.user.AigetSawCompareCardsIndexWeight()
	rateCompare += robot.user.AigetSawRemainCompareWeight(len(userList))
	rateCompare += robot.user.AigetSawOtherCompareWeight(global.USER_OPTION_COMPARE, robot.game.CompareCount)
	rateCompare += robot.user.AigetSawSelfCompareWeight(global.USER_OPTION_COMPARE)
	rateCompare += robot.user.AigetSawRaiseAmountWeight(*robot.game.GameConfig, global.USER_OPTION_COMPARE, robot.game.MinAction)
	rateCompare += robot.user.AigetSawRoundWeight(global.USER_OPTION_COMPARE, robot.game.Round)
	rateCompare += robot.user.AigetSawCardType(global.USER_OPTION_COMPARE)
	//弃牌
	////log.Traceln("弃牌基础概率::::",rateGiveUp)
	rateGiveUp += robot.user.AigetSawGiveUpSawWeight(sawCount)
	rateGiveUp += robot.user.AigetSawGiveUpCardsIndexWeight()
	////log.Traceln("AigetSawGiveUpSawWeight >>> ",rateGiveUp)
	rateGiveUp += robot.user.AigetSawRemainGiveUpWeight(len(userList))
	////log.Traceln("AigetSawRemainGiveUpWeight >>> ",rateGiveUp)
	rateGiveUp += robot.user.AigetSawOtherCompareWeight(global.USER_OPTION_GIVE_UP, robot.game.CompareCount)
	////log.Traceln("AigetSawOtherCompareWeight >>> ",rateGiveUp)
	rateGiveUp += robot.user.AigetSawSelfCompareWeight(global.USER_OPTION_GIVE_UP)
	////log.Traceln("AigetSawSelfCompareWeight >>> ",rateGiveUp)
	rateGiveUp += robot.user.AigetSawRaiseAmountWeight(*robot.game.GameConfig, global.USER_OPTION_GIVE_UP, robot.game.MinAction)
	////log.Traceln("AigetSawRaiseAmountWeight >>> ",rateGiveUp)
	rateGiveUp += robot.user.AigetSawRoundWeight(global.USER_OPTION_GIVE_UP, robot.game.Round)
	////log.Traceln("AigetSawRoundWeight >>> ",rateGiveUp)
	rateGiveUp += robot.user.AigetSawCardType(global.USER_OPTION_GIVE_UP)
	////log.Traceln("AigetSawCardType >>> ",rateGiveUp)

	////log.Traceln("ai看了牌，follow: ", rateFollow, " raise: ", rateRaise, " compare: ", rateCompare, " give up: ", rateGiveUp)
	robot.chooseOneToAction(rateFollow, rateRaise, rateCompare, rateGiveUp)
}

func (robot *Robot) chooseOneToAction(rateFollow, rateRaise, rateCompare, rateGiveUp int) {
	if robot.game.MinAction >= robot.game.GameConfig.RaiseAmount[2] {
		rateRaise = 0
	}
	if robot.game.Round <= 3 {
		rateCompare = 0
	}
	totalRate := 0
	if rateFollow > 0 {
		totalRate += rateFollow
	}
	if rateRaise > 0 {
		totalRate += rateRaise
	}
	if rateCompare > 0 {
		totalRate += rateCompare
	}
	if rateGiveUp > 0 {
		totalRate += rateGiveUp
	}
	////log.Traceln("总概率》》》》》》: ", totalRate)
	////log.Traceln("看弃牌概率《《《《《《《 ", rateGiveUp, "   ", totalRate)
	if rateGiveUp > 0 && robot.user.RateToExecWithIn(rateGiveUp, totalRate) {
		//log.Traceln("机器人弃牌：：：", robot.AiUser.GetID())
		if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, 0, global.USER_OPTION_GIVE_UP)); err != nil {
		}
		//robot.Action(robot.CurActionUser, global.USER_OPTION_GIVE_UP, 0)
		return
	}
	////log.Traceln("看比牌概率《《《《《《《", rateCompare, "   ", totalRate)
	if rateCompare > 0 && robot.user.RateToExecWithIn(rateCompare, totalRate) && robot.game.Round < uint(robot.game.GameConfig.MaxRound)-1 {
		//log.Traceln("机器人比牌：：：", robot.AiUser.GetID())
		if robot.user.CardType >= poker.CardTypeSJ {
			log.Traceln("机器人比牌：：：", robot.AiUser.GetID(), "牌型超过了顺金：", robot.user.CardType)
			log.Traceln("比牌概率：" + fmt.Sprintf(`%d  %d`, rateCompare, totalRate))
			robot.game.Table.WriteLogs(robot.AiUser.GetID(), "比牌概率："+fmt.Sprintf(`%d  %d`, rateCompare, totalRate))
		}
		toCompareUser := robot.game.getCompareUser()
		////log.Traceln("机器人获取到的要比牌的玩家： ",toCompareUser)
		if len(toCompareUser) == 0 {
			//log.Traceln("机器人获取到的要比牌的玩家为空： ", toCompareUser)
			return
		}
		toCompareUser = append(toCompareUser, robot.user)
		if len(toCompareUser) == 0 {

		} else {
			if toCompareUser[0] != nil {
				////log.Traceln("机器人：",robot.CurActionUser.Id," 选择要比牌的：",toCompareUser[0].Id)
				robot.game.CompareCards(robot.user, toCompareUser[0], toCompareUser)
				if robot.game.IsSatisfyEnd() {
					//robot.Table.AddTimer(global.COMPARE_CARDS_DELAY, func() {
					time.AfterFunc(2*time.Second, func() {
						robot.game.EndGame(false)
					})
					return
				} else {
					//robot.SetNextActionUser("")
				}
			} else {
				//log.Traceln("居然为kong 机器人选择比牌玩家")
			}
			return
		}
	}
	////log.Traceln("看加注概率《《《《《《《", rateRaise, "   ", totalRate)
	if rateRaise > 0 && robot.user.RateToExecWithIn(rateRaise, totalRate) {
		//log.Traceln("机器人加注：：：", robot.AiUser.GetID())
		if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, robot.game.getRaiseAmount(), global.USER_OPTION_RAISE)); err != nil {
		}
		return
	}
	////log.Traceln("执行跟注《《《<<<")
	//log.Traceln("机器人跟注：：：", robot.AiUser.GetID())
	if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, 0, global.USER_OPTION_FOLLOW)); err != nil {
		log.Traceln("跟注err : ", err)
	}
}

//对方全押，考虑是否要全押
func (robot *Robot) aiAllIn() {
	////log.Traceln("有用户进入全押：：：：：：：")
	if !robot.user.IsSawCards {
		//看牌
		if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, 0, global.USER_OPTION_SEE_CARDS)); err != nil {
		}
		//robot.Action(robot.CurActionUser, global.USER_OPTION_SEE_CARDS, 0)
		return
	}
	//全押的话需要看当前是否为最大再决定是否跟注
	maxUser := poker.GetMaxUser(robot.game.GetStatusUserList(global.USER_CUR_STATUS_ING))
	if maxUser[0].Id != robot.user.Id {
		if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, 0, global.USER_OPTION_GIVE_UP)); err != nil {
		}
		//robot.Action(robot.CurActionUser, global.USER_OPTION_GIVE_UP, 0)
	} else {
		if robot.user.RateToExec(100) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, 0, global.USER_OPTION_FOLLOW)); err != nil {
			}
			//robot.Action(robot.CurActionUser, global.USER_OPTION_FOLLOW, 0)
		} else {
			if err := robot.AiUser.SendMsgToServer(global.C2S_USER_ACTION, NewC2sUserAction(robot.user.Id, 0, global.USER_OPTION_GIVE_UP)); err != nil {
			}
			//robot.Action(robot.CurActionUser, global.USER_OPTION_GIVE_UP, 0)
		}
	}
}
