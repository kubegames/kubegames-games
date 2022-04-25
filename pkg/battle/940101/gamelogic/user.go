package gamelogic

import (
	majiangcom "github.com/kubegames/kubegames-games/internal/pkg/majiang"
	"github.com/kubegames/kubegames-games/pkg/battle/940101/def"
	errenmajiang "github.com/kubegames/kubegames-games/pkg/battle/940101/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

type UserData struct {
	User            player.PlayerInterface       //玩家接口
	HandCards       [majiangcom.MaxCardValue]int //玩家手牌
	OutCards        [144]int32                   //用户出牌的牌
	OutCardsIndex   int                          //用户出牌的index
	PengCards       [4]int                       //碰牌
	PengPaiNumber   int                          //碰牌的数量
	GangPai         [4]int                       //玩家杠牌
	GangPaiNumber   int                          // 玩家杠牌数量
	AnGangPai       [4]int                       //暗杠的牌，这个在可能会算番
	AnGangPaiNumber int                          //暗杠牌数量
	ChiPai          [12]int                      //吃持牌
	ChiPaiNumber    int                          // 玩家吃牌的数量
	Hua             []int32                      //花牌
	Opt             int                          //玩家操作
	CurrOpt         int                          //记录玩家的当前操作
	LastAddCard     int                          //玩家摸牌记录
	SiGuiYiNumber   int                          //四归一数量
	TingPai         bool                         // 听牌
	TianTing        bool                         // 是否天听
	WanPengNumber   int                          // 手牌中万牌碰的数量
	WanGangNumber   int                          // 手牌中万牌杠的数量
	WanChiNumber    int                          // 手牌中万牌吃的数量
	WANNumber       int                          // 手牌中万牌的总数量
	ZiPengNumber    int                          // 手牌中字牌碰的数量
	ZiGangNumber    int                          // 手牌中字牌杠的数量
	ZINumber        int                          // 手牌中字牌的总数量
	AutoStatus      bool                         // 托管状态
	IsZiMo          bool                         // 是否自摸
	OptCards        []*errenmajiang.OptRetMsg    // 用户操作顺序的牌
	ReLineFlag      bool                         // 重连标识
	TestFlag        int32                        // 配牌标识
	CtrlHuCards     []int32                      // 可控制要胡的牌
	HuOptCard       int32                        // 最后操作的牌
	HasMingGang     bool                         //明杠
	TingPaiMsg      *errenmajiang.NoticeTingCrad //听牌消息
	CheatSrc        string                       //作弊来源
}

func (ud *UserData) MoPai(CardValue int) {
	huCard := int32(0)
	ud.Opt, huCard = majiangcom.CanHu(ud.HandCards, int32(CardValue))
	var GangCard [4]int
	var BaGang [4]int
	ud.Opt |= majiangcom.CanAnGang(ud.HandCards, CardValue, GangCard)
	ud.Opt |= majiangcom.CanMingGang(ud.PengCards, CardValue, ud.HandCards, BaGang)
	ud.IsChangeTing(GangCard, BaGang, 0)

	if ud.TingPai {
		if ud.OptCards[len(ud.OptCards)-1].Opt&0xE0 != 0 {
			ud.GetTingOdds()
		}
	}

	if CardValue != 0 {
		ud.HandCards[CardValue] += 1
		ud.LastAddCard = CardValue
	}

	if ud.Opt == 0 {
		ud.Opt = majiangcom.OptTypeOutCard
	}

	if ud.Opt&majiangcom.OptTypeHu != 0 && huCard > 0 {
		ud.HuOptCard = huCard
	}

	if CardValue >= majiangcom.Hua[0] && ud.TingPai {
		for i := range ud.TingPaiMsg.Cards[0].Fan {
			ud.TingPaiMsg.Cards[0].Fan[i] += 1
		}

		ud.User.SendMsg(int32(errenmajiang.ReMsgIDS2C_NoticeTing), ud.TingPaiMsg)
	}

	ud.GetHuOdds()
	log.Tracef("%v 用户的手牌为：%v %v %v ", ud.User.GetID(), majiangcom.GetHandCardString(ud.HandCards), ud.PengCards, ud.Opt)
	log.Tracef("杠牌%v, 吃牌%v", ud.GangPai, ud.ChiPai)
	log.Tracef("%v 打了那些牌 %v ", ud.User.GetID(), majiangcom.GetOutCardString(ud.OutCards))
}

var Cards = [16]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 31, 32, 33, 34, 35, 36, 37}

//获取听牌倍数
func (ud *UserData) GetTingOdds() {
	temp := *ud
	ud.TingPaiMsg = new(errenmajiang.NoticeTingCrad)

	msg := new(errenmajiang.TingCardMsg)
	for j := 0; j < 16; j++ {
		temp.HandCards = ud.HandCards
		ret, _ := majiangcom.CanHu(temp.HandCards, int32(Cards[j]))
		if ret != 0 {
			temp.HandCards[Cards[j]]++
			temp.HuOptCard = int32(Cards[j])
			mask, noMask := GetCardsOdds(&temp)
			// 没有字一色和清一色就是混一色，字一色和清一色都不记番混一色
			if temp.TianTing {
				mask |= def.TianTing
			} else {
				mask |= def.TingPai
			}
			if mask&def.ZiYiSe == 0 && mask&def.QingYiSe == 0 {
				mask |= def.HunYiSe
			}
			HuValues := mask & ^noMask
			// 总番数
			totalDouble := int32(0)
			for idx, val := range def.FanTypeArray {
				if HuValues&val != 0 {
					log.Tracef("%v", def.FanNameArray[idx])
					if val != def.SiGuiYi {
						totalDouble += def.FanDoubleArray[idx]
						log.Tracef("%v", def.FanDoubleArray[idx])
					} else {

						totalDouble += def.FanDoubleArray[idx] * int32(temp.SiGuiYiNumber)
					}
				}
			}

			if len(temp.Hua) > 0 {
				totalDouble += int32(len(temp.Hua))
			}

			if temp.GangPaiNumber+temp.PengPaiNumber+temp.ChiPaiNumber == 0 {
				mask |= def.MengQianQing
			}

			msg.Fan = append(msg.Fan, totalDouble)
			msg.OutCardValue = 0
			msg.HuCardValue = append(msg.HuCardValue, int32(Cards[j]))
		}
	}

	ud.TingPaiMsg.Cards = append(ud.TingPaiMsg.Cards, msg)
	log.Tracef("发送消息听牌 %v", ud.TingPaiMsg.Cards)
	ud.User.SendMsg(int32(errenmajiang.ReMsgIDS2C_NoticeTing), ud.TingPaiMsg)
}

//获取听牌倍数
func (ud *UserData) GetHuOdds() {
	if ud.TingPai {
		return
	}

	temp := *ud
	ud.TingPaiMsg = new(errenmajiang.NoticeTingCrad)
	for i := 0; i < 16; i++ {
		if temp.HandCards[Cards[i]] > 0 {
			msg := new(errenmajiang.TingCardMsg)
			for j := 0; j < 16; j++ {
				temp.HandCards = ud.HandCards
				temp.HandCards[Cards[i]]--
				ret, _ := majiangcom.CanHu(temp.HandCards, int32(Cards[j]))
				if ret != 0 {
					temp.HandCards[Cards[j]]++
					temp.HuOptCard = int32(Cards[j])
					mask, noMask := GetCardsOdds(&temp)
					// 没有字一色和清一色就是混一色，字一色和清一色都不记番混一色
					if mask&def.ZiYiSe == 0 && mask&def.QingYiSe == 0 {
						mask |= def.HunYiSe
					}
					HuValues := mask & ^noMask
					// 总番数
					totalDouble := int32(0)
					for idx, val := range def.FanTypeArray {
						if HuValues&val != 0 {
							if val != def.SiGuiYi {
								totalDouble += def.FanDoubleArray[idx]
							} else {

								totalDouble += def.FanDoubleArray[idx] * int32(temp.SiGuiYiNumber)
							}
						}
					}

					if len(temp.Hua) > 0 {
						totalDouble += int32(len(temp.Hua))
					}

					if temp.GangPaiNumber+temp.PengPaiNumber+temp.ChiPaiNumber == 0 {
						mask |= def.MengQianQing
					}

					msg.Fan = append(msg.Fan, totalDouble)
					msg.OutCardValue = int32(Cards[i])
					msg.HuCardValue = append(msg.HuCardValue, int32(Cards[j]))
				}
			}

			if msg.OutCardValue != 0 {
				log.Tracef("长度%v", len(ud.TingPaiMsg.Cards))
				ud.TingPaiMsg.Cards = append(ud.TingPaiMsg.Cards, msg)
			}
		}
	}

	log.Tracef("长度%v", len(ud.TingPaiMsg.Cards))
	if len(ud.TingPaiMsg.Cards) > 0 {
		log.Tracef("发送消息听牌 %v", ud.TingPaiMsg.Cards)
		ud.User.SendMsg(int32(errenmajiang.ReMsgIDS2C_NoticeTing), ud.TingPaiMsg)
	}
}

//获取玩家的操作，是否能吃
func (ud *UserData) GetUserOpt(bChi bool, CardValue int) {
	huCard := int32(0)
	ud.Opt, huCard = majiangcom.CanHu(ud.HandCards, int32(CardValue))
	ud.Opt = ud.Opt | majiangcom.CanChi(ud.HandCards, CardValue) | majiangcom.CanPengAndGang(ud.HandCards, CardValue)

	if ud.TingPai {
		ud.Opt &= ^(majiangcom.OptTypeYouChi | majiangcom.OptTypeZhongChi |
			majiangcom.OptTypeZuoChi | majiangcom.OptTypePeng)
	}

	var GangCard [4]int
	var BaGang [4]int
	ud.IsChangeTing(GangCard, BaGang, CardValue)

	log.Tracef("%v玩家有操作：%v %v %v", ud.User.GetID(), ud.Opt, majiangcom.GetHandCardString(ud.HandCards), CardValue)
	if ud.Opt&majiangcom.OptTypeHu != 0 && huCard > 0 {
		ud.HuOptCard = huCard
	}
}

//发送操作消息
func (ud *UserData) SendOptMsg(time int64) {
	msg := new(errenmajiang.OptMsg)
	msg.OptType = int32(ud.Opt)
	msg.Time = time
	ud.User.SendMsg(int32(errenmajiang.ReMsgIDS2C_Opt), msg)
}

//发送等待消息
func (ud *UserData) SendWaitMsg(time int64) {
	msg := new(errenmajiang.WaitMsg)
	msg.WaitTime = time
	ud.User.SendMsg(int32(errenmajiang.ReMsgIDS2C_Wait), msg)
}

//设置玩家的当前操作
func (ud *UserData) SetCurrOpt(Opt int) {
	ud.CurrOpt = Opt
	ud.Opt = 0
}

//吃牌
func (ud *UserData) SetChi(v int, opt int) {
	if opt == majiangcom.OptTypeZuoChi {
		ud.HandCards[v-2] -= 1
		ud.HandCards[v-1] -= 1
		ud.ChiPai[ud.ChiPaiNumber] = v - 2
		ud.ChiPai[ud.ChiPaiNumber+1] = v - 1
		ud.ChiPai[ud.ChiPaiNumber+2] = v
	} else if opt == majiangcom.OptTypeZhongChi {
		ud.HandCards[v-1] -= 1
		ud.HandCards[v+1] -= 1
		ud.ChiPai[ud.ChiPaiNumber] = v - 1
		ud.ChiPai[ud.ChiPaiNumber+1] = v
		ud.ChiPai[ud.ChiPaiNumber+2] = v + 1
	} else {
		ud.HandCards[v+1] -= 1
		ud.HandCards[v+2] -= 1
		ud.ChiPai[ud.ChiPaiNumber] = v
		ud.ChiPai[ud.ChiPaiNumber+1] = v + 1
		ud.ChiPai[ud.ChiPaiNumber+2] = v + 2
	}
	ud.ChiPaiNumber += 3
}

//碰牌
func (ud *UserData) SetPeng(v int) {
	if ud.HandCards[v] >= 2 {
		ud.HandCards[v] -= 2
		ud.PengCards[ud.PengPaiNumber] = v
		ud.PengPaiNumber++
	}

	log.Tracef("碰牌 %v", ud.PengCards)
}

//明杠
func (ud *UserData) SetMingGangPai(v int) bool {
	if ud.HandCards[v] > 0 {
		ud.HandCards[v] -= 1
	} else {
		return false
	}
	for index, peng := range ud.PengCards {
		if v == peng {
			ud.PengCards[index] = 0
			ud.GangPai[ud.GangPaiNumber] = v
			ud.GangPaiNumber++
			ud.PengPaiNumber--
			ud.HasMingGang = true
			return true
		}
	}
	return false
}

//一般杠
func (ud *UserData) SetGangPai(v int) {
	ud.HandCards[v] -= 3
	ud.GangPai[ud.GangPaiNumber] = v
	ud.GangPaiNumber++

}

//暗杠
func (ud *UserData) SetAnGangPai(v int) {
	ud.HandCards[v] -= 4
	ud.AnGangPai[ud.AnGangPaiNumber] = v
	ud.AnGangPaiNumber++
}

//重置玩家数据
func (ud *UserData) ResetData() {
	for i := 1; i < majiangcom.MaxCardValue; i++ {
		ud.HandCards[i] = 0
	}
	ud.HandCards = [majiangcom.MaxCardValue]int{} //玩家手牌
	ud.PengCards = [4]int{}                       //碰牌
	ud.PengPaiNumber = 0                          //碰牌的数量
	ud.GangPai = [4]int{}                         //玩家杠牌
	ud.GangPaiNumber = 0                          // 玩家杠牌数量
	ud.AnGangPai = [4]int{}                       //暗杠的牌，这个在可能会算番
	ud.AnGangPaiNumber = 0                        //暗杠牌数量
	ud.ChiPai = [12]int{}                         //吃持牌
	ud.ChiPaiNumber = 0                           // 玩家吃牌的数量
	ud.Hua = []int32{}                            //花牌
	ud.Opt = 0                                    //玩家操作
	ud.CurrOpt = 0                                //记录玩家的当前操作
	ud.LastAddCard = 0                            //玩家摸牌记录
	ud.SiGuiYiNumber = 0                          //四归一数量
	ud.TingPai = false                            // 听牌
	ud.TianTing = false                           // 是否天听
	ud.WanPengNumber = 0                          // 手牌中万牌碰的数量
	ud.WanGangNumber = 0                          // 手牌中万牌杠的数量
	ud.WanChiNumber = 0                           // 手牌中万牌吃的数量
	ud.WANNumber = 0                              // 手牌中万牌的总数量
	ud.ZiPengNumber = 0                           // 手牌中字牌碰的数量
	ud.ZiGangNumber = 0                           // 手牌中字牌杠的数量
	ud.ZINumber = 0                               // 手牌中字牌的总数量
	ud.AutoStatus = false                         // 托管状态
	ud.IsZiMo = false                             // 是否自摸
	ud.OptCards = []*errenmajiang.OptRetMsg{}     // 用户操作顺序的牌
	ud.ReLineFlag = false                         // 重连标识
	ud.OutCards = [144]int32{}
	ud.OutCardsIndex = 0
	ud.HasMingGang = false
}

func (ud *UserData) GetUserHasCardNumber() int {
	count := 0
	for _, val := range ud.HandCards {
		count += val
	}
	return count
}

func (ud *UserData) GetHanCardWANNumber(cards [majiangcom.MaxCardValue]int) {
	PengNum := 0
	GangNum := 0
	ShunNum := 0
	TotalNum := 0

	tempCards := cards

	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
		if cards[i] > 0 {
			TotalNum += cards[i]
		}
		if tempCards[i] == 3 {
			PengNum++
			tempCards[i] -= 3
		} else if tempCards[i] == 4 {
			GangNum++
			tempCards[i] -= 4
		}
	}
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[6]; i++ {
		if tempCards[i] > 0 && tempCards[i+1] > 0 && tempCards[i+2] > 0 {
			tempCards[i]--
			tempCards[i+1]--
			tempCards[i+2]--
			i--
			ShunNum++
		}
	}

	ud.WanPengNumber = PengNum
	ud.WanGangNumber = GangNum
	ud.WanChiNumber = ShunNum
	ud.WANNumber = TotalNum
}

func (ud *UserData) GetHanCardZINumber(cards [majiangcom.MaxCardValue]int) {
	PengNum := 0
	GangNum := 0
	TotalNum := 0

	tempCards := cards

	for i := majiangcom.Zi[0]; i <= majiangcom.Bai[0]; i++ {
		if cards[i] > 0 {
			TotalNum += cards[i]
		}
		if tempCards[i] == 3 {
			PengNum++
			tempCards[i] -= 3
		} else if tempCards[i] == 4 {
			GangNum++
			tempCards[i] -= 4
		}
	}
	ud.ZiPengNumber = PengNum
	ud.ZiGangNumber = GangNum
	ud.ZINumber = TotalNum
}

func (ud *UserData) SetOutCard(CardValue int32) {
	ud.OutCards[ud.OutCardsIndex] = CardValue
	ud.OutCardsIndex++
}

func (ud *UserData) SubOutCard() {
	ud.OutCardsIndex--
	ud.OutCards[ud.OutCardsIndex] = 0
}

func (ud *UserData) IsChangeTing(AnGangCard [4]int, BaGang [4]int, Gang int) {
	//听牌的时候检查杠牌问题
	if !ud.TingPai ||
		(ud.Opt&(majiangcom.OptTypeAnGang|majiangcom.OptTypeGang|majiangcom.OptTypeMingGang) == 0) ||
		ud.TingPaiMsg == nil {
		return
	}

	opt := 0
	for i := 0; i < 4; i++ {
		if AnGangCard[i] != 0 {
			temp := ud.HandCards
			temp[AnGangCard[i]] -= 4
			if ud.IsHuChange(temp) {
				opt |= majiangcom.OptTypeAnGang
			}
		}

		if BaGang[i] != 0 {
			temp := ud.HandCards
			temp[BaGang[i]] -= 1
			if ud.IsHuChange(temp) {
				opt |= majiangcom.OptTypeMingGang
			}
		}
	}

	if Gang != 0 {
		temp := ud.HandCards
		temp[Gang] -= 3
		if ud.IsHuChange(temp) {
			opt |= majiangcom.OptTypeGang
		}
	}

	ud.Opt &= ^opt
}

func (ud *UserData) IsHuChange(HandCards [majiangcom.MaxCardValue]int) bool {
	tempcount := 0
	for j := 0; j < 16; j++ {
		temp := HandCards
		ret, _ := majiangcom.CanHu(temp, int32(Cards[j]))
		if ret != 0 {
			bFind := false
			for _, v := range ud.TingPaiMsg.Cards[0].HuCardValue {
				if v == int32(Cards[j]) {
					bFind = true
				}
			}

			if !bFind {
				ud.Opt &= ^(majiangcom.OptTypeAnGang | majiangcom.OptTypeGang | majiangcom.OptTypeMingGang)
				return true
			}

			tempcount++
		}
	}

	if tempcount != len(ud.TingPaiMsg.Cards[0].HuCardValue) {
		return true
	}

	return false
}
