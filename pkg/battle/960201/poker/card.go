package poker

import (
	"errors"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/poker"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

var Deck = []byte{
	0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1,
	0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2, 0xe2,
	0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3, 0xe3,
	0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4, 0xe4,
}

//测试牌型编码效率
func testCardEncode() {
	var cards []byte
	start := time.Now()
	count := 0
	for a := 0; a < 52-4; a++ {
		for b := a + 1; b < 52-3; b++ {
			for c := b + 1; c < 52-2; c++ {
				for d := c + 1; d < 52-1; d++ {
					for e := d + 1; e < 52; e++ {

						cards = []byte{Deck[a], Deck[b], Deck[c], Deck[d], Deck[e]}
						//encodeCard(getCardType(cards))
						count++
					}
				}
			}
		}
	}
	end := time.Now()
	log.Traceln("cards : ", cards)
	log.Traceln(end.Sub(start), count)
}

//比较得出手上牌最大的用户
func GetMaxUser(userList []*data.User) (maxUserRes []*data.User) {
	for _, user := range userList {
		cardType, cardSortRes := GetCardTypeJH(user.Cards)
		user.CardEncode = GetEncodeCard(cardType, cardSortRes)
		user.Cards = cardSortRes
		user.CardType = cardType
	}
	//先单独处理为对子的情况
	if isDz, maxRes := IsDzMax(userList); isDz {
		//log.Traceln("最大的牌型是对子")
		return maxRes
	}

	maxUserRes = make([]*data.User, 0)
	if len(userList) == 0 {
		log.Traceln("userlist 为空 ！！！！！！")
		return
	}
	maxReq := userList[0]

	//获取所有的牌的类型
	for i := 1; i < len(userList); i++ {
		if userList[i].CardEncode > maxReq.CardEncode {
			maxReq = userList[i]
			//比如 清空之前有两个顺子，结果后面的牌出现了同花
			if len(maxUserRes) != 0 {
				maxUserRes = make([]*data.User, 0)
			}
		} else if userList[i].CardEncode == maxReq.CardEncode {
			//相等则将当前牌型都存入进暂时的最大牌型数组
			maxUserRes = append(maxUserRes, userList[i])
		}
	}
	maxUserRes = append(maxUserRes, maxReq)
	return
}

//炸金花比较两副牌
func CompareZjhCards(cards1, cards2 []byte) bool {
	c1Type, c1CardsRes := GetCardTypeJH(cards1)
	c2Type, c2CardsRes := GetCardTypeJH(cards2)
	if c1Type == CardTypeDZ && c2Type == CardTypeDZ {
		//走对子的比牌流程
		c1Dz, _ := GetCardValueAndColor(c1CardsRes[1])
		c2Dz, _ := GetCardValueAndColor(c2CardsRes[1])
		if c1Dz > c2Dz {
			return true
		} else if c1Dz < c2Dz {
			return false
		} else {
			//return GetEncodeCard(c1Type,c1CardsRes) > GetEncodeCard(c2Type,c2CardsRes)
			if GetEncodeCard(c1Type, c1CardsRes) > GetEncodeCard(c2Type, c2CardsRes) {
				return true
			} else if GetEncodeCard(c1Type, c1CardsRes) < GetEncodeCard(c2Type, c2CardsRes) {
				return false
			} else {
				return cards1[0] > cards2[0]
			}
		}
	} else {

		if GetEncodeCard(c1Type, c1CardsRes) > GetEncodeCard(c2Type, c2CardsRes) {
			return true
		} else if GetEncodeCard(c1Type, c1CardsRes) < GetEncodeCard(c2Type, c2CardsRes) {
			return false
		} else {
			return cards1[0] > cards2[0]
		}
	}
}

//随便输入一个数字，得出它的牌值和花色
func GetCardValueAndColor(value byte) (cardValue, cardColor byte) {
	cardValue = value & 240 //byte的高4位总和是240
	cardColor = value & 15  //byte的低4位总和是15
	return
}

//将牌 进行牌型编码并返回
func GetEncodeCard(cardType int, cards []byte) (cardEncode int) {

	cardEncode = (cardType) << 20
	for i, card := range cards {
		cardEncode |= (int(card) >> 4) << uint((5-i-1)*4)
	}
	return
}

//将用户手上的牌进行编码
func userCardEncode(users []*data.User) {
	for _, user := range users {
		cardType, cardSortRes := GetCardTypeJH(user.Cards)
		user.CardEncode = GetEncodeCard(cardType, cardSortRes)
		user.Cards = cardSortRes
	}
}

//是否含有某张牌
func HasCard(card byte, cards []byte) (hasKing bool) {
	for _, v := range cards {
		cv, _ := GetCardValueAndColor(v)
		if cv == card {
			hasKing = true
			break
		}
	}
	return
}

//生成3张牌 TODO 暂时先随便生成3张牌，后面要根据相应的策略来生成牌
func GenerateCards() (cards []byte) {

	cards = make([]byte, 3)
	c0 := Deck[rand.RandInt(1, 52)]
	c1 := Deck[time.Now().UnixNano()%52]
	c2 := Deck[(time.Now().UnixNano()+123)%52]
	cards[0] = c0
	cards[1] = c1
	cards[2] = c2
	return
}

//如果是对子，则单独处理并返回结果
func IsDzMax(userList []*data.User) (isDz bool, maxUserRes []*data.User) {
	dzUsers := make([]*data.User, 0)
	for _, v := range userList {
		if v.CardType > CardTypeDZ {
			return false, nil
		}
		if v.CardType == CardTypeDZ {
			isDz = true
			dzUsers = append(dzUsers, v)
		}
	}
	if !isDz {
		return false, nil
	}

	//如果只有一个玩家是对子，说明就是最大的，直接返回结果
	if len(dzUsers) == 1 {
		log.Traceln("只有一个对子")
		return true, dzUsers
	}

	maxUserRes = make([]*data.User, 0)
	maxUser := dzUsers[0]
	for _, v := range dzUsers {
		maxDzCv, _ := GetCardValueAndColor(maxUser.Cards[1])
		//金花牌型三张牌，中间那张牌一定是对子的那张
		cv, _ := GetCardValueAndColor(v.Cards[1])
		if maxDzCv < cv {
			maxUser = v
		} else if maxDzCv == cv {
			//相同对子，对牌进行编码
			maxDzCvEncode, cvEncode := poker.GetEncodeCard(poker.CardTypeDZ, maxUser.Cards), poker.GetEncodeCard(poker.CardTypeDZ, v.Cards)
			if maxDzCvEncode < cvEncode {
				maxUser = v
			}
		}
	}
	maxUserRes = append(maxUserRes, maxUser)
	return true, maxUserRes
}

//
////获取对子牌型中的对子
//func GetDzCardDz(cards []byte) byte {
//	cards = sortCards(cards)
//	for _,card := range cards{
//		cv,_ := GetCardValueAndColor(card)
//
//	}
//}

//获取指定牌型的牌
func GetCardTypeCards(cardType int) (cards []byte) {
	//cards = make([]byte,3)
	switch cardType {
	case CardTypeBZ:
		cards = GetCardsBz()
	case CardTypeSJ:
		cards = GetCardsSj()
	case CardTypeJH:
		cards = GetCardsJh()
	case CardTypeSZ:
		cards = GetCardsSz()
	case CardTypeDZ:
		cards = GetCardsDz()
	case CardTypeSingle:
		cardsArr := make([][]byte, 0)
		cards1 := []byte{0x31, 0x41, 0x72}
		cards2 := []byte{0x31, 0x41, 0x82}
		cards3 := []byte{0x51, 0x41, 0x72}
		cards4 := []byte{0xd1, 0x41, 0x72}
		cards5 := []byte{0xe1, 0x41, 0x72}
		cards6 := []byte{0xa1, 0x41, 0x72}
		cardsArr = append(cardsArr, cards1)
		cardsArr = append(cardsArr, cards2)
		cardsArr = append(cardsArr, cards3)
		cardsArr = append(cardsArr, cards4)
		cardsArr = append(cardsArr, cards5)
		cardsArr = append(cardsArr, cards6)
		cards = cardsArr[rand.RandInt(0, 5)]
	}
	return
}

type CardsArray []byte

//随机获取豹子牌
func GetCardsBz() []byte {
	bzArr := make([]CardsArray, 0)
	cBase := []byte{0x21, 0x22, 0x23}
	bzArr = append(bzArr, cBase)
	var i byte = 1
	for i = 1; i <= 13; i++ {
		cv := []byte{cBase[0] + i*16, cBase[1] + i*16, cBase[2] + i*16}
		//log.Traceln(fmt.Sprintf(`%x`,cv))
		bzArr = append(bzArr, cv)
	}

	index := rand.RandInt(0, len(bzArr))
	return bzArr[index]
}

//随机获取顺金牌
func GetCardsSj() []byte {
	bzArr := make([]CardsArray, 0)
	cBase := []byte{0x21, 0x31, 0x41}
	bzArr = append(bzArr, cBase)
	var i byte = 1
	for i = 1; i < 12; i++ {
		cv := []byte{cBase[0] + i*16, cBase[1] + i*16, cBase[2] + i*16}
		//log.Traceln(fmt.Sprintf(`%x`,cv))
		bzArr = append(bzArr, cv)
	}

	index := rand.RandInt(0, len(bzArr))
	return bzArr[index]
}

//随机获取金花牌
func GetCardsJh() []byte {
	bzArr := make([]CardsArray, 0)
	cBase0 := []byte{0x31, 0x71, 0x41}
	cBase1 := []byte{0xd1, 0x21, 0x41}
	cBase2 := []byte{0xb1, 0xa1, 0x31}
	cBase3 := []byte{0x31, 0xa1, 0x61}
	cBase4 := []byte{0x31, 0x91, 0x71}
	cBase5 := []byte{0x31, 0x71, 0x81}
	cBase6 := []byte{0x51, 0x71, 0x91}
	cBase7 := []byte{0xb1, 0x71, 0xa1}
	bzArr = append(bzArr, cBase0)
	bzArr = append(bzArr, cBase1)
	bzArr = append(bzArr, cBase2)
	bzArr = append(bzArr, cBase3)
	bzArr = append(bzArr, cBase4)
	bzArr = append(bzArr, cBase5)
	bzArr = append(bzArr, cBase6)
	bzArr = append(bzArr, cBase7)

	index := rand.RandInt(0, len(bzArr))
	return bzArr[index]
}

//随机获取顺子牌
func GetCardsSz() []byte {
	bzArr := make([]CardsArray, 0)
	cBase := []byte{0x21, 0x32, 0x41}
	bzArr = append(bzArr, cBase)
	var i byte = 1
	for i = 1; i < 11; i++ {
		cv := []byte{cBase[0] + i*16, cBase[1] + i*16, cBase[2] + i*16}
		//log.Traceln(fmt.Sprintf(`%x`,cv))
		bzArr = append(bzArr, cv)
	}

	index := rand.RandInt(0, len(bzArr))
	return bzArr[index]
}

//随机获取对子牌
func GetCardsDz() []byte {
	bzArr := make([]CardsArray, 0)
	cBase := []byte{0x21, 0x22, 0x41}
	bzArr = append(bzArr, cBase)
	var i byte = 1
	for i = 1; i <= 13; i++ {
		cv := []byte{cBase[0] + i*16, cBase[1] + i*16, cBase[2] + i*16}
		if cv[2] > 0xe1 {
			cv[2] = 0xe3
		}
		//log.Traceln(fmt.Sprintf(`%x`,cv))
		bzArr = append(bzArr, cv)
	}

	index := rand.RandInt(0, len(bzArr))
	return bzArr[index]
}

//输入4张牌，获取牌型最大的牌
func GetMaxCardsIn4(arr []byte) (cards []byte, err error) {
	if len(arr) != 4 {
		log.Traceln("必须4张牌")
		err = errors.New("必须4张牌")
		return
	}
	maxCardType := CardTypeSingle - 1
	maxCards := make([]byte, 3)
	cards = make([]byte, 3)
	for i := 0; i < len(arr)-2; i++ {
		for j := i + 1; j < len(arr)-1; j++ {
			for k := j + 1; k < len(arr); k++ {
				//log.Traceln("4张牌中的三张：",fmt.Sprintf(`%x`,arr[i])," ",fmt.Sprintf(`%x`,arr[j])," ",fmt.Sprintf(`%x`,arr[k]))
				maxCards[0], maxCards[1], maxCards[2] = arr[i], arr[j], arr[k]
				cardType, _ := GetCardTypeJH(maxCards)
				if cardType > maxCardType {
					maxCardType = cardType
					cards[0], cards[1], cards[2] = arr[i], arr[j], arr[k]
				}
			}
		}
	}
	//log.Traceln("4选3返回的牌：：：",cards)
	return
}

func GetCardsCNName(cards []byte) (name string) {
	for _, v := range cards {
		name += getCardCNName(v)
	}
	return
}

//获取牌的中文名
func getCardCNName(card byte) string {
	switch card {
	case 0x21:
		return "方块2"
	case 0x22:
		return "樱花2"
	case 0x23:
		return "红桃2"
	case 0x24:
		return "黑桃2"
	case 0x31:
		return "方块3"
	case 0x32:
		return "樱花3"
	case 0x33:
		return "红桃3"
	case 0x34:
		return "黑桃3"
	case 0x41:
		return "方块4"
	case 0x42:
		return "樱花4"
	case 0x43:
		return "红桃4"
	case 0x44:
		return "黑桃4"
	case 0x51:
		return "方块5"
	case 0x52:
		return "樱花5"
	case 0x53:
		return "红桃5"
	case 0x54:
		return "黑桃5"
	case 0x61:
		return "方块6"
	case 0x62:
		return "樱花6"
	case 0x63:
		return "红桃6"
	case 0x64:
		return "黑桃6"
	case 0x71:
		return "方块7"
	case 0x72:
		return "樱花7"
	case 0x73:
		return "红桃7"
	case 0x74:
		return "黑桃7"
	case 0x81:
		return "方块8"
	case 0x82:
		return "樱花8"
	case 0x83:
		return "红桃8"
	case 0x84:
		return "黑桃8"
	case 0x91:
		return "方块9"
	case 0x92:
		return "樱花9"
	case 0x93:
		return "红桃9"
	case 0x94:
		return "黑桃9"
	case 0xa1:
		return "方块10"
	case 0xa2:
		return "樱花10"
	case 0xa3:
		return "红桃10"
	case 0xa4:
		return "黑桃10"
	case 0xb1:
		return "方块J"
	case 0xb2:
		return "樱花J"
	case 0xb3:
		return "红桃J"
	case 0xb4:
		return "黑桃J"
	case 0xc1:
		return "方块Q"
	case 0xc2:
		return "樱花Q"
	case 0xc3:
		return "红桃Q"
	case 0xc4:
		return "黑桃Q"
	case 0xd1:
		return "方块K"
	case 0xd2:
		return "樱花K"
	case 0xd3:
		return "红桃K"
	case 0xd4:
		return "黑桃K"
	case 0xe1:
		return "方块A"
	case 0xe2:
		return "樱花A"
	case 0xe3:
		return "红桃A"
	case 0xe4:
		return "黑桃A"
	}
	return "错误牌型"
}

func GetCardTypeCNName(cardType int) string {
	switch cardType {
	case CardTypeBZ:
		return "豹子"
	case 7:
		return "顺金"
	case 6:
		return "顺金A23"
	case 5:
		return "金花"
	case 4:
		return "顺子"
	case 3:
		return "顺子A23"
	case 2:
		return "对子"
	case 1:
		return "单张"
	}
	return "错误牌型"
}
