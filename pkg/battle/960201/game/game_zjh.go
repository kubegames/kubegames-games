package game

import (
	"errors"
	"fmt"
	rand2 "math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
)

//洗牌
func (gp *Game) ShuffleCards() {
	for i := 0; i < len(gp.Cards); i++ {
		index1 := rand2.Intn(len(gp.Cards))
		index2 := rand2.Intn(len(gp.Cards))
		gp.Cards[index1], gp.Cards[index2] = gp.Cards[index2], gp.Cards[index1]
	}
}

//获取牌
func (gp *Game) DealCards() byte {
	card := gp.Cards[0]
	gp.Cards = append(gp.Cards[:0], gp.Cards[1:]...)
	return card
}

//获取指定牌型的牌，随机30次，没获取到就随便返回三张 因为有顺子和顺子A23
func (gp *Game) GetSpeCardTypeCards(cardType, ct2 int) (cards []byte, cardTypeRes int) {
	cards = make([]byte, 3)
	for i := 0; i < 50; i++ {
		cards[0], cards[1], cards[2] = gp.Cards[0], gp.Cards[1], gp.Cards[2]
		if ct, _ := poker.GetCardTypeJH(cards); ct == cardType || ct == ct2 {
			gp.DealCards()
			gp.DealCards()
			gp.DealCards()
			cardTypeRes = cardType
			return
		} else {
			gp.ShuffleCards()
		}
	}
	//TODO 3月1号修改 没有随机发到指定牌型的牌就指定某几张牌
	//log.Traceln("走到3月1号修改 没有随机发到指定牌型的牌就指定某几张牌")
	cards = gp.getSpeCardTypeCards(cardType)
	log.Traceln("getSpeCardTypeCards : cardType", cardType, "cards : ", fmt.Sprintf(`%x`, cards))
	gp.DelCardInCards(cards[0])
	gp.DelCardInCards(cards[1])
	gp.DelCardInCards(cards[2])
	cardTypeRes, _ = poker.GetCardTypeJH(cards)
	return
}

func (game *Game) InitStartGameConfig() {
	game.HoseLampArr = game.Table.GetMarqueeConfig()
	//log.Traceln("下发的跑马灯配置：", fmt.Sprintf(`%+v`, game.HoseLampArr))
}

//开始游戏
func (zjh *Game) StartGame() {
	if zjh.CurStatus == global.TABLE_CUR_STATUS_START_SEND_CARD || zjh.CurStatus == global.TABLE_CUR_STATUS_ING {
		return
	}
	log.Traceln("开赛222 ", zjh.Table.GetID())
	zjh.Table.Broadcast(global.S2C_TICKER_START, &msg.S2CTickerStart{Ticker: global.START_GAME_TIME})
	zjh.Table.AddTimer(3*1000, func() {
		countUserNew := 0
		for _, user := range zjh.userListArr {
			if user != nil && user.CurStatus != global.USER_CUR_STATUS_ING && !user.IsLeave {
				countUserNew++
			}
		}
		log.Traceln("countUserNew ", countUserNew)
		if countUserNew <= 1 {
			log.Traceln("countUserNew <= 1 不开赛")
			zjh.CurStatus = global.TABLE_CUR_STATUS_WAIT
			zjh.Table.Broadcast(global.S2C_KEEP_MATCH, &msg.S2CUid{Uid: 1})
			return
		}

		zjh.Table.WriteLogs(0, "开始游戏，牌桌id : "+fmt.Sprintf(`%d`, zjh.Table.GetID())+"时间："+time.Now().Format("2006-1-2 15:04:05")+"参与该局的玩家有：")
		//if zjh.CurStatus != global.TABLE_CUR_STATUS_START_SEND_CARD && zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
		banker, err := zjh.GetBanker()
		if err != nil {
			banker = zjh.SetBanker(nil)
		}
		zjh.Round++
		zjh.CurActionUser = banker

		zjh.Table.StartGame()
		for _, v := range zjh.userListArr {
			if v == nil {
				continue
			}
			if v.IsLeave {
				continue
			}
			v.CurAmount += zjh.MinAction
			zjh.PoolAmount += zjh.MinAction
			v.Amount += zjh.MinAction
			v.Score -= zjh.MinAction
			//if !v.User.IsRobot(){
			//	log.Traceln("发送消息13，房间id：",zjh.Table.GetID(),"uid : ",v.User.GetID(),time.Now())
			//}
			v.CurStatus = global.USER_CUR_STATUS_ING
			//_=v.User.SendMsg(global.S2C_START_SEND_CARDS,zjh.GetTableInfo(v))
		}

		zjh.CurStatus = global.TABLE_CUR_STATUS_START_SEND_CARD
		zjh.BroadMsgSelfFirst(global.S2C_START_SEND_CARDS)
		//开始玩家发言倒计时
		zjh.Table.AddTimer(8000, func() {
			if zjh.CurStatus == global.TABLE_CUR_STATUS_START_SEND_CARD {
				//log.Traceln("8s之后客户端还没发送开始发牌，则主动发牌")
				zjh.SendCardOver(zjh.Banker)
			}
		})
		//}
	})
}

//发牌
func (zjh *Game) SendCards() {
	//给桌上每个玩家发牌 有4个人则发8副牌，选出最大的4副，其他人数同理
	userCount := zjh.GetTableUserCount()
	cardsArr := make([][]byte, 0)

	//3月1号修改的发牌
	ctCountMap := make(map[int]int)
	for i := 0; i < 5; i++ {
		//game.CardTypeCountTotal = 0
		zjh.CardTypeCountTotal = zjh.CardTypeCountMap[poker.CardTypeSingle][ctCountMap[poker.CardTypeSingle]] + zjh.CardTypeCountMap[poker.CardTypeDZ][ctCountMap[poker.CardTypeDZ]] +
			zjh.CardTypeCountMap[poker.CardTypeSZ][ctCountMap[poker.CardTypeSZ]] + zjh.CardTypeCountMap[poker.CardTypeJH][ctCountMap[poker.CardTypeJH]] +
			zjh.CardTypeCountMap[poker.CardTypeSJ][ctCountMap[poker.CardTypeSJ]] + zjh.CardTypeCountMap[poker.CardTypeBZ][ctCountMap[poker.CardTypeBZ]]
		singleRate := zjh.CardTypeCountMap[poker.CardTypeSingle][ctCountMap[poker.CardTypeSingle]]
		dzRate := zjh.CardTypeCountMap[poker.CardTypeDZ][ctCountMap[poker.CardTypeDZ]]
		szRate := zjh.CardTypeCountMap[poker.CardTypeSZ][ctCountMap[poker.CardTypeSZ]]
		jhRate := zjh.CardTypeCountMap[poker.CardTypeJH][ctCountMap[poker.CardTypeJH]]
		sjRate := zjh.CardTypeCountMap[poker.CardTypeSJ][ctCountMap[poker.CardTypeSJ]]
		bzRate := zjh.CardTypeCountMap[poker.CardTypeBZ][ctCountMap[poker.CardTypeBZ]]
		bz := bzRate
		sj := bzRate + sjRate
		jh := sj + jhRate
		sz := jh + szRate
		dz := sz + dzRate
		single := dz + singleRate
		index := rand.RandInt(0, zjh.CardTypeCountTotal)
		//log.Traceln("CardTypeCountTotal 概率：：：：：",index,zjh.CardTypeCountTotal,bz,sj,jh,sz,dz,single)
		switch {
		case index <= bz:
			//发豹子
			log.Traceln("豹子rate：", bzRate, zjh.CardTypeCountTotal)
			cards, ct := zjh.GetSpeCardTypeCards(poker.CardTypeBZ, poker.CardTypeBZ)
			ctCountMap[ct]++
			cardsArr = append(cardsArr, cards)
		case index > bz && index <= sj:
			//发顺金
			cards, ct := zjh.GetSpeCardTypeCards(poker.CardTypeSJ, poker.CardTypeSJ123)
			ctCountMap[ct]++
			cardsArr = append(cardsArr, cards)
		case index > sj && index <= jh:
			//发金花
			cards, ct := zjh.GetSpeCardTypeCards(poker.CardTypeJH, poker.CardTypeJH)
			ctCountMap[ct]++
			cardsArr = append(cardsArr, cards)
		case index > jh && index <= sz:
			//发顺子
			cards, ct := zjh.GetSpeCardTypeCards(poker.CardTypeSZ, poker.CardTypeSZA23)
			ctCountMap[ct]++
			cardsArr = append(cardsArr, cards)
		case index > sz && index <= dz:
			//发对子
			cards, ct := zjh.GetSpeCardTypeCards(poker.CardTypeDZ, poker.CardTypeDZ)
			ctCountMap[ct]++
			cardsArr = append(cardsArr, cards)
		case index > dz && index <= single:
			//发单张
			cards, ct := zjh.GetSpeCardTypeCards(poker.CardTypeSingle, poker.CardTypeSingle)
			ctCountMap[ct]++
			cardsArr = append(cardsArr, cards)
		default:
			log.Traceln(" index : ", index, zjh.CardTypeCountTotal)
			//panic("de")
		}
	}
	//3月1号修改的发牌

	cardsArr = poker.SortCardsArrFromBig(cardsArr)

	prob := zjh.Table.GetRoomProb()
	if prob == 0 {
		zjh.Table.WriteLogs(0, "获取到系统作弊率为0，实际就使用1000作为房间作弊率")
		prob = 1000
	}
	for _, v := range zjh.userListArr {
		if v == nil {
			continue
		}
		if v.IsLeave {
			continue
		}
		v.CurStatus = global.USER_CUR_STATUS_ING

		v.SetCheatByChair(prob)
	}
	//根据作弊率分配牌
	log.Traceln("得到的牌：", fmt.Sprintf(`%x`, cardsArr), "房间作弊率：：：：>>>>>>> ", prob)
	zjh.cheatMaxSecond(cardsArr, userCount)
	for _, user := range zjh.userListArr {
		if user != nil {
			log.Traceln("玩家：", user.Id, " 当前的牌在牌桌中第几大：", user.CardIndexInTable, "作弊率：", user.CheatRate)
			zjh.Table.WriteLogs(user.User.GetID(), fmt.Sprintf(`%s：%d,当前的牌在牌桌中第几大：%d  实际使用的作弊率：%d`, aiRealStr(user.User.IsRobot()), user.User.GetID(), user.CardIndexInTable, user.CheatRate))
		}
	}
}

//玩家发言，eg：弃牌就giveup,加注则需要传amount
func (zjh *Game) Action(user *data.User, option string, amount int64) (errcode int32) {
	if user == nil {
		return
	}
	if user.IsLeave {
		return
	}
	if zjh.IsAllUserAllIn() {
		log.Traceln("所有人都all in")
		return
	}
	if zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
		log.Traceln("游戏不在进行中")
		user.User.SendMsg(global.ERROR_CODE_GAME_NOT_ING, &msg.C2SIntoGame{})
		return
	}
	if user.CurStatus != global.USER_CUR_STATUS_ING {
		log.Traceln("用户当前状态为：", user.CurStatus, " 没在游戏中，不能发言")
		return
	}
	if user.User.GetID() == zjh.CompareIds[0] || user.User.GetID() == zjh.CompareIds[1] {
		log.Traceln("用户：", user.User.GetID(), " 比牌阶段不能操作: ", option)
		return
	}
	if zjh.CurActionUser.User.GetID() != user.User.GetID() && option != global.USER_OPTION_CANCEL_FOLLOW_ALL_THE_WAY && option != global.USER_OPTION_FOLLOW_ALL_THE_WAY && option != global.USER_OPTION_GIVE_UP {
		log.Traceln("当前发言玩家为: ", zjh.CurActionUser.Id, "  还未到该user :", user.Id, "发言")
		user.User.SendMsg(global.ERROR_CODE_NOT_CUR_USER_ACTION, &msg.C2SIntoGame{})
		return
	}
	switch option {
	//弃牌
	case global.USER_OPTION_GIVE_UP:
		zjh.Table.WriteLogs(user.User.GetID(), fmt.Sprintf(`%s用户id：%d 弃牌`, aiRealStr(user.User.IsRobot()), user.User.GetID()))
		log.Traceln("玩家", user.Id, " 发言：", option, " 弃牌", "是否机器人：", user.IsAi)
		zjh.userActionGiveUp(user)
		if zjh.IsSatisfyEnd() {
			zjh.EndGame(false)
			return
		}
		user.IsActioned = true
		if user.User.GetID() != zjh.CurActionUser.User.GetID() {
			return
		}
	//全押
	case global.USER_OPTION_ALL_IN:
		log.Traceln("玩家", user.Id, " 发言：", option, " 全押")
		if code := zjh.userActionAllIn(user); code != global.ERROR_CODE_OK {
			_ = user.User.SendMsg(code, user.GetUserMsgInfo(false))
			return
		}
		user.IsAllIn = true
		user.IsActioned = true
	//加注
	case global.USER_OPTION_RAISE:
		//log.Traceln("加注金额：", amount)
		////log.Traceln("玩家", user.Id, " 发言：", option, " 加注")
		if code := zjh.userActionRaise(user, amount); code != global.ERROR_CODE_OK {
			errcode = code
			_ = user.User.SendMsg(code, user.GetUserMsgInfo(false))
			return
		}
		user.IsActioned = true
	//看牌
	case global.USER_OPTION_SEE_CARDS:
		if zjh.Round <= 1 {
			log.Traceln("第一轮不能看牌")
			return
		}
		////log.Traceln("玩家", user.Id, " 发言：", option, " 看牌")
		if user.IsSawCards {
			//log.Traceln("该玩家看过牌了")
			return
		}
		zjh.Table.WriteLogs(user.User.GetID(), aiRealStr(user.User.IsRobot())+"用户id："+fmt.Sprintf(`%d`, user.User.GetID())+" 看牌 余额："+score.GetScoreStr(user.Score))
		zjh.userActionSeeCards(user)
	//跟注
	case global.USER_OPTION_FOLLOW:
		log.Traceln("玩家", user.Id, " 发言：", option, " 跟注")
		if code := zjh.userActionFollow(user); code != global.ERROR_CODE_OK {
			log.Traceln("跟注错误： ", code)
			_ = user.User.SendMsg(code, &msg.C2SIntoGame{})
			return
		}
		if zjh.IsAllIn {
			user.IsAllIn = true
		}
		user.IsActioned = true
		//todo 加上机器人换牌逻辑
		if user.User.IsRobot() {
			zjh.checkChangeCards(user)
		}
	//跟到底
	case global.USER_OPTION_FOLLOW_ALL_THE_WAY:
		//log.Traceln("玩家", user.Id, " 发言：", option, " 跟到底")
		//nowTime1 := time.Now()
		user.IsFollowAllTheWay = true
		_ = user.User.SendMsg(global.S2C_USER_ACTION, &msg.S2CUserActionInfo{UserId: user.Id, Option: option, Amount: zjh.MinAction,
			UserName: user.Name, CurAmount: user.CurAmount, PoolAmount: zjh.PoolAmount, MinAction: zjh.MinAction, Score: user.Score})
		if user == zjh.CurActionUser {
			zjh.userActionFollow(user)
			//log.Traceln("跟到底耗时1： ", time.Now().Sub(nowTime1))
		} else {
			return
		}
		user.IsActioned = true
		zjh.Table.WriteLogs(user.User.GetID(), aiRealStr(user.User.IsRobot())+"用户id："+fmt.Sprintf(`%d`, user.User.GetID())+" 跟到底 余额："+score.GetScoreStr(user.Score))
		////log.Traceln("跟到底耗时2： ", time.Now().Sub(nowTime1))
	//取消跟到底
	case global.USER_OPTION_CANCEL_FOLLOW_ALL_THE_WAY:
		//log.Traceln("玩家", user.Id, " 发言：", option, " 取消跟到底")
		zjh.userActionCancelFollowAllTheWay(user)
		////log.Traceln("取消跟到底耗时： ", time.Now().Sub(nowTime1))
		zjh.Table.WriteLogs(user.User.GetID(), aiRealStr(user.User.IsRobot())+"用户id："+fmt.Sprintf(`%d`, user.User.GetID())+" 取消跟到底 余额："+score.GetScoreStr(user.Score))
		return
	default:
		log.Traceln("非法user action :", option)
		return
	}

	//如果所有玩家都全押，就直接结束比赛
	allUserAllIn := zjh.IsAllUserAllIn()
	if allUserAllIn {
		log.Traceln("allUserAllIn ======== ")
		zjh.CompareCards(nil, nil, zjh.GetStatusUserList(global.USER_CUR_STATUS_ING))
		zjh.Table.AddTimer(2*1000, func() {
			zjh.EndGame(true)
		})
		return
	}

	//所有玩家都发过言了那就轮数++，并且把当前在玩儿的玩家发言状态置为false
	zjh.JudgeAndFinishCurRound()
	if zjh.CurStatus == global.TABLE_CUR_STATUS_ING {
		zjh.SetNextActionUser("Action", option)
	}
	log.Traceln("游戏状态：：：", zjh.CurStatus)
	return global.ERROR_CODE_OK
}

//设置下一个发言玩家
func (zjh *Game) SetNextActionUser(from, option string) {
	zjh.CompareIds[0], zjh.CompareIds[1] = 0, 0
	if zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
		log.Traceln("游戏不在进行 nil --------")
		return
	}
	if zjh.Round > uint(zjh.GameConfig.MaxRound) {
		log.Traceln("已经超过最后一轮了，不再设置")
		return
	}
	if zjh.CurActionUser == nil {
		log.Traceln("zjh.CurActionUser == nil")
		return
	}
	if zjh.CurActionUser.IsLeave {
		log.Traceln("zjh.CurActionUser.IsLeave")
		return
	}
	nextChair := zjh.CurActionUser.ChairId
	//如果是看牌的话，就依旧保持当前玩家发言
	if option != global.USER_OPTION_SEE_CARDS {
		//玩家发完言之后就设置下一个即将发言的玩家
		nextChair = zjh.GetNextActionChair()
		zjh.SetCurActionUser(zjh.GetUserByChairId(nextChair))
	}
	zjh.SetTicker(0, 10000)
	zjh.setCanAllinValue(zjh.CurActionUser)
	zjh.CurActionUser.CurRaiseAmount = zjh.getMaxRaiseAmount()
	//最后一轮第一个发言的玩家
	//isLastRound := false
	if int32(zjh.Round) == zjh.GameConfig.MaxRound && !zjh.IsSetLastRoundUser {
		zjh.CurActionUser.IsLastActionUser = true
		//isLastRound = true
		zjh.IsSetLastRoundUser = true
	}
	zjh.Table.Broadcast(global.S2C_CUR_ACTION_USER, zjh.CurActionUser.GetUserMsgInfo(false))
	//log.Traceln("当前发言玩家>>> ",zjh.CurActionUser.User.GetID())
	//如果玩家跟到底，则直接进行下一个玩家发言
	if zjh.CurActionUser.IsFollowAllTheWay {
		zjh.Table.AddTimer(2000, func() {
			//time.Sleep(2*time.Second)
			if zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
				return
			}
			//log.Traceln("玩家自动跟注：",zjh.CurActionUser.User.GetID())
			if zjh.CurActionUser.IsFollowAllTheWay {
				zjh.Action(zjh.CurActionUser, global.USER_OPTION_FOLLOW, 0)
			}
		})
	}

}

//根据chairid获取玩家
func (zjh *Game) GetUserByChairId(chairId uint) *data.User {
	for _, v := range zjh.userListArr {
		if v != nil && v.ChairId == chairId {
			return v
		}
	}
	return nil
}

//比牌，返回获胜的用户
//如果user不为空，则user是发起者
func (zjh *Game) CompareCards(initiator, comparedUser *data.User, userList []*data.User) (winners []*data.User) {
	if userList == nil || len(userList) == 0 {
		return
	}
	if initiator != nil && comparedUser == nil {
		log.Traceln("initiator != nil && comparedUser == nil")
		return
	}
	if initiator == nil {
		zjh.Table.WriteLogs(0, "系统比牌")
	} else {
		var compareId int64 = 0
		if comparedUser != nil {
			compareId = comparedUser.User.GetID()
		}
		zjh.Table.WriteLogs(initiator.User.GetID(), fmt.Sprintf(`%s用户id：%d 发起比牌，对方用户id：%d`, aiRealStr(initiator.User.IsRobot()), initiator.User.GetID(), compareId))
	}
	zjh.CompareCount++
	//将用户都装进彼此比牌的数组中
	for _, useri := range userList {
		for _, userj := range userList {
			if !hasUser(userj, useri.ComparedUsers) {
				useri.ComparedUsers = append(useri.ComparedUsers, userj)
			}
		}
	}

	loserIds := make([]int64, 0)
	losers := make([]*data.User, 0)
	//比牌都双倍，如果看牌就4倍
	if initiator != nil {
		initiator.IsActioned = true
		needAmount := zjh.MinAction * 2
		if initiator.IsSawCards {
			needAmount *= 2
		}
		initiator.Amount += needAmount
		initiator.Score -= needAmount
		//if _, err := initiator.User.SetScore(initiator.Table.GetGameNum(), -needAmount, "比牌",
		//	zjh.Table.GetRoomRate(), 0, global.SET_SCORE_BET, 100104); err != nil {
		//	log.Tracef("user raise set score err : ", err)
		//}
		initiator.CurAmount += needAmount
		zjh.PoolAmount += needAmount

	}

	winners = poker.GetMaxUser(userList)
	for _, v := range userList {
		if v.CurStatus == global.USER_CUR_STATUS_ING && v.Id != winners[0].Id {
			v.CurStatus = global.USER_CUR_STATUS_LOSE
			loserIds = append(loserIds, v.Id)
			losers = append(losers, v)
		}
	}

	winners[0].CompareWinCount++

	//组装返回给客户端的信息
	userResList := &msg.S2CCompareRes{
		WinId: winners[0].Id, LoseIds: loserIds, IsSystem: true,
	}
	////log.Traceln("比玩牌  输家id： ", userResList.LoseIds)
	if initiator != nil {
		userResList.FirstId = initiator.Id
		userResList.FirstAmount = initiator.Score
		userResList.IsSystem = false
	}
	if comparedUser != nil {
		userResList.ComparedId = comparedUser.User.GetID()
	}
	if initiator != nil {
		////log.Traceln("比牌结束，发起者： ", initiator.Id, " 广播。。。")
		zjh.Table.Broadcast(global.S2C_COMPARE_CARDS, userResList)
		zjh.Table.AddTimer(3*1000, func() {
			if !zjh.IsSatisfyEnd() {
				for _, loser := range losers {
					selfMsg := &msg.S2CUserSeeCards{UserId: loser.Id, CardType: int32(loser.CardType), Cards: loser.Cards}
					loser.User.SendMsg(global.S2C_GIVE_UP_CARDS_FOR_SELF, selfMsg)
				}
				zjh.SetNextActionUser("338 CompareCards", "")
			}
		})

	} else {
		if !zjh.IsSatisfyEnd() {
			zjh.SetNextActionUser("420 CompareCards", "")
		}
	}

	for _, v := range zjh.userListArr {
		if v != nil && !v.IsLeave && v.CurStatus == global.USER_CUR_STATUS_ING && v.User.IsRobot() {
			zjh.checkChangeCards(v)
		}
	}

	return winners
}

//结束本局比赛，清空一些缓存之类的操作
//目前只有 比牌、弃牌需要判断是否符合结束比赛条件
func (zjh *Game) EndGame(isAutoCompare bool) {

	zjh.CurStatus = global.TABLE_CUR_STATUS_WAIT
	winner, err := zjh.getStatusUser(global.USER_CUR_STATUS_ING)
	if err != nil {
		log.Traceln("EndGame err :::: ", err)
		return
	}
	//战绩
	var records []*platform.PlayerRecord

	//玩家赢钱金额加上
	tax := zjh.Table.GetRoomRate()
	winnerOldScofre := winner.User.GetScore()
	//winnerOldScore := winner.Score
	winner.User.SetScore(winner.Table.GetGameNum(), zjh.PoolAmount-winner.CurAmount, tax)
	winAmount := winner.User.GetScore() - winnerOldScofre
	taxScore := (zjh.PoolAmount - winner.CurAmount) * tax / 10000
	log.Traceln("赢家税收：", taxScore, "产出：", zjh.PoolAmount-taxScore, "总投注：", winner.Amount, winner.CurAmount)
	//winner.User.SetBetsAmount(winner.Amount)
	//winner.User.SendRecord(zjh.Table.GetGameNum(), zjh.PoolAmount-taxScore)
	//winner.User.SendRecord(zjh.Table.GetGameNum(), zjh.PoolAmount-taxScore-winner.Amount, winner.Amount, taxScore, zjh.PoolAmount-taxScore, "赢家产出")
	if !winner.User.IsRobot() {
		records = append(records, &platform.PlayerRecord{
			PlayerID:     uint32(winner.User.GetID()),
			GameNum:      zjh.Table.GetGameNum(),
			ProfitAmount: zjh.PoolAmount - taxScore - winner.Amount,
			BetsAmount:   winner.Amount,
			DrawAmount:   taxScore,
			OutputAmount: zjh.PoolAmount - taxScore,
			Balance:      winner.User.GetScore(),
			UpdatedAt:    time.Now(),
			CreatedAt:    time.Now(),
		})
	}

	loserAmountList := make([]*msg.S2CUserAmount, 0)
	for _, user := range zjh.userListArr {
		if user == nil {
			continue
		}
		if user.Id != winner.Id && user.CurStatus != global.USER_CUR_STATUS_LOOK {
			zjh.Table.WriteLogs(user.Id, aiRealStr(user.User.IsRobot())+"输家："+fmt.Sprintf(`%d`, user.User.GetID())+" 输了："+score.GetScoreStr(user.CurAmount)+
				" 余额："+score.GetScoreStr(user.User.GetScore()-user.Amount))
			loserAmount := &msg.S2CUserAmount{UserId: user.Id, Amount: user.CurAmount, LoseReason: user.LoseReason}
			loserAmountList = append(loserAmountList, loserAmount)
		}
		if len(user.ComparedUsers) == 0 {
			user.ComparedUsers = make([]*data.User, 1)
			user.ComparedUsers[0] = user
		}
	}
	//给每个玩家发送消息
	//log.Traceln("税收：",tax, winAmount)

	//先结算 -----------
	for _, userv := range zjh.userListArr {
		if userv == nil {
			continue
		}
		if userv.User.GetID() != winner.User.GetID() && !userv.IsLeave {
			userv.User.SetScore(userv.Table.GetGameNum(), -userv.Amount, zjh.Table.GetRoomRate())
			userv.User.SendChip(userv.CurAmount)
			log.Traceln("为输家发送战绩, ", userv.User.GetID(), userv.Amount)
			//userv.User.SetEndCards(fmt.Sprintf(`玩家在该局投入%d元`, userv.Amount))
			//userv.User.SetBetsAmount(userv.Amount)
			if userv.Amount != 0 {
				//userv.User.SendRecord(zjh.Table.GetGameNum(), -userv.Amount, userv.Amount, 0, 0, fmt.Sprintf(`玩家在该局投入%d元`, userv.Amount))
				if !userv.User.IsRobot() {
					records = append(records, &platform.PlayerRecord{
						PlayerID:     uint32(userv.User.GetID()),
						GameNum:      zjh.Table.GetGameNum(),
						ProfitAmount: -userv.Amount,
						BetsAmount:   userv.Amount,
						DrawAmount:   0,
						OutputAmount: 0,
						Balance:      userv.User.GetScore(),
						UpdatedAt:    time.Now(),
						CreatedAt:    time.Now(),
					})
				}
			}
			userv.Amount = 0
		}
	}

	//发送战绩
	if len(records) > 0 {
		if _, err := zjh.Table.UploadPlayerRecord(records); err != nil {
			log.Warnf("upload player record error %s", err.Error())
		}
	}

	//先结算 +++++++++++++++
	for _, userv := range zjh.userListArr {
		if userv == nil {
			continue
		}
		s2cSeeList := make([]*msg.S2CUserSeeCards, 0)
		for _, comparedUserv := range userv.ComparedUsers {
			s2cSee := &msg.S2CUserSeeCards{UserId: comparedUserv.Id, Cards: comparedUserv.Cards, CardType: int32(comparedUserv.CardType), ChairId: int32(comparedUserv.ChairId)}
			s2cSeeList = append(s2cSeeList, s2cSee)
		}
		//log.Traceln("玩家：", userv.Id, "比过牌的玩家列表：", userv.ComparedUsers)
		endGameMsg := &msg.S2CEndGame{
			WinAmount: winAmount, Winner: winner.GetUserMsgInfo(false), ComparedUserArr: s2cSeeList,
			LoserAmount: loserAmountList, AllUserAmount: make([]*msg.S2CUserAmount, 0), IsAutoCompare: isAutoCompare,
		}

		for _, v := range zjh.userListArr {
			if v != nil {
				//log.Traceln("比赛结束，下分：",v.User.GetID(),v.Score,v.User.GetScore())
				endGameMsg.AllUserAmount = append(endGameMsg.AllUserAmount, &msg.S2CUserAmount{UserId: v.User.GetID(), Amount: v.User.GetScore()})
			}
		}
		if !userv.IsLeave {
			_ = userv.User.SendMsg(global.S2C_FINISH_GAME, endGameMsg)
		}
	}

	//跑马灯
	zjh.TriggerHorseLamp(winner, winAmount)
	zjh.Table.WriteLogs(winner.User.GetID(), "结束比赛，"+aiRealStr(winner.User.IsRobot())+"赢家："+fmt.Sprintf(`%d`, winner.User.GetID())+
		" 赢的钱："+score.GetScoreStr(winAmount)+" 余额："+score.GetScoreStr(winner.User.GetScore())+
		"时间："+time.Now().Format("2006-1-2 15:04:05"))
	//赢家使用底注为打码量
	winner.User.SendChip(zjh.GameConfig.RaiseAmount[0] / 2)
	zjh.Table.EndGame()
	zjh.ResetTable2()
}

//是否触发跑马灯,有特殊条件就是and，没有特殊条件满足触发金额即可
func (zjh *Game) TriggerHorseLamp(winner *data.User, winAmount int64) {
	isTriggerSpe := false
	for _, v := range zjh.HoseLampArr {
		if strings.TrimSpace(v.SpecialCondition) == "" {
			if winAmount >= v.AmountLimit && fmt.Sprintf(`%d`, zjh.Table.GetRoomID()) == v.RoomId {
				log.Traceln("创建没有特殊条件的跑马灯")
				if err := zjh.Table.CreateMarquee(winner.User.GetNike(), winAmount, zjh.GetSpecialSlogan(winner.CardType), v.RuleId); err != nil {
				}
				break
			}
		} else {
			log.Traceln("跑马灯有特殊条件 : ", strings.TrimSpace(v.SpecialCondition))
			specialArr := strings.Split(v.SpecialCondition, ",")
			for _, specialStr := range specialArr {
				specialInt, err := strconv.Atoi(specialStr)
				if err != nil {
					log.Traceln("strconv.Atoi err : ", specialStr)
					continue
				}
				if winAmount >= v.AmountLimit && ((specialInt+4 == winner.CardType) || (specialInt+5 == winner.CardType)) && fmt.Sprintf(`%d`, zjh.Table.GetRoomID()) == v.RoomId {
					//if winAmount >= v.AmountLimit && (specialInt == winner.CardType) && fmt.Sprintf(`%d`, zjh.Table.GetRoomID()) == v.RoomId {
					if err := zjh.Table.CreateMarquee(winner.User.GetNike(), winAmount, zjh.GetSpecialSlogan(winner.CardType), v.RuleId); err != nil {
					}
					isTriggerSpe = true
					break
				}
			}
			if isTriggerSpe {
				break
			}
		}

	}
}

//获取庄家
func (zjh *Game) GetBanker() (banker *data.User, err error) {
	if zjh.Banker != nil {
		return zjh.Banker, nil
	}
	return nil, errors.New("没有庄家")
}

//设置并返回该局的庄家
func (zjh *Game) SetBanker(user *data.User) *data.User {
	if user != nil {
		zjh.Banker = user
		return user
	}

	onlineUserArr := make([]*data.User, 0)
	for _, v := range zjh.userListArr {
		if v != nil {
			onlineUserArr = append(onlineUserArr, v)
		}
	}

	index := 0
	userCount := len(onlineUserArr)
	if userCount > 1 {
		index = rand.RandInt(0, userCount-1)
	}
	i := 0
	for _, v := range onlineUserArr {
		if i == index {
			zjh.Banker = v
			return v
		}
		i++
	}
	if len(onlineUserArr) >= 1 {
		return onlineUserArr[0]
	}
	//log.Traceln("  SetBanker  nil 。。。 ", index)
	return nil
}

//设置当前发言玩家
func (zjh *Game) SetCurActionUser(user *data.User) {
	zjh.CurActionUser = user
}

func (zjh *Game) SetTicker(nowSecond int, triggerTime time.Duration) {

	if zjh.timerJob != nil {
		zjh.Table.DeleteJob(zjh.timerJob)
	}
	zjh.timerJob, _ = zjh.Table.AddTimer(int64(triggerTime), func() {
		//log.Traceln("倒计时时间到。。。")
		zjh.procTimerEvent()
	})
	//log.Tracef("结束")
}

//是否满足结束游戏条件
func (zjh *Game) IsSatisfyEnd() bool {

	if int32(zjh.Round) > zjh.GameConfig.MaxRound {
		return true
	}
	userList := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
	//正在游戏的玩家数量为1就可以结束比赛
	if len(userList) <= 1 {
		return true
	}

	return false
}

//获取当前牌桌信息
func (zjh *Game) GetTableInfo(user *data.User) *msg.S2CTableInfo {
	userInfoArr := make([]*msg.S2CUserInfo, 0)
	for _, v := range zjh.userListArr {
		if v == nil {
			continue
		}
		userInfo := v.GetUserMsgInfo(false)
		if user.User.IsRobot() {
			//log.Traceln("用户：",userInfo.UserId,"curAction：",userInfo.CurActionAmount," amount : ",userInfo.Amount)
		}
		userInfoArr = append(userInfoArr, userInfo)
	}
	if user != nil && !user.IsLeave {
		for k, v := range userInfoArr {
			if v.UserId == user.Id {
				userInfoArr = append(userInfoArr[k:], userInfoArr[:k]...)
			}
		}
	}
	s2cTableInfo := &msg.S2CTableInfo{
		CurStatus:   int32(zjh.CurStatus),
		PoolAmount:  zjh.PoolAmount,
		UserInfoArr: userInfoArr,
		RaiseAmount: zjh.GameConfig.RaiseAmount,
		Round:       int32(zjh.Round),
		TotalRound:  zjh.GameConfig.MaxRound,
		MinAction:   zjh.MinAction,
		Level:       zjh.Table.GetLevel(),
		RoomId:      int64(zjh.Table.GetRoomID()),
		IsAllIn:     zjh.IsAllIn,
		IsActioned:  user.IsActioned,
		UserStatus:  int32(user.CurStatus),
	}
	//log.Traceln("s2cTableInfo : ", user.Id, s2cTableInfo)
	if user != nil && !user.IsLeave {
		s2cTableInfo.IsSawCards = user.IsSawCards
		if s2cTableInfo.IsSawCards {
			s2cTableInfo.UserCards = &msg.S2CUserSeeCards{
				UserId: user.Id, Cards: user.Cards, CardType: int32(user.CardType), ChairId: int32(user.ChairId),
			}
		}
	}

	banker, err := zjh.GetBanker()
	if err == nil {
		s2cTableInfo.Banker = banker.GetUserMsgInfo(false)
	}
	if zjh.CurActionUser != nil {
		s2cTableInfo.CurUserInfo = zjh.CurActionUser.GetUserMsgInfo(false)
	}
	if zjh.timerJob != nil {
		//log.Traceln("GetTimeDifference : ",zjh.timerJob.GetTimeDifference())
		s2cTableInfo.TriggerTime = int32(zjh.timerJob.GetTimeDifference() / 1000)    //10 - int32(zjh.nowSecond)    //int32(zjh.TriggerTime-time.Now().UnixNano()/1e6) / 1000
		s2cTableInfo.LeftActionTime = int32(zjh.timerJob.GetTimeDifference() / 1000) //int32(zjh.TriggerTime-time.Now().UnixNano()/1e6) / 1000
	}
	s2cTableInfo.Limit = zjh.Table.GetEntranceRestrictions()
	//log.Traceln("消息0： ",s2cTableInfo.UserInfoArr)
	////log.Traceln("user Id : ", user.User.GetID(), "game id : ", zjh.Table.GetID(), "s2cMinaction : ", s2cTableInfo.MinAction, " zjh minaction: ", zjh.MinAction)
	return s2cTableInfo
}

//给当前牌桌上信息 按当前玩家在第一个的顺序发下来
func (zjh *Game) BroadMsgSelfFirst(subCmd int32) {
	for _, v := range zjh.userListArr {
		if v == nil {
			continue
		}
		if v.IsLeave {
			continue
		}
		tableInfo := zjh.GetTableInfo(v)
		if err := v.User.SendMsg(subCmd, tableInfo); err != nil && !v.IsAi {
			//log.Traceln("BroadMsgSelfFirst err : ", err)
			continue
		}
	}
}

//广播
func (zjh *Game) Broadcast(subCmd int32, pb proto.Message) {
	for _, v := range zjh.userListArr {
		if v != nil && !v.IsLeave {
			_ = v.User.SendMsg(subCmd, pb)
		}
	}
}

//为玩家分配座椅号
func (zjh *Game) SetCurUserChairId(user *data.User) {
	//先获取出当前所有的空座位
	emptyChairArr := make([]uint, 0)
	for i := 1; i <= 5; i++ {
		if chairUser := zjh.GetUserByChairId(uint(i)); chairUser != nil {
			continue
		}
		emptyChairArr = append(emptyChairArr, uint(i))
		////log.Traceln("玩家: ", user.Id, " 座椅号：", user.ChairId)
	}
	index := 0
	if len(emptyChairArr) > 1 {
		index = rand.RandInt(0, len(emptyChairArr)-1)
	}
	user.ChairId = emptyChairArr[index]
}

//结束当前轮，要做一些比如清空用户的相关信息，轮数++等操作
func (zjh *Game) JudgeAndFinishCurRound() {
	if zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
		//log.Traceln("当前牌桌不在游戏中，JudgeAndFinishCurRound")
		return
	}
	userOnlineList := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
	for _, v := range userOnlineList {
		if !v.IsActioned {
			////log.Traceln("没发言玩家：", v.Id)
			return
		}
	}

	//判断如果出现有玩家余额小于minAction ，就结束比赛
	if !zjh.IsAllIn {
		for _, v := range userOnlineList {
			minAction := zjh.MinAction
			if v.IsSawCards {
				minAction *= 2
			}
			if v.Score < minAction {
				log.Traceln("玩家金币数小于最小下注金额。。。系统比牌，结束比赛")
				zjh.CompareCards(nil, nil, userOnlineList)
				if zjh.IsSatisfyEnd() {
					zjh.CurStatus = global.TABLE_CUR_STATUS_SYSTEM_COMPARE
					////log.Traceln("比完牌，结束比赛。。。")
					zjh.Table.AddTimer(global.COMPARE_CARDS_DELAY, func() {
						//time.AfterFunc(2*time.Second, func() {
						zjh.EndGame(true)
					})
				}
				return
			}
		}
	}

	zjh.Round++
	if int32(zjh.Round) == zjh.GameConfig.MaxRound {
		onlineUsers := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
		for _, v := range onlineUsers {
			v.IsFollowAllTheWay = false
		}
	}
	if int32(zjh.Round) > zjh.GameConfig.MaxRound {
		//比牌，再结束游戏
		onlineUsers := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
		zjh.CompareCards(nil, nil, onlineUsers)
		if zjh.IsSatisfyEnd() {
			//log.Traceln("最后一轮，比完牌，结束比赛。。。111")
			zjh.CurStatus = global.TABLE_CUR_STATUS_SYSTEM_COMPARE
			zjh.Table.AddTimer(global.COMPARE_CARDS_DELAY, func() {
				//time.AfterFunc(2*time.Second, func() {
				zjh.EndGame(true)
			})
		}
		return
	}
	for _, v := range userOnlineList {
		if v.CurStatus == global.USER_CUR_STATUS_ING {
			v.IsActioned = false
		}
	}
	zjh.BroadMsgSelfFirst(global.S2C_FINISH_CUR_ROUND)

}

//重置牌桌各个状态，保持为新房间
func (zjh *Game) ResetTable2() {
	//zjh.TriggerEvent = "游戏结束，房间为重置状态"
	zjh.CurStatus = global.TABLE_CUR_STATUS_WAIT
	zjh.CurActionUser = nil
	zjh.MinAction = zjh.GameConfig.MinAction
	zjh.MaxAction = zjh.GameConfig.MaxAllIn
	zjh.PoolAmount = 0
	zjh.Round = 0
	zjh.CurActionIsSawCard = false
	zjh.IsAllIn = false
	zjh.IsSetLastRoundUser = false
	zjh.isChangeCards = false
	//zjh.nowSecond = 0
	zjh.Cards = make([]byte, 0)
	for _, v := range poker.Deck {
		zjh.Cards = append(zjh.Cards, v)
	}
	zjh.ShuffleCards()
	zjh.ShuffleCards()
	////log.Traceln("总抽水：", global.ProfitTotal)
	//aiCount := 0 //机器人个数
	for _, v := range zjh.userListArr {
		if v == nil {
			continue
		}
		if v.IsLeave {
			continue
		}
		_ = v.User.SendMsg(global.S2C_LEAVE_TABLE, v.GetUserMsgInfo(false))
		zjh.Table.KickOut(v.User)
	}
	for i := range zjh.userListArr {
		zjh.userListArr[i] = nil
	}
	zjh.Banker = nil

	//关闭桌子
	zjh.Table.Close()
}

//客户端发牌动画结束，通知服务器开始倒计时
func (zjh *Game) SendCardOver(user *data.User) {
	zjh.lock.Lock()
	if zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
		////log.Traceln("。。。玩家发牌动画完成，进入玩家发言倒计时。。。")
		//给用户发牌
		zjh.SendCards()
		zjh.CurStatus = global.TABLE_CUR_STATUS_ING
		zjh.SetTicker(0, 10000)
		zjh.SetNextActionUser("SendCardOver", "")
	}
	zjh.lock.Unlock()
}

//获取制定状态的玩家列表
func (zjh *Game) GetStatusUserList(status int) (userList []*data.User) {
	userList = make([]*data.User, 0)
	for _, v := range zjh.userListArr {
		if v == nil {
			continue
		}
		if v.CurStatus == status {
			userList = append(userList, v)
		}
	}
	return
}

//true : 返回 "机器人"
func aiRealStr(flag bool) string {
	if flag {
		return "【机器人】"
	}
	return "【真实玩家】"
}
