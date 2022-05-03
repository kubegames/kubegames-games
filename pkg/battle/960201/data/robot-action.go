package data

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960201/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
)

//获取未看牌中 牌序
func (user *User) AigetNotSeeFollowCardsIndexWeight() (weight int) {

	switch user.CardIndexInTable {
	case 1:
		weight = user.AiCharacter.NotSeeFollowCardsIndex1
	case 2:
		weight = user.AiCharacter.NotSeeFollowCardsIndex2
	case 3:
		weight = user.AiCharacter.NotSeeFollowCardsIndex3
	case 4:
		weight = user.AiCharacter.NotSeeFollowCardsIndex4
	case 5:
		weight = user.AiCharacter.NotSeeFollowCardsIndex5
	}
	return
}
func (user *User) AigetNotSeeRaiseCardsIndexWeight() (weight int) {

	switch user.CardIndexInTable {
	case 1:
		weight = user.AiCharacter.NotSeeRaiseCardsIndex1
	case 2:
		weight = user.AiCharacter.NotSeeRaiseCardsIndex2
	case 3:
		weight = user.AiCharacter.NotSeeRaiseCardsIndex3
	case 4:
		weight = user.AiCharacter.NotSeeRaiseCardsIndex4
	case 5:
		weight = user.AiCharacter.NotSeeRaiseCardsIndex5
	}
	return
}
func (user *User) AigetNotSeeCompareCardsIndexWeight() (weight int) {

	switch user.CardIndexInTable {
	case 1:
		weight = user.AiCharacter.NotSeeCompareCardsIndex1
	case 2:
		weight = user.AiCharacter.NotSeeCompareCardsIndex2
	case 3:
		weight = user.AiCharacter.NotSeeCompareCardsIndex3
	case 4:
		weight = user.AiCharacter.NotSeeCompareCardsIndex4
	case 5:
		weight = user.AiCharacter.NotSeeCompareCardsIndex5
	}
	return
}

//获取未看牌中 看牌人数的权重-跟注
func (user *User) AigetNotSeeFollowSawWeight(sawCount int) (weight int) {

	switch sawCount {
	case 1:
		weight = user.AiCharacter.NotSeeFollowSaw1
	case 2:
		weight = user.AiCharacter.NotSeeFollowSaw2
	case 3:
		weight = user.AiCharacter.NotSeeFollowSaw3
	case 4:
		weight = user.AiCharacter.NotSeeFollowSaw4
	}
	return
}

//获取未看牌中 看牌人数的权重-加注
func (user *User) AigetNotSeeRaiseSawWeight(sawCount int) (weight int) {

	switch sawCount {
	case 1:
		weight = user.AiCharacter.NotSeeRaiseSaw1
	case 2:
		weight = user.AiCharacter.NotSeeRaiseSaw2
	case 3:
		weight = user.AiCharacter.NotSeeRaiseSaw3
	case 4:
		weight = user.AiCharacter.NotSeeRaiseSaw4
	}
	return
}

//获取未看牌 看牌人数的权重-比牌
func (user *User) AigetNotSeeCompareSawWeight(sawCount int) (weight int) {

	switch sawCount {
	case 1:
		weight = user.AiCharacter.NotSeeCompareSaw1
	case 2:
		weight = user.AiCharacter.NotSeeCompareSaw2
	case 3:
		weight = user.AiCharacter.NotSeeCompareSaw3
	case 4:
		weight = user.AiCharacter.NotSeeCompareSaw4
	}
	return
}

////////////////////剩余人数
//获取未看牌 剩余人数的权重-跟注
func (user *User) AigetNotSeeRemainFollowWeight(lenUserList int) (weight int) {

	switch lenUserList {
	case 2:
		weight = user.AiCharacter.NotSeeFollowRemain2
	case 3:
		weight = user.AiCharacter.NotSeeFollowRemain3
	case 4:
		weight = user.AiCharacter.NotSeeFollowRemain4
	case 5:
		weight = user.AiCharacter.NotSeeFollowRemain5
	}
	return
}

//获取未看牌 剩余人数的权重-加注
func (user *User) AigetNotSeeRemainRaiseWeight(lenUserList int) (weight int) {

	switch lenUserList {
	case 2:
		weight = user.AiCharacter.NotSeeRaiseRemain2
	case 3:
		weight = user.AiCharacter.NotSeeRaiseRemain3
	case 4:
		weight = user.AiCharacter.NotSeeRaiseRemain4
	case 5:
		weight = user.AiCharacter.NotSeeRaiseRemain5
	}
	return
}

//获取未看牌 剩余人数的权重-比牌
func (user *User) AigetNotSeeRemainCompareWeight(lenUserList int) (weight int) {

	switch lenUserList {
	case 2:
		weight = user.AiCharacter.NotSeeCompareRemain2
	case 3:
		weight = user.AiCharacter.NotSeeCompareRemain3
	case 4:
		weight = user.AiCharacter.NotSeeCompareRemain4
	case 5:
		weight = user.AiCharacter.NotSeeCompareRemain5
	}
	return
}

////////////////其他用户比牌胜利次数
func (user *User) AigetNotSeeOtherCompareWeight(option string, compareCount int) (weight int) {

	switch compareCount {
	case 1:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeOtherCompare1Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeOtherCompare1Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeOtherCompare1Compare
		}
	case 2:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeOtherCompare2Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeOtherCompare2Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeOtherCompare2Compare
		}
	case 3:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeOtherCompare3Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeOtherCompare3Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeOtherCompare3Compare
		}
	}
	return
}

///////////////自己比牌胜利次数
func (user *User) AigetNotSeeSelfCompareWeight(option string) (weight int) {

	switch user.CompareWinCount {
	case 1:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeSelfCompare1Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeSelfCompare1Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeSelfCompare1Compare
		}
	case 2:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeSelfCompare2Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeSelfCompare2Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeSelfCompare2Compare
		}
	case 3:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeSelfCompare3Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeSelfCompare3Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeSelfCompare3Compare
		}
	}
	return
}

////////////////当前加注挡位
func (user *User) AigetNotSeeRaiseAmountWeight(gameConfig config.GameConfig, option string, minAction int64) (weight int) {

	switch minAction {
	case gameConfig.RaiseAmount[0]:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeRaiseAmount1Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeRaiseAmount1Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeRaiseAmount1Compare
		}
	case gameConfig.RaiseAmount[1]:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeRaiseAmount2Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeRaiseAmount2Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeRaiseAmount2Compare
		}
	case gameConfig.RaiseAmount[2]:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeRaiseAmount3Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeRaiseAmount3Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeRaiseAmount3Compare
		}
	}
	return
}

////////////////当前轮数
func (user *User) AigetNotSeeRoundWeight(option string, round uint) (weight int) {

	if round <= 2 {
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeRound2Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeRound2Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeRound2Compare
		}
		return
	}
	if round <= 5 {
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeRound5Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeRound5Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeRound5Compare
		}
		return
	}
	if round <= 10 {
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeRound10Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeRound10Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeRound10Compare
		}
		return
	}

	if round <= 20 {
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.NotSeeRound20Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.NotSeeRound20Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.NotSeeRound20Compare
		}
	}
	return
}

/////////////////////////////////以下是看了牌的操作
//获取看牌中 牌序
func (user *User) AigetSawFollowCardsIndexWeight() (weight int) {

	switch user.CardIndexInTable {
	case 1:
		weight = user.AiCharacter.SawFollowCardsIndex1
	case 2:
		weight = user.AiCharacter.SawFollowCardsIndex2
	case 3:
		weight = user.AiCharacter.SawFollowCardsIndex3
	case 4:
		weight = user.AiCharacter.SawFollowCardsIndex4
	case 5:
		weight = user.AiCharacter.SawFollowCardsIndex5
	}
	return
}
func (user *User) AigetSawRaiseCardsIndexWeight() (weight int) {

	switch user.CardIndexInTable {
	case 1:
		weight = user.AiCharacter.SawRaiseCardsIndex1
	case 2:
		weight = user.AiCharacter.SawRaiseCardsIndex2
	case 3:
		weight = user.AiCharacter.SawRaiseCardsIndex3
	case 4:
		weight = user.AiCharacter.SawRaiseCardsIndex4
	case 5:
		weight = user.AiCharacter.SawRaiseCardsIndex5
	}
	return
}
func (user *User) AigetSawCompareCardsIndexWeight() (weight int) {

	switch user.CardIndexInTable {
	case 1:
		weight = user.AiCharacter.SawCompareCardsIndex1
	case 2:
		weight = user.AiCharacter.SawCompareCardsIndex2
	case 3:
		weight = user.AiCharacter.SawCompareCardsIndex3
	case 4:
		weight = user.AiCharacter.SawCompareCardsIndex4
	case 5:
		weight = user.AiCharacter.SawCompareCardsIndex5
	}
	return
}
func (user *User) AigetSawGiveUpCardsIndexWeight() (weight int) {

	switch user.CardIndexInTable {
	case 1:
		weight = user.AiCharacter.SawGiveUpCardsIndex1
	case 2:
		weight = user.AiCharacter.SawGiveUpCardsIndex2
	case 3:
		weight = user.AiCharacter.SawGiveUpCardsIndex3
	case 4:
		weight = user.AiCharacter.SawGiveUpCardsIndex4
	case 5:
		weight = user.AiCharacter.SawGiveUpCardsIndex5
	}
	return
}

//获取未看牌中 看牌人数的权重-跟注
func (user *User) AigetSawFollowSawWeight(sawCount int) (weight int) {

	switch sawCount {
	case 1:
		weight = user.AiCharacter.SawFollowSaw1
	case 2:
		weight = user.AiCharacter.SawFollowSaw2
	case 3:
		weight = user.AiCharacter.SawFollowSaw3
	case 4:
		weight = user.AiCharacter.SawFollowSaw4
	}
	return
}

//获取未看牌中 看牌人数的权重-加注
func (user *User) AigetSawRaiseSawWeight(sawCount int) (weight int) {

	switch sawCount {
	case 1:
		weight = user.AiCharacter.SawRaiseSaw1
	case 2:
		weight = user.AiCharacter.SawRaiseSaw2
	case 3:
		weight = user.AiCharacter.SawRaiseSaw3
	case 4:
		weight = user.AiCharacter.SawRaiseSaw4
	}
	return
}

//获取未看牌 看牌人数的权重-比牌
func (user *User) AigetSawCompareSawWeight(sawCount int) (weight int) {

	switch sawCount {
	case 1:
		weight = user.AiCharacter.SawCompareSaw1
	case 2:
		weight = user.AiCharacter.SawCompareSaw2
	case 3:
		weight = user.AiCharacter.SawCompareSaw3
	case 4:
		weight = user.AiCharacter.SawCompareSaw4
	}
	return
}

//获取未看牌 看牌人数的权重-弃牌
func (user *User) AigetSawGiveUpSawWeight(sawCount int) (weight int) {

	switch sawCount {
	case 1:
		weight = user.AiCharacter.SawGiveUpSaw1
	case 2:
		weight = user.AiCharacter.SawGiveUpSaw2
	case 3:
		weight = user.AiCharacter.SawGiveUpSaw3
	case 4:
		weight = user.AiCharacter.SawGiveUpSaw4
	}
	return
}

////////////////////剩余人数
//获取未看牌 剩余人数的权重-跟注
func (user *User) AigetSawRemainFollowWeight(lenUserList int) (weight int) {

	switch lenUserList {
	case 2:
		weight = user.AiCharacter.SawFollowRemain2
	case 3:
		weight = user.AiCharacter.SawFollowRemain3
	case 4:
		weight = user.AiCharacter.SawFollowRemain4
	case 5:
		weight = user.AiCharacter.SawFollowRemain5
	}
	return
}

//获取未看牌 剩余人数的权重-加注
func (user *User) AigetSawRemainRaiseWeight(lenUserList int) (weight int) {

	switch lenUserList {
	case 2:
		weight = user.AiCharacter.SawRaiseRemain2
	case 3:
		weight = user.AiCharacter.SawRaiseRemain3
	case 4:
		weight = user.AiCharacter.SawRaiseRemain4
	case 5:
		weight = user.AiCharacter.SawRaiseRemain5
	}
	return
}

//获取未看牌 剩余人数的权重-比牌
func (user *User) AigetSawRemainCompareWeight(lenUserList int) (weight int) {

	switch lenUserList {
	case 2:
		weight = user.AiCharacter.SawCompareRemain2
	case 3:
		weight = user.AiCharacter.SawCompareRemain3
	case 4:
		weight = user.AiCharacter.SawCompareRemain4
	case 5:
		weight = user.AiCharacter.SawCompareRemain5
	}
	return
}

//获取未看牌 剩余人数的权重-弃牌
func (user *User) AigetSawRemainGiveUpWeight(lenUserList int) (weight int) {

	switch lenUserList {
	case 2:
		weight = user.AiCharacter.SawGiveUpRemain2
	case 3:
		weight = user.AiCharacter.SawGiveUpRemain3
	case 4:
		weight = user.AiCharacter.SawGiveUpRemain4
	case 5:
		weight = user.AiCharacter.SawGiveUpRemain5
	}
	return
}

////////////////其他用户比牌胜利次数
func (user *User) AigetSawOtherCompareWeight(option string, compareCount int) (weight int) {

	switch compareCount {
	case 1:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawOtherCompare1Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawOtherCompare1Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawOtherCompare1Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawOtherCompare1GiveUp
		}
	case 2:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawOtherCompare2Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawOtherCompare2Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawOtherCompare2Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawOtherCompare2GiveUp
		}
	case 3:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawOtherCompare3Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawOtherCompare3Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawOtherCompare3Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawOtherCompare3GiveUp
		}
	}
	return
}

///////////////自己比牌胜利次数
func (user *User) AigetSawSelfCompareWeight(option string) (weight int) {

	switch user.CompareWinCount {
	case 1:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawSelfCompare1Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawSelfCompare1Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawSelfCompare1Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawSelfCompare1GiveUp
		}
	case 2:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawSelfCompare2Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawSelfCompare2Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawSelfCompare2Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawSelfCompare2GiveUp
		}
	case 3:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawSelfCompare3Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawSelfCompare3Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawSelfCompare3Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawSelfCompare3GiveUp
		}
	}
	return
}

////////////////当前加注挡位
func (user *User) AigetSawRaiseAmountWeight(gameConfig config.GameConfig, option string, minAction int64) (weight int) {

	switch minAction {
	case gameConfig.RaiseAmount[0]:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawRaiseAmount1Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawRaiseAmount1Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawRaiseAmount1Compare
		}
	case gameConfig.RaiseAmount[1]:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawRaiseAmount2Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawRaiseAmount2Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawRaiseAmount2Compare
		}
	case gameConfig.RaiseAmount[2]:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawRaiseAmount3Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawRaiseAmount3Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawRaiseAmount3Compare
		}
	}
	return
}

////////////////当前轮数
func (user *User) AigetSawRoundWeight(option string, round uint) (weight int) {

	if round <= 2 {
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawRound2Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawRound2Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawRound2Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawRound2GiveUp
		}
		//log.Traceln("当前轮数 2 rate ::::::::::: ",weight," option : ",option)
		return
	}
	if round <= 5 {
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawRound5Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawRound5Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawRound5Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawRound5GiveUp
		}
		//log.Traceln("当前轮数 5 rate ::::::::::: ",weight," option : ",option)
		return
	}
	if round <= 10 {
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawRound10Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawRound10Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawRound10Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawRound10GiveUp
		}
		//log.Traceln("当前轮数 10  rate ::::::::::: ",weight," option : ",option)
		return
	}

	if round <= 20 {
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawRound20Follow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawRound20Raise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawRound20Compare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawRound20GiveUp
		}
	}
	//log.Traceln("当前轮数 rate ::::::::::: ",weight," option : ",option)
	return
}

/////////////已看牌，机器人牌型
func (user *User) AigetSawCardType(option string) (weight int) {

	switch user.CardType {
	case global.CardTypeSingle:
		switch option {
		case global.USER_OPTION_FOLLOW:
			if user.HasCard(0xe0, user.Cards) {
				weight = user.AiCharacter.SawCardTypeSingleFollowOverA
			} else {
				weight = user.AiCharacter.SawCardTypeSingleFollowBelowA
			}
		case global.USER_OPTION_RAISE:
			if user.HasCard(0xe0, user.Cards) {
				weight = user.AiCharacter.SawCardTypeSingleRaiseOverA
			} else {
				weight = user.AiCharacter.SawCardTypeSingleRaiseBelowA
			}
		case global.USER_OPTION_COMPARE:
			if user.HasCard(0xe0, user.Cards) {
				weight = user.AiCharacter.SawCardTypeSingleCompareOverA
			} else {
				weight = user.AiCharacter.SawCardTypeSingleCompareBelowA
			}
		case global.USER_OPTION_GIVE_UP:
			if user.HasCard(0xe0, user.Cards) {
				weight = user.AiCharacter.SawCardTypeSingleGiveUpOverA
			} else {
				weight = user.AiCharacter.SawCardTypeSingleGiveUpBelowA
			}
		}
	case global.CardTypeDZ:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawCardTypeDzFollow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawCardTypeDzRaise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawCardTypeDzCompare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawCardTypeDzGiveUp

		}
	case global.CardTypeSZ:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawCardTypeSzFollow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawCardTypeSzRaise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawCardTypeSzCompare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawCardTypeSzGiveUp

		}
	case global.CardTypeSZA23:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawCardTypeSzFollow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawCardTypeSzRaise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawCardTypeSzCompare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawCardTypeSzGiveUp

		}
	case global.CardTypeJH:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawCardTypeJhFollow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawCardTypeJhRaise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawCardTypeJhCompare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawCardTypeJhGiveUp

		}
	case global.CardTypeSJ:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawCardTypeSjFollow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawCardTypeSjRaise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawCardTypeSjCompare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawCardTypeSjGiveUp

		}
	case global.CardTypeSJ123:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawCardTypeSjFollow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawCardTypeSjRaise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawCardTypeSjCompare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawCardTypeSjGiveUp

		}
	case global.CardTypeBZ:
		switch option {
		case global.USER_OPTION_FOLLOW:
			weight = user.AiCharacter.SawCardTypeBzFollow
		case global.USER_OPTION_RAISE:
			weight = user.AiCharacter.SawCardTypeBzRaise
		case global.USER_OPTION_COMPARE:
			weight = user.AiCharacter.SawCardTypeBzCompare
		case global.USER_OPTION_GIVE_UP:
			weight = user.AiCharacter.SawCardTypeBzGiveUp

		}
	}
	return
}
