package game

import (
	"common/rand"
	"fmt"
	"game_buyu/crazy_red/config"
)

//3月5号新增机器人发包
//func (game *Game) AiSendNew() {
//	redCount := game.redList.Len()
//	aiSend := config.GetAiSendNew(redCount)
//	if aiSend == nil {
//		return
//	}
//	sendRedCount := aiSend.RedCount[GetSendRedIndex(aiSend)]
//	fmt.Println("发送红包数量：",sendRedCount)
//
//	//策略：每秒找两个机器人发两个包，如果最后发的数量不足，再申请剩余数量的机器人来发剩下的所有包
//	for i := aiSend.Interval[0]; i <= aiSend.Interval[len(aiSend.Interval)-1]; i += 1000 {
//		fmt.Println("第",i,"毫秒准备发包")
//		if sendRedCount <= 0 {
//			return
//		}
//		ais := game.randGetAis(2)
//		for _,ai := range ais {
//			game.ProcSendRed()
//		}
//	}
//}

//获取发送红包的对应概率下标
func GetSendRedIndex(aiSend *config.AiSendNew) int {
	sendCountTotalRate := 0
	for _,v := range aiSend.SendCountRate {
		sendCountTotalRate += v
	}
	index := rand.RandInt(0,sendCountTotalRate)
	fmt.Println("发送红包数量的index：",index)
	tmpSendCountRate := 0
	for i:=0;i<len(aiSend.SendCountRate);i++ {
		if index >= tmpSendCountRate && index < tmpSendCountRate+aiSend.SendCountRate[i] {
			return i
		}
		tmpSendCountRate += aiSend.SendCountRate[i]
	}
	return -1
}