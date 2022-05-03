package game

import (
	"game_poker/doudizhu/msg"
	"testing"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
)

func TestDouDizhu_GameStart(t *testing.T) {
	putCardsReq := &msg.PutCardsReq{
		Cards: []byte{51},
	}
	buffer, err := proto.Marshal(putCardsReq)
	if err != nil {
		log.Errorf("proto marshal fail : %v", err.Error())
		return
	}

	req := &msg.PutCardsReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析出牌入参错误: %v", err.Error())
		return
	}

	t.Log(putCardsReq.Cards[0])
	t.Log(req.Cards[0])

	//t.Log(putCardsReq.Cards[0] == req.Cards[0])
}

func TestRobot_Init(t *testing.T) {
	var a [][]int
	a = append(a, []int{})
	t.Log(a)
	t.Log(len(a))
}
