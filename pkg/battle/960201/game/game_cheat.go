package game

import (
	"fmt"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

//TODO 3月1号新增
func (game *Game) IsCardInCards(card byte) bool {
	for _, v := range game.Cards {
		if v == card {
			return true
		}
	}
	return false
}

//TODO 3月1号新增
func (game *Game) DelCardInCards(card byte) {
	for k, v := range game.Cards {
		if v == card {
			game.Cards = append(game.Cards[:k], game.Cards[k+1:]...)
			return
		}
	}
}

func (game *Game) sendBz() (cards []byte) {
	cards = make([]byte, 3)
	cardsArr := make([][]byte, 0)
	var i byte
	count := 0
	for i = 0x21; i < 0xe1; i += 16 {
		//log.Traceln(fmt.Sprintf(`%x`,i))
		if count >= 14 {
			break
		}
		i2, i3, i4 := i+1, i+2, i+3
		if game.IsCardInCards(i) && game.IsCardInCards(i2) && game.IsCardInCards(i3) {
			cardsTmp := make([]byte, 3)
			cardsTmp[0], cardsTmp[1], cardsTmp[2] = i, i2, i3
			cardsArr = append(cardsArr, cardsTmp)
			count++
			continue
		}
		if game.IsCardInCards(i) && game.IsCardInCards(i3) && game.IsCardInCards(i4) {
			cardsTmp := make([]byte, 3)
			cardsTmp[0], cardsTmp[1], cardsTmp[2] = i, i3, i4
			cardsArr = append(cardsArr, cardsTmp)
			count++
			continue
		}
		if game.IsCardInCards(i2) && game.IsCardInCards(i3) && game.IsCardInCards(i4) {
			cardsTmp := make([]byte, 3)
			cardsTmp[0], cardsTmp[1], cardsTmp[2] = i2, i3, i4
			cardsArr = append(cardsArr, cardsTmp)
			count++
			continue
		}
	}
	if len(cardsArr) == 0 {
		log.Traceln("豹子 =======> len(cardsArr) == 0")
		cards[0], cards[1], cards[2] = game.Cards[0], game.Cards[1], game.Cards[2]
	} else {
		index := rand.RandInt(0, len(cardsArr)-1)
		log.Traceln("index : ", index, fmt.Sprintf(`%x`, cardsArr))
		cards = cardsArr[index]
		log.Traceln("cards ::: ", fmt.Sprintf(`%x`, cards))
	}
	return
}

//发顺金
func (game *Game) sendSj() (cards []byte) {
	cards = make([]byte, 3)
	cardsArr := make([][]byte, 0)
	var i byte
	count := 0
	for i = 0x21; i < 0xc1; i += 16 {
		if count >= 20 {
			break
		}
		second := i + 16
		third := i + 32
		if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
			cardsTmp := make([]byte, 3)
			cardsTmp[0], cardsTmp[1], cardsTmp[2] = i, second, third
			cardsArr = append(cardsArr, cardsTmp)
			count++
			continue
		}
	}
	count = 0
	for i = 0x22; i < 0xc2; i += 16 {
		if count >= 30 {
			break
		}
		second := i + 16
		third := i + 32
		if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
			cardsTmp := make([]byte, 3)
			cardsTmp[0], cardsTmp[1], cardsTmp[2] = i, second, third
			cardsArr = append(cardsArr, cardsTmp)
			count++
			continue
		}
	}
	count = 0
	for i = 0x23; i < 0xc3; i += 16 {
		if count >= 30 {
			break
		}
		second := i + 16
		third := i + 32
		if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
			cardsTmp := make([]byte, 3)
			cardsTmp[0], cardsTmp[1], cardsTmp[2] = i, second, third
			cardsArr = append(cardsArr, cardsTmp)
			count++
			continue
		}
	}
	if game.IsCardInCards(0xe1) && game.IsCardInCards(0x21) && game.IsCardInCards(0x31) {
		cardsTmp := make([]byte, 3)
		cardsTmp[0], cardsTmp[1], cardsTmp[2] = 0xe1, 0x21, 0x31
		cardsArr = append(cardsArr, cardsTmp)
	}
	if game.IsCardInCards(0xe2) && game.IsCardInCards(0x22) && game.IsCardInCards(0x32) {
		cardsTmp := make([]byte, 3)
		cardsTmp[0], cardsTmp[1], cardsTmp[2] = 0xe2, 0x22, 0x32
		cardsArr = append(cardsArr, cardsTmp)
	}
	if game.IsCardInCards(0xe3) && game.IsCardInCards(0x23) && game.IsCardInCards(0x33) {
		cardsTmp := make([]byte, 3)
		cardsTmp[0], cardsTmp[1], cardsTmp[2] = 0xe3, 0x23, 0x33
		cardsArr = append(cardsArr, cardsTmp)
	}
	if len(cardsArr) == 0 {
		log.Traceln("顺金 =======> len(cardsArr) == 0")
		cards[0], cards[1], cards[2] = game.Cards[0], game.Cards[1], game.Cards[2]
	} else {
		index := rand.RandInt(0, len(cardsArr)-1)
		cards = cardsArr[index]
		log.Traceln("发顺金index：", index, "cards ::: ", fmt.Sprintf(`%x`, cards))
	}
	return
}

//TODO 3月1号新增
func (game *Game) getSpeCardTypeCards(cardType int) (cards []byte) {
	cards = make([]byte, 3)
	switch cardType {
	case poker.CardTypeBZ:
		log.Traceln("发豹子")
		cards = game.sendBz()
		return
	case poker.CardTypeSJ:
		log.Traceln("发顺金")
		cards = game.sendSj()
		return
	case poker.CardTypeSJ123:
		cards = game.sendSj()
		return
	case poker.CardTypeJH:
		var i byte
		for i = 0x52; i < 0xa4; i += 16 {
			second := i + 32
			third := i + 64
			if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
				cards[0], cards[1], cards[2] = i, second, third
				return
			}
		}
		for i = 0x23; i < 0xb4; i += 16 {
			second := i + 16
			third := i + 64
			if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
				cards[0], cards[1], cards[2] = i, second, third
				return
			}
		}
		for i = 0x21; i < 0xb4; i += 16 {
			second := i + 16
			third := i + 64
			if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
				cards[0], cards[1], cards[2] = i, second, third
				return
			}
		}
	case poker.CardTypeSZ:
		var i byte
		for i = 0x21; i < 0xc4; i += 16 {
			second := i + 17
			third := i + 33
			if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
				cards[0], cards[1], cards[2] = i, second, third
				return
			}
		}
		for i = 0x22; i < 0xc4; i += 16 {
			second := i + 17
			third := i + 33
			if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
				cards[0], cards[1], cards[2] = i, second, third
				return
			}
		}
		if game.IsCardInCards(0xe1) && game.IsCardInCards(0x23) && game.IsCardInCards(0x34) {
			cards[0], cards[1], cards[2] = 0xe1, 0x23, 0x34
			return
		}
		if game.IsCardInCards(0xe2) && game.IsCardInCards(0x23) && game.IsCardInCards(0x34) {
			cards[0], cards[1], cards[2] = 0xe2, 0x23, 0x34
			return
		}
		if game.IsCardInCards(0xe3) && game.IsCardInCards(0x23) && game.IsCardInCards(0x34) {
			cards[0], cards[1], cards[2] = 0xe3, 0x23, 0x34
			return
		}

	case poker.CardTypeSZA23:
		return game.getSpeCardTypeCards(poker.CardTypeSZ)
	case poker.CardTypeDZ:
		var i byte
		for i = 0x21; i < 0xc4; i += 16 {
			second := i + 1
			third := i + 33
			if game.IsCardInCards(i) && game.IsCardInCards(second) && game.IsCardInCards(third) {
				cards[0], cards[1], cards[2] = i, second, third
				return
			}
		}
	case poker.CardTypeSingle:
		cards[0], cards[1], cards[2] = game.Cards[0], game.Cards[1], game.Cards[2]
		return
	}
	log.Traceln("要发的牌型 ==================  ：", cardType)
	//panic("没有发到指定的牌")
	return
}

//检查是否需要换牌
func (game *Game) checkChangeCards(ai *data.User) {
	if !ai.User.IsRobot() {
		log.Traceln("checkChangeCards 居然会不是机器人？？？")
		return
	}
	if game.isChangeCards {
		return
	}
	//满足既是吃分状态，同时又是真人玩家为最大牌
	needCheck := false
	var ncUser *data.User
	for _, v := range game.userListArr {
		//真人、没离开、最大牌、系统吃分
		if v != nil && !v.User.IsRobot() && !v.IsLeave && v.CardIndexInTable == 1 && v.CheatRate > 0 {
			needCheck = true
			ncUser = v
			break
		}
	}
	if !needCheck {
		return
	}
	//机器人总投注
	aiInvest := game.getAiInvestTotal()
	times := aiInvest / ncUser.Amount
	if times > config.ChangeCardsArr[len(config.ChangeCardsArr)-1].Times {
		log.Traceln("新增换牌》》》》》超过了最大倍数：", times)
		game.changeCardsForAi(ncUser, ai)
		return
	}
	for _, v := range config.ChangeCardsArr {
		if v.Times == times && v.Cheat == ncUser.CheatRate && rand.RateToExec(v.ChangeRate) {
			log.Traceln("新增换牌》》》》》times : ", times, "玩家作弊率：", ncUser.CheatRate, " 配置作弊率: ", v.Cheat, v.ChangeRate)
			game.changeCardsForAi(ncUser, ai)
			return
		}
	}
}

//给机器人换牌 比给定的玩家牌大 ai：需要换牌的机器人
func (game *Game) changeCardsForAi(ncUser *data.User, ai *data.User) {
	game.isChangeCards = true
	needCardType := ncUser.CardType + 1
	if ncUser.CardType == poker.CardTypeBZ {
		log.Traceln("新增换牌-----用户的牌为豹子")
		needCardType = poker.CardTypeBZ
	}
	cards, ct := game.GetSpeCardTypeCards(needCardType, needCardType)
	log.Traceln("新增换牌-----换牌前：", fmt.Sprintf(`%x %d`, ai.Cards, ai.CardType))
	ai.Cards, ai.CardType = cards, ct
	log.Traceln("新增换牌-----换牌后：", fmt.Sprintf(`%x %d`, ai.Cards, ai.CardType))
	game.Table.WriteLogs(ai.User.GetID(), aiRealStr(ai.User.IsRobot())+"用户id："+fmt.Sprintf(`%d`, ai.User.GetID())+
		" 玩家点控输状态下，机器人投入高于数倍玩家投入时，触发换牌，机器人换牌之后：的牌为： "+poker.GetCardsCNName(ai.Cards)+" 牌型："+poker.GetCardTypeCNName(ai.CardType))
	//todo 后续要删除
	//res := &msg.S2CSeeOtherCards{
	//	UserCards:make([]*msg.S2CUserSeeCards,0),
	//}
	//for _,v := range game.userListArr {
	//	if v != nil {
	//		res.UserCards = append(res.UserCards,&msg.S2CUserSeeCards{
	//			UserId:v.User.GetID(),ChairId:int32(v.ChairId),
	//			CardType:int32(v.CardType),Cards:v.Cards,
	//		})
	//	}
	//}
	//log.Traceln("换牌之后给ncUser：",ncUser.User.GetID()," 发送消息")
	//ncUser.User.SendMsg(global.S2C_SEE_OTHER_CARDS, res)

	return
}

//获取机器人投注总额
func (game *Game) getAiInvestTotal() (total int64) {
	for _, v := range game.userListArr {
		if v != nil && v.User.IsRobot() {
			total += v.Amount
		}
	}
	return
}
