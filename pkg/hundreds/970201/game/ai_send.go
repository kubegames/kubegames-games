package game

import (
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/msg"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

//3月5号新增机器人发包 5s轮询一次
func (game *Game) AiSendNew() {
	if game.isClosed {
		log.Traceln("房间正在重制，机器人不再发红包")
		return
	}
	go func() {
		redCount := game.redList.Len()
		aiSend := config.GetAiSendNew(redCount)
		if aiSend == nil {
			log.Traceln("aiSend config 为空，红包数量：", redCount)
			return
		}
		redCountIndex := GetSendRedIndex(aiSend)
		//log.Traceln("redCountIndex : ",redCountIndex,aiSend.SendCount)
		sendRedCount := aiSend.SendCount[redCountIndex]
		//log.Traceln("发送红包数量：",sendRedCount)

		//策略：每秒找两个机器人发两个包，如果最后发的数量不足，再申请剩余数量的机器人来发剩下的所有包
		for i := aiSend.Interval[0]; i <= aiSend.Interval[len(aiSend.Interval)-1]; i += 1000 {
			//log.Traceln("第",i,"毫秒准备发包")
			if sendRedCount <= 0 {
				return
			}
			ais := game.randGetAis(2)
			if ais == nil || len(ais) == 0 {
				return
			}
			for _, ai := range ais {
				if sendRedCount <= 0 {
					return
				}
				game.AiSend(ai, aiSend)
				sendRedCount--
			}
			time.Sleep(time.Second)
		}
		//log.Traceln("还剩 ",sendRedCount," 个红包没发完")
		ais := game.randGetAis(sendRedCount)
		for _, ai := range ais {
			if sendRedCount <= 0 {
				return
			}
			game.AiSend(ai, aiSend)
			sendRedCount--
		}
	}()
}

//机器人发包
func (game *Game) AiSend(ai *data.User, aiSend *config.AiSendNew) {
	c2sMsg := &msg.C2SSendRed{
		MineNum: int64(rand.RandInt(0, 9)), Amount: game.GetAiSendAmount(aiSend),
		Count: 10, UserId: ai.Id}
	c2sMsgB, _ := proto.Marshal(c2sMsg)
	game.ProcSendRed(c2sMsgB, ai)
}

//获取机器人发包金额
func (game *Game) GetAiSendAmount(aiSend *config.AiSendNew) int64 {
	sendAmountTotalRate := 0
	for _, v := range aiSend.SendAmountRate {
		sendAmountTotalRate += v
	}
	index := rand.RandInt(0, sendAmountTotalRate)
	//log.Traceln("发送红包金额的随机值：",index)
	tmpSendCountRate := 0
	for i := 0; i < len(aiSend.SendAmountRate); i++ {
		if index >= tmpSendCountRate && index < tmpSendCountRate+aiSend.SendAmountRate[i] {
			//log.Traceln("发送红包金额的index：",i)
			return game.sendAmount[i]
		}
		tmpSendCountRate += aiSend.SendAmountRate[i]
	}
	panic("-1")
	return -1
}

//获取发送红包的对应概率下标
func GetSendRedIndex(aiSend *config.AiSendNew) int {
	sendCountTotalRate := 0
	for _, v := range aiSend.SendCountRate {
		sendCountTotalRate += v
	}
	index := rand.RandInt(0, sendCountTotalRate)
	//log.Traceln("发送红包数量的概率随机值：",index)
	tmpSendCountRate := 0
	for i := 0; i < len(aiSend.SendCountRate); i++ {
		//log.Traceln("tmpSendCountRate  +aiSend.SendCountRate[i] ",tmpSendCountRate,tmpSendCountRate+aiSend.SendCountRate[i])
		if index >= tmpSendCountRate && index < tmpSendCountRate+aiSend.SendCountRate[i] {
			return i
		}
		tmpSendCountRate += aiSend.SendCountRate[i]
	}
	return -1
}
