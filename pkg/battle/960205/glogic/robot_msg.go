package glogic

import (
	"math/rand"

	"github.com/kubegames/kubegames-games/pkg/battle/960205/msg"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
)

var (
	bytBuffer []byte
)

func (r *Robot) ReqRobZhuangEnd(buffer []byte) {
	bytBuffer = buffer
	r.SendMsgRobZhuangEnd()
}

func (r *Robot) SendMsgRobZhuangEnd() {
	reqMsg := &msg.RobZhuangEndResp{}
	if err := proto.Unmarshal(bytBuffer, reqMsg); err != nil {
		log.Errorf("proto unmarshal bet request fail: %v", err)
		return
	}

	log.Tracef("机器人:%v开始抢庄", r.User.GetID())
	btnIndex := int64(rand.Intn(5)) // 随机五个按钮的索引
	log.Tracef("机器人:%v,按下:%d按钮", r.User.GetID(), btnIndex)
	t := rand.Intn(4) + 2
	r.User.AddTimer(int64(t*1000), func() {
		if err := r.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SRobZhuangEnd), &msg.RobZhuangEndResp{
			BtnIndex: btnIndex,
		}); err != nil {
			log.Errorf("机器人抢庄出错:%s", err.Error())
		}
	})

}

func (r *Robot) ReqUserBetEnd(buffer []byte) {
	bytBuffer = buffer
	r.SnedMsgUserBetEnd()
}

func (r *Robot) SnedMsgUserBetEnd() {
	req := &msg.UserBetEndResp{}
	if err := proto.Unmarshal(bytBuffer, req); err != nil {
		log.Errorf("proto unmarshal bet request fail: %v", err)
		return
	}
	var (
		btnM int64 = 0
	)
	for k, v := range r.GameLogic.UserAllList {
		// 通过用户id来判断是哪个用户
		if v.InterUser.GetID() == r.User.GetID() && v.InterUser.GetChairID() != r.GameLogic.RobZhuangIndex {
			btnIndex := rand.Intn(5)
			switch btnIndex {
			case 0:
				btnM = 1
			case 1:
				btnM = int64(r.GameLogic.BetConfList[k][0])
			case 2:
				btnM = int64(r.GameLogic.BetConfList[k][1])
			case 3:
				btnM = int64(r.GameLogic.BetConfList[k][2])
			case 4:
				btnM = int64(r.GameLogic.BetConfList[k][3])
			default:
				btnM = 1 // 不抢庄
			}
			log.Tracef("机器人:%v,下注:%v倍", r.User.GetID(), btnM)
			t := rand.Intn(5000)
			r.User.AddTimer(int64(t), func() {
				_ = r.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SUserBetEnd), &msg.UserBetEndResp{
					//UserIndex: int32(k),
					//Multiple:  btnM, // 找到此用户按下下注按钮索引的值
					BtnIndex: int64(btnIndex),
				})
			})
		}
	}
}
