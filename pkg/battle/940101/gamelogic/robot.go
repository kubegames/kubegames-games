package gamelogic

import (
	"strconv"

	"github.com/golang/protobuf/proto"
	majiangcom "github.com/kubegames/kubegames-games/internal/pkg/majiang"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/940101/config"
	errenmajiang "github.com/kubegames/kubegames-games/pkg/battle/940101/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

const (
	allchi  = 0xE  //左吃又吃中吃
	allgang = 0xE0 //摸杠明杠
)

type Robot struct {
	User         player.RobotInterface
	TableCards   [majiangcom.MaxCardValue]int //牌桌上的牌包括其余玩家的杠吃碰牌
	HandCards    [majiangcom.MaxCardValue]int //机器人手牌
	TingSatusNum int                          //听牌人数
	GameLogic    *Game                        // 游戏逻辑
	TimerJob     *player.Job
	OutCard      int   //最后一张操作牌值
	TingStatus   bool  //trul 听牌
	CurrTingNum  int   //当前听牌数量
	ChairId      int32 //机器人椅子ID
	baoTing      bool  //报听，报听后不操作
	TingQiDui    bool  //听七对

}

func (r *Robot) Init(User player.RobotInterface, g table.TableHandler) {
	r.User = User
	r.GameLogic = g.(*Game)
	for i := 0; i < UserNum; i++ {
		if r.GameLogic.ud[i].User.GetID() == User.GetID() {
			r.ChairId = int32(r.GameLogic.ud[i].User.GetChairID())
			break
		}
	}
}

func (r *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case int32(errenmajiang.ReMsgIDS2C_Opt):
		r.HandCards = r.GameLogic.ud[r.ChairId].HandCards
		//r.isSelfOp = false
		r.OutCard = r.GameLogic.LastOutCard
		r.Robotop(buffer)
		break
	case int32(errenmajiang.ReMsgIDS2C_BroadOptRet):
		r.OpResult(buffer)
		break
	//case int32(errenmajiang.ReMsgIDS2C_MoPai):
	//	msg := &errenmajiang.UserMoPaiMessage{}
	//	proto.Unmarshal(buffer, msg)
	//	log.Traceln("机器人摸牌值",msg)
	//	break
	case int32(errenmajiang.ReMsgIDS2C_BroadTing):
		msg := &errenmajiang.TingCardMessage{}
		proto.Unmarshal(buffer, msg)
		if msg.TingUserIndex == r.ChairId {
			r.baoTing = true
			//log.Traceln("------机器人报听报听成功------", msg)
			return
		} else {
			r.TingSatusNum = 1
			return
		}
	case int32(errenmajiang.ReMsgIDS2C_SettleMsg):
		r.ReSetData()
	}
}

func (r *Robot) ReSetData() {
	r.TableCards = [majiangcom.MaxCardValue]int{}
	r.HandCards = [majiangcom.MaxCardValue]int{}
	r.TingSatusNum = 0
	r.OutCard = 0
	r.TingStatus = false
	r.CurrTingNum = 0
	r.baoTing = false
	r.TingQiDui = false
}

//补花后手牌
func (r *Robot) RobotHandcards(b []byte) {
	msg := &errenmajiang.BuHuaMsg{}
	err := proto.Unmarshal(b, msg)
	if err != nil {
		log.Errorf("机器人解析当前操作玩家信息失败: %v", err)
		return
	}
	if msg.ChairId != r.ChairId {
		return
	}
	//for _, v := range msg.HanCards {
	//	r.HandCards[v]++
	//}
	for _, val := range msg.BuHuaDatas {
		if val < int32(majiangcom.Hua[0]) {
			r.HandCards[val]++
		}
	}
}

//记录摸牌值
func (r *Robot) RecordMoPai(b []byte) {
	msg := &errenmajiang.UserMoPaiMessage{}
	err := proto.Unmarshal(b, msg)
	if err != nil {
		log.Errorf("机器人解析当前操作玩家信息失败: %v", err)
		return
	}
	if msg.CardValues == 0 {
		return
	} else {
		if msg.CardValues > 37 {
			return
		} else {
			r.HandCards[msg.CardValues] += 1
			//r.MoPai = int(msg.CardValues)
		}
	}
}

//机器人操作
func (r *Robot) Robotop(b []byte) {
	msg := &errenmajiang.OptMsg{}
	err := proto.Unmarshal(b, msg)
	if err != nil {
		log.Errorf("机器人解析当前操作玩家信息失败: %v", err)
		return
	}

	chi := 0x0
	peng := 0x0
	gang := 0x0

	//如果操作有胡直接胡
	if int32(r.GameLogic.ud[r.ChairId].Opt)&msg.OptType != 0 && msg.OptType&majiangcom.OptTypeHu != 0 {
		r.SendOpMsg(majiangcom.OptTypeHu)
		//log.Traceln("服务器机器人胡牌后手牌", majiangcom.GetHandCards(r.GameLogic.ud[1].HandCards))
		return
	}
	//如果是报听的话只操作胡的消息
	if r.baoTing {
		if msg.OptType&0xE0 != 0 {
			if msg.OptType&(majiangcom.OptTypeMingGang|majiangcom.OptTypeAnGang) != 0 {
				r.findHandCardsMoGangCard(msg.OptType)
				if !r.tingStatusGangOp(msg.OptType) {
					r.SendOpMsg(0)
					return
				}
				return
			} else {
				if !r.tingStatusGangOp(majiangcom.OptTypeGang) {
					r.SendOpMsg(0)
					return
				}
				return
			}
		}
		return
	}
	//出牌消息
	if msg.OptType == majiangcom.OptTypeOutCard {
		if r.TingStatus {
			r.tingStatusOutCardOp()
			return
		} else {
			r.outCardOp()
			return
		}
	}

	//除了胡以外，剩以下6中情况
	//获取是否有杠操作
	if msg.OptType&allgang != 0 {
		gang = 0x4
	}
	//获取是否有碰操作
	if msg.OptType&majiangcom.OptTypePeng != 0 {
		peng = 0x2
	}
	//获取是否有吃操作
	if msg.OptType&allchi != 0 {
		chi = 0x1
	}
	//获取除胡意外的6中操作情况temp，分别做权值计算并得出最佳操作方式并记录操作。
	temp := gang | peng | chi
	switch temp {
	//有杠有碰有吃
	case 0x7:
		//log.Traceln("======0x7=====")
		if r.TingStatus == true {
			if !r.tingStatusGangOp(majiangcom.OptTypeGang) {
				r.tingStatusPenOp()
			}
			break
		} else {
			opt := r.CPGop(msg.OptType, temp)
			r.SendOpMsg(opt)
			break
		}
	//有杠有碰无吃
	case 0x6:
		//log.Traceln("======0x6=====")
		if r.TingStatus == true {
			if !r.tingStatusGangOp(majiangcom.OptTypeGang) {
				r.tingStatusPenOp()
			}
			break
		} else {
			opt := r.CPGop(msg.OptType, temp)
			r.SendOpMsg(opt)
			break
		}
		//有杠无碰无吃
	case 0x4:
		//log.Traceln("======0x4=====")
		//摸杠检测哪张牌是杠牌
		r.findHandCardsMoGangCard(msg.OptType)
		if r.TingStatus == true {
			if !r.tingStatusGangOp(msg.OptType) {
				r.SendOpMsg(0)
				break
			}
			break
		} else {
			opt := r.moGangOp(msg.OptType)
			r.SendOpMsg(opt)
			break
		}
	//无杠有碰有吃
	case 0x3:
		//log.Traceln("======0x3=====")
		if r.TingStatus == true {
			r.tingStatusPenOp()
			break
		} else {
			opt := r.CPGop(msg.OptType, temp)
			r.SendOpMsg(opt)
			break
		}
		//无杠有碰无吃
	case 0x2:
		//log.Traceln("======0x2=====")
		if r.TingStatus == true {
			r.tingStatusPenOp()
			break
		} else {
			a, _ := r.pengOp()
			if a == true {
				r.SendOpMsg(majiangcom.OptTypePeng)
				break
			} else {
				r.SendOpMsg(0)
				break
			}
		}
		//无杠无碰有吃
	case 0x1:
		//log.Traceln("======0x1=====")
		if r.TingStatus == true {
			r.SendOpMsg(0)
			break
		} else {
			opt := r.CPGop(msg.OptType, temp)
			r.SendOpMsg(opt)
			break
		}
	}
}

//碰杠吃过操作结果记录牌桌上的牌，手牌
func (r *Robot) OpResult(b []byte) {
	msg := &errenmajiang.BroadOptRetMsg{}
	err := proto.Unmarshal(b, msg)
	if err != nil {
		log.Errorf("机器人解析当前操作玩家信息失败: %v", err)
		return
	}
	switch msg.Opt {
	case majiangcom.OptTypeZuoChi:
		r.TableCards[msg.CardValue-1] += 1
		r.TableCards[msg.CardValue-2] += 1
		//if r.isSelfOp {
		//	r.HandCards[msg.CardValue-1] -= 1
		//	r.HandCards[msg.CardValue-2] -= 1
		//	r.isSelfOp = false
		//}
	case majiangcom.OptTypeZhongChi:
		r.TableCards[msg.CardValue+1] += 1
		r.TableCards[msg.CardValue-1] += 1
		//if r.isSelfOp {
		//	r.HandCards[msg.CardValue+1] -= 1
		//	r.HandCards[msg.CardValue-1] -= 1
		//	r.isSelfOp = false
		//}
	case majiangcom.OptTypeYouChi:
		r.TableCards[msg.CardValue+1] += 1
		r.TableCards[msg.CardValue+2] += 1
		//if r.isSelfOp {
		//	r.HandCards[msg.CardValue+1] -= 1
		//	r.HandCards[msg.CardValue+2] -= 1
		//	r.isSelfOp = false
		//}
	case majiangcom.OptTypePeng:
		r.TableCards[msg.CardValue] += 2
		//if r.isSelfOp {
		//	r.HandCards[msg.CardValue] -= 2
		//	r.pengCunt += 1
		//	r.isSelfOp = false
		//}
	case majiangcom.OptTypeMingGang:
		r.TableCards[msg.CardValue] = 4
		//if r.isSelfOp {
		//	r.HandCards[msg.CardValue] = 0
		//	r.isSelfOp = false
		//
		//}
	case majiangcom.OptTypeGang:
		r.TableCards[msg.CardValue] = 4
		//if r.isSelfOp {
		//	r.HandCards[msg.CardValue] = 0
		//	r.isSelfOp = false
		//}
	case majiangcom.OptTypeAnGang:
		r.TableCards[msg.CardValue] = 4
		//if r.isSelfOp {
		//	r.HandCards[msg.CardValue] = 0
		//	r.isSelfOp = false
		//}
	case majiangcom.OptTypeOutCard:
		r.TableCards[msg.CardValue] += 1

	}
}

//吃碰杠过操作对比
func (r *Robot) CPGop(OptType int32, temp int) (Opt int32) {
	zuoChiWeight := 0.0
	zhongChiWeight := 0.0
	youChiWeight := 0.0
	gangWeight := 0.0
	pengWeight := 0.0
	switch {
	case temp&0x4 != 0:
		_, gangWeight = r.mingGangOp()
	case temp&0x2 != 0:
		_, pengWeight = r.pengOp()
	case temp&0x1 != 0:

		if OptType&majiangcom.OptTypeZuoChi != 0 {
			_, zuoChiWeight = r.chiOp(majiangcom.OptTypeZuoChi)
		}
		if OptType&majiangcom.OptTypeZhongChi != 0 {
			_, zhongChiWeight = r.chiOp(majiangcom.OptTypeZhongChi)
		}
		if OptType&majiangcom.OptTypeYouChi != 0 {
			_, youChiWeight = r.chiOp(majiangcom.OptTypeYouChi)
		}
	}
	max := gangWeight
	Opt = majiangcom.OptTypeGang
	if pengWeight > max {
		max = pengWeight
		Opt = majiangcom.OptTypePeng
	}
	if zuoChiWeight > max {
		max = zuoChiWeight
		Opt = majiangcom.OptTypeZuoChi
	}
	if zhongChiWeight > max {
		max = zhongChiWeight
		Opt = majiangcom.OptTypeZhongChi
	}
	if youChiWeight > max {
		max = youChiWeight
		Opt = majiangcom.OptTypeYouChi
	}
	if max == 0.0 {
		Opt = 0
	}
	return Opt
}

//对家打出的杠牌
func (r *Robot) mingGangOp() (bool, float64) {
	WardCardsNum := r.GameLogic.MaJiang.GetLastCardsCount()
	var opCard int
	opLaterHandCards := r.HandCards
	opLaterHandCards[r.OutCard] -= 3
	opCard = r.OutCard
	isOp, weight := majiangcom.MingGangCardOP(r.HandCards, opLaterHandCards, opCard, WardCardsNum, r.TingSatusNum)

	return isOp, weight
}

//摸牌杠
func (r *Robot) moGangOp(OptType int32) (Opt int32) {
	WardCardsNum := r.GameLogic.MaJiang.GetLastCardsCount()
	if OptType&majiangcom.OptTypeAnGang != 0 {
		Opt = majiangcom.OptTypeAnGang
	} else {
		Opt = majiangcom.OptTypeMingGang
	}
	var opCard int
	opLaterHandCards := r.HandCards
	opLaterHandCards[r.OutCard] = 0
	opCard = r.OutCard
	isOp := majiangcom.MoGangCardOP(r.HandCards, opLaterHandCards, opCard, WardCardsNum, r.TingSatusNum)

	if isOp == true {
		return Opt
	} else {
		return 0
	}

}

func (r *Robot) pengOp() (bool, float64) {
	var opCard int
	opLaterHandCards := r.HandCards
	opLaterHandCards[r.OutCard] -= 2
	opCard = r.OutCard
	isOp, weight := majiangcom.PengCardOP(r.HandCards, opLaterHandCards, opCard)
	return isOp, weight
}

func (r *Robot) chiOp(chiType int) (bool, float64) {
	var opCard int
	var x, y int

	if chiType == majiangcom.OptTypeZuoChi {
		x = -1
		y = -2
	} else if chiType == majiangcom.OptTypeZhongChi {
		x = -1
		y = 1
	} else {
		x = +2
		y = +1
	}
	opLaterHandCards := r.HandCards
	opLaterHandCards[r.OutCard+x] -= 1
	opLaterHandCards[r.OutCard+y] -= 1
	opCard = r.OutCard
	isOp, weight := majiangcom.EatCardOP(r.HandCards, opLaterHandCards, opCard)
	return isOp, weight
}

//听牌，打出的牌和值 输入14张手牌，返回打出的牌值 和 听牌剩余牌的数量
func (r *Robot) TingMaxNumCardAndNum(handCards [majiangcom.MaxCardValue]int) (bool, int, int) {
	// temp :=r.HandCards
	temp := handCards
	max := 0 //可胡牌最大数量
	var outCardTing int
	count := 0
	for k, v := range temp {
		if v >= 1 {
			temp[k] -= 1
			isTing, canTingCount := r.tingCheck(temp)
			if isTing == true {
				outCardTing = k
				count++

				if canTingCount > max {
					max = canTingCount
					outCardTing = k
				}
			}
			//else {
			//	outCardTing = majiangcom.PlayCardOP(handCards, r.TableCards)
			//}
			temp[k] += 1
		}
	}
	if count == 0 {
		return false, outCardTing, 0
	}

	return true, outCardTing, max
}

func (r *Robot) TingMaxNumCardAndNum1(handCards [majiangcom.MaxCardValue]int) map[int][]int {
	tingMaps := map[int][]int{}
	for _, val := range majiangcom.GetHandCards(handCards) {
		tempCard1 := handCards
		tempCard1[val] -= 1
		var HuCards []int
		tingFlag := false
		for i := majiangcom.Wan[0]; i <= majiangcom.Bai[0]; i++ {
			if opt, _ := majiangcom.CanHu(tempCard1, int32(i)); opt&majiangcom.OptTypeHu != 0 {
				tingFlag = true
				HuCards = append(HuCards, i)
			}
		}
		if tingFlag && len(HuCards) > 0 {
			tingMaps[int(val)] = HuCards
		}
	}
	return tingMaps
}

//可听牌牌值和数量。输入打出牌后的13张手牌
func (r *Robot) tingCheck(handCards [majiangcom.MaxCardValue]int) (bool, int) {
	var canHuPaiArr [14]int
	count := 0
	canTingCount := 0
	for i := 1; i <= 9; i++ {
		if opt, _ := majiangcom.CanHu(handCards, int32(i)); opt == majiangcom.OptTypeHu {
			canHuPaiArr[count] = i
			count++
		}

	}
	for i := 31; i <= 37; i++ {
		if opt, _ := majiangcom.CanHu(handCards, int32(i)); opt == majiangcom.OptTypeHu {
			canHuPaiArr[count] = i
			count++
		}
	}

	if count == 0 {
		return false, canTingCount
	}

	for i := 0; i < count; i++ {
		canTingCount += r.TableCards[canHuPaiArr[i]]
		canTingCount += r.HandCards[canHuPaiArr[i]]
		canTingCount += r.GameLogic.ud[0].HandCards[canHuPaiArr[i]]
	}
	//log.Traceln(canTingCount, count)
	//log.Traceln(canHuPaiArr)
	canTingCount = 4*count - canTingCount
	return true, canTingCount
}

//碰杠吃过胡操作消息
func (r *Robot) SendOpMsg(opType int32) {
	frq := &errenmajiang.OptRetMsg{}
	frq.Opt = opType
	frq.CardValue = int32(r.OutCard)

	//if opType == 0 {
	//	log.Warnf(" -----------放弃操作-------")
	//}
	r.TimerJob, _ = r.User.AddTimer(int64(r.robotOpTime(opType)), func() {
		if err := r.User.SendMsgToServer(int32(errenmajiang.MsgIDC2S_OptRet), frq); err != nil {
			log.Errorf("send server msg fail: %v", err.Error())
		}
	})
}

func (r *Robot) GetGangCardValue(opType int32) int {
	switch 0xE0 & opType {
	case majiangcom.OptTypeMingGang:
		for _, val := range r.GameLogic.ud[r.ChairId].PengCards {
			if r.GameLogic.ud[r.ChairId].HandCards[val] > 0 {
				return val
			}
		}
	case majiangcom.OptTypeGang:
		for val, num := range r.GameLogic.ud[r.ChairId].HandCards {
			if num == 3 {
				return val
			}
		}
	case majiangcom.OptTypeAnGang:
		for val, num := range r.GameLogic.ud[r.ChairId].HandCards {
			if num == 4 {
				return val
			}
		}
	}
	return 0
}

//出牌消息
func (r *Robot) SendPlayCardOpMsg(card int) {
	frq := &errenmajiang.OptRetMsg{}
	frq.CardValue = int32(card)
	frq.Opt = majiangcom.OptTypeOutCard
	if frq.CardValue == 0 {
		log.Warnf("机器人出牌值等于0检查程序   %v", frq.CardValue)
	}
	r.TimerJob, _ = r.User.AddTimer(int64(r.robotOpTime(majiangcom.OptTypeOutCard)), func() {
		err := r.User.SendMsgToServer(int32(errenmajiang.MsgIDC2S_OptRet), frq)
		if err != nil {
			log.Errorf("send server msg fail: %v", err.Error())
		}
	})
}
func (r *Robot) SendTingOpMsg(card int) {
	frq := &errenmajiang.TingCardMessage{}
	frq.TingCardValue = int32(card)
	frq.TingUserIndex = 1
	if frq.TingCardValue == 0 {
		log.Warnf("报听出牌值为0，检查程序")
	}

	rand.Rand4()
	r.TimerJob, _ = r.User.AddTimer(int64(rand.Int31n(config.ErRenMaJiang.OutCardTime)), func() {
		err := r.User.SendMsgToServer(int32(errenmajiang.MsgIDC2S_TingCard), frq)
		if err != nil {
			log.Errorf("send server msg fail: %v", err.Error())
		}
	})
}

//检测是否可以听牌，输入14张手牌，可听，返回true 牌值，剩余牌数量 不可听 返回false 0，0
func (r *Robot) IsCanTing() (bool, int, int) {
	tmp := r.HandCards
	//tmp:= ha
	ting := false
	duicount := 0
	//开始的位置
	startindex := 0
	//结束的位置记录开始和结束可以减少循环次数
	endindex := 0
	//最多7对将牌
	jiangcount := 0

	var t [2]int
	ct := 0
	for i := 1; i <= majiangcom.Bai[0]; i++ {
		if tmp[i] >= 1 {
			if tmp[i] >= 2 {
				jiangcount++
				if tmp[i] == 4 {
					duicount++
				}
			}
			if startindex == 0 {
				startindex = i
			}

			if tmp[i] == 1 || tmp[i] == 3 {
				if ct <= 1 {
					t[ct] = i
				}
				ct++
			}
			endindex = i
		}
	}
	//6对听牌
	if (jiangcount + duicount) == 6 {
		r.TingQiDui = true
		//log.Traceln("------听七对----")
		//fmt.Print(tmp)
		a := r.TableCards[t[0]]
		a += r.HandCards[t[0]]
		a += r.GameLogic.ud[0].HandCards[t[0]]
		b := r.TableCards[t[1]]
		b += r.HandCards[t[1]]
		b += r.GameLogic.ud[0].HandCards[t[1]]
		if a < b {
			return true, t[0], 4 - a
		} else {
			return true, t[1], 4 - b
		}

	}
	tmpfu := tmp
	ziSingleNum := 0
	//去除字刻
	for i := majiangcom.Zi[0]; i <= endindex; i++ {
		if tmpfu[i] == 0 || tmpfu[i] == 3 {
			tmpfu[i] = 0
			continue
		}
		if tmpfu[i] == 1 {
			ziSingleNum++
		}
	}
	//log.Traceln(ziSingleNum)
	//如果单张字牌数量超过3则不符合听牌情况
	if ziSingleNum >= 3 {
		return false, 0, 0
	}
	//找出链子刻子并从数组种剔除
	tempa := tmpfu
	count := 0 //下次循环的起始下标
	for i := startindex; i < 11; i++ {
		b := 0
		resdiuCard := 0 // 剩余手牌统计
		duict := 0      //对子统计
		singlect := 0   //单牌统计
		lianCt := 0     //相隔相邻统计
		for k := startindex; k < 11; k++ {
			a := k + count
			if k+count == 11-1 {
				b = startindex
				a = b
			}
			if k+count > 11-1 {
				b += 1
				a = b
			}
			if tmpfu[a] == 0 || tmpfu[a] == 3 {
				tmpfu[a] = 0
				continue
			}
			for j := 0; j <= tmpfu[a]; j++ {

				if tmpfu[a+1] == 0 || tmpfu[a+2] == 0 {
					continue
				}
				tmpfu[a] -= 1
				tmpfu[a+1] -= 1
				tmpfu[a+2] -= 1
			}
		}
		//log.Traceln(tmpfu)
		//统计剩余的牌数量，对子数量，单牌数量无相隔相隔，相邻相隔牌数量。向前对比
		for k, v := range tmpfu {
			//如果是大于最后一张牌结束循环
			if k > endindex {
				break
			}
			if v == 0 {
				continue
			}
			if v >= 1 {
				resdiuCard += v
			}
			if resdiuCard > 5 {
				break
			}
			if v == 2 {
				duict++
			}
			//剩余牌中=1的牌牌<29如果+1+2有值为连拍，
			if v == 1 {
				if k < 30 {
					if tmpfu[k+1] == 1 || tmpfu[k+2] == 1 {
						lianCt++
					} else {
						singlect++
					}
				} else {
					singlect++
				}
			}
		}
		//log.Traceln(resdiuCard,singlect)
		if resdiuCard > 5 {
			count++
			tmpfu = tempa
			continue
		}
		if resdiuCard == 2 {
			_, outCardTing, num := r.TingMaxNumCardAndNum(tmp)
			return true, outCardTing, num
		}
		if resdiuCard == 5 {
			if singlect > 2 {
				count++
				tmpfu = tempa
				continue
			}
			if duict == 2 {
				_, outCardTing, num := r.TingMaxNumCardAndNum(tmp)
				return true, outCardTing, num
			}
			if duict == 1 {
				_, outCardTing, num := r.TingMaxNumCardAndNum(tmp)
				return true, outCardTing, num
			}
		}

	}
	return ting, 0, 0
}

func (r *Robot) tingStatusGangOp(opType int32) bool {
	tem := r.HandCards
	tem[r.OutCard] = 0
	a, _ := r.tingCheck(tem)
	if a == true {
		r.SendOpMsg(opType)
		return true
	} else {
		return false
	}
}

func (r *Robot) tingStatusPenOp() {
	tem := r.HandCards
	tem[r.OutCard] -= 2
	canTing, _, num := r.TingMaxNumCardAndNum(tem)
	if canTing == true {
		if r.CurrTingNum < num {
			r.SendOpMsg(majiangcom.OptTypePeng)
			return
		} else {
			r.SendOpMsg(0)
			return
		}
	} else {
		r.SendOpMsg(0)
		return
	}
}

func (r *Robot) findHandCardsMoGangCard(OptType int32) {
	if OptType&majiangcom.OptTypeMingGang != 0 && r.GameLogic.ud[r.ChairId].PengPaiNumber > 0 {
		for _, val := range r.GameLogic.ud[r.ChairId].PengCards {
			if r.GameLogic.ud[r.ChairId].HandCards[val] == 1 {
				r.OutCard = val
				return
			}
		}
	} else if OptType&majiangcom.OptTypeAnGang != 0 {
		for k, v := range r.GameLogic.ud[r.ChairId].HandCards {
			if v == 4 {
				r.OutCard = k
				return
			}
		}
	}
}

func (r *Robot) robotOpTime(opt int32) int64 {
	randTime := rand.RandInt(1, 101)
	tempConfig := config.ErRenMaJiang.OutCardTimeMaps
	if opt&majiangcom.OptTypeOutCard == 0 {
		tempConfig = config.ErRenMaJiang.OptCardTimeMaps
	}
	resTime := int64(2)
	for key, val := range tempConfig {
		keyInt, _ := strconv.Atoi(key)
		if randTime <= keyInt {
			resTime = rand.Int63n(val)
			break
		}
	}
	return resTime + 1000
}

//听牌状态出牌操作
func (r *Robot) tingStatusOutCardOp() {
	//baoTingNum := 4
	//if r.TingQiDui {
	//	baoTingNum = 2
	//}
	//_, cardValue, num := r.TingMaxNumCardAndNum(r.HandCards)
	tingMaps := r.TingMaxNumCardAndNum1(r.GameLogic.ud[1].HandCards)
	if len(tingMaps) > 0 {
		tingLen := 0
		tingOutCard := 0
		for val, vals := range tingMaps {
			if len(vals) > tingLen {
				tingLen = len(vals)
				tingOutCard = val
			}
		}
		r.SendTingOpMsg(tingOutCard)
		//r.SendPlayCardOpMsg(cardValue)
		return
	}
	//else {
	//	log.Traceln("机器听牌状态出牌", cardValue)
	//	r.SendPlayCardOpMsg(cardValue)
	//	r.CurrTingNum = num
	//	return
	//}
}

//非听牌状态出牌操作
func (r *Robot) outCardOp() {
	//检测是否听牌
	baoTingNum := 4
	canTing, cardValue, num := r.IsCanTing()
	if canTing == true {
		if r.TingQiDui {
			baoTingNum = 2
		}
		if num >= baoTingNum {
			r.SendTingOpMsg(cardValue)
			//r.SendPlayCardOpMsg(cardValue)
			return
		} else {
			r.SendPlayCardOpMsg(cardValue)
			r.TingStatus = true
			r.CurrTingNum = num
			return
		}
		//无听普通出牌
	} else {
		card := majiangcom.PlayCardOP(r.HandCards, r.TableCards)
		r.SendPlayCardOpMsg(card)
		return
	}
}
