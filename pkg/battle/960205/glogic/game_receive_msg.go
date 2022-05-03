package glogic

import (
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960205/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// ReceiveMsgRobBtnEnd 接收用户按下抢庄按钮消息
func (game *ErBaGangGame) ReceiveMsgRobBtnEnd(buffer []byte, user player.PlayerInterface) {
	log.Tracef("接收用户按下抢庄按钮消息")
	game.BtnCount++
	if game.State != Game_Zhuang {
		return
	}
	reqMsg := &msg.RobZhuangEndResp{}
	if err := proto.Unmarshal(buffer, reqMsg); err != nil {
		log.Errorf("proto unmarshal bet request fail: %v", err)
		return
	}
	log.Tracef("抢庄按钮索引是:%v", reqMsg.BtnIndex)
	log.Tracef("游戏局数:%d", game.GameCount)
	var (
		robZhuangMultiple int64 = 0
	)
	//game.RobZhuangMultipleList
	// 找出用户用户所引以及按下按钮的抢庄倍数
	for i := 0; i < len(SendUserAllList); i++ {
		// 找到按下按钮的用户
		if user.GetChairID() == i {
			// 找出按钮索引的值
			switch reqMsg.BtnIndex {
			case 0:
				robZhuangMultiple = 0 // 不抢庄
			case 1:
				robZhuangMultiple = int64(game.RobZhuangConfList[i][0])
			case 2:
				robZhuangMultiple = int64(game.RobZhuangConfList[i][1])
			case 3:
				robZhuangMultiple = int64(game.RobZhuangConfList[i][2])
			case 4:
				robZhuangMultiple = int64(game.RobZhuangConfList[i][3])
			default:
				robZhuangMultiple = 0 // 不抢庄
			}
			// 存下用户的抢庄倍数
			game.RobZhuangMultipleList[i] = robZhuangMultiple
			log.Tracef("id为:%d的用户抢庄倍数是:%v", user.GetChairID(), robZhuangMultiple)
			game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2COneRobZhuangEnd), &msg.OneRobZhuangEndRes{
				UserIndex: int32(i),
				Multiple:  robZhuangMultiple,
			})
			break
		}
	}
	game.checkZhuangEnd()
}

func (game *ErBaGangGame) checkZhuangEnd() {
	for _, v := range game.RobZhuangMultipleList {
		if v == -1 {
			return
		}
	}
	if game.checkState(Game_Zhuang) {
		game.changeState()
		game.SendMsgRobEnd()
	}
}

// ReceiveMsgBetBtnEnd 用户按下下注按钮
func (game *ErBaGangGame) ReceiveMsgBetBtnEnd(buffer []byte, user player.PlayerInterface) {
	log.Tracef("用户%s按下下注按钮", user.GetNike())
	if game.State != Game_Bet {
		return
	}
	game.BtnCount++
	reqMsg := &msg.UserBetEndResp{}
	if err := proto.Unmarshal(buffer, reqMsg); err != nil {
		log.Errorf("proto unmarshal bet request fail: %v", err)
		return
	}
	var (
		betMultiple int64 = 0
	)
	for k := range game.UserAllList {
		// 如果不是庄家就发送消息
		if k != game.RobZhuangIndex {
			if k == user.GetChairID() {
				switch reqMsg.BtnIndex {
				case 0:
					betMultiple = 1
				case 1:
					betMultiple = int64(game.BetConfList[k][0])
				case 2:
					betMultiple = int64(game.BetConfList[k][1])
				case 3:
					betMultiple = int64(game.BetConfList[k][2])
				case 4:
					betMultiple = int64(game.BetConfList[k][3])
				default:
					betMultiple = 1
				}
				if betMultiple == 0 {
					betMultiple = 1
				}
				game.BetMultipleList[k] = betMultiple
				log.Tracef("用户下注倍数:%v", betMultiple)
				game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2CUserBetInfoEnd), &msg.UserBetInfoEndRes{
					UserIndex: int32(k),
					Multiple:  betMultiple,
				})
				break
			}
		}

	}
	game.checkBetEnd()

}

func (game *ErBaGangGame) checkBetEnd() {
	for k, v := range game.BetMultipleList {
		if k != game.RobZhuangIndex && v == -1 {
			return
		}
	}
	if game.checkState(Game_Bet) {
		game.changeState()
		game.SendMsgBetEnd()
	}
}

func (game *ErBaGangGame) deposit(buffer []byte, userInter player.PlayerInterface) {
	user := game.UserAllList[userInter.GetChairID()]
	if user != nil {
		user.IsDeposit = false
	}
}

func (game *ErBaGangGame) test(buffer []byte, userInter player.PlayerInterface) {
	return
	// req := &msg.TestResp{}
	// proto.Unmarshal(buffer, req)
	// fmt.Println(req)
	// if len(req.GetCards()) != 40 {
	// 	return
	// }
	// cards := make([]byte, 0)
	// for _, v := range req.GetCards() {
	// 	cards = append(cards, byte(v))
	// }
	// fmt.Println(cards)
	// game.GamePoker.Cards = cards
}
