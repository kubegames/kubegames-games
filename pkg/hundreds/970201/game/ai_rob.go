package game

import (
	"common/rand"
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/msg"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

//3月5号新增机器人抢包 5s轮询一次
func (game *Game) AiRobNew() {
	go func() {
		redCount := game.redList.Len()
		aiRob := config.GetAiRobNew(redCount)
		if aiRob == nil {
			log.Traceln("aiRob config 为空，红包数量：", redCount)
			return
		}
		robCount := aiRob.RobCount[GetRobRedIndex(aiRob)]
		ais := game.randGetAis(aiRob.AiCount)
		for i := aiRob.Interval[0]; i <= aiRob.Interval[len(aiRob.Interval)-1]; i += 1000 {
			for _, ai := range ais {
				for e := game.redList.Front(); e != nil; e = e.Next() {
					red := e.Value.(*Red)
					if red.RobbedCount >= red.RedFlood {
						continue
					}
					game.aiRob(ai, red)
					time.Sleep(300 * time.Millisecond)
					game.aiRob(ai, red)
					time.Sleep(300 * time.Millisecond)
					game.aiRob(ai, red)
					robCount -= 3
					if robCount <= 0 {
						return
					}
				}
			}
			time.Sleep(time.Second)
		}
	}()

}

func (game *Game) aiRob(ai *data.User, red *Red) {
	//log.Traceln("机器人抢包 aiRob")
	c2sMsg := &msg.C2SRobRed{RedId: red.Id}
	c2sMsgB, _ := proto.Marshal(c2sMsg)
	game.ProcRobRed(c2sMsgB, ai)
}

//获取发送红包的对应概率下标
func GetRobRedIndex(aiRob *config.AiRobNew) int {
	sendCountTotalRate := 0
	for _, v := range aiRob.RobCountRate {
		sendCountTotalRate += v
	}
	index := rand.RandInt(0, sendCountTotalRate)
	//log.Traceln("发送红包数量的概率随机值：",index)
	tmpSendCountRate := 0
	for i := 0; i < len(aiRob.RobCountRate); i++ {
		//log.Traceln("tmpSendCountRate  +aiSend.SendCountRate[i] ",tmpSendCountRate,tmpSendCountRate+aiSend.SendCountRate[i])
		if index >= tmpSendCountRate && index < tmpSendCountRate+aiRob.RobCountRate[i] {
			return i
		}
		tmpSendCountRate += aiRob.RobCountRate[i]
	}
	return -1
}
