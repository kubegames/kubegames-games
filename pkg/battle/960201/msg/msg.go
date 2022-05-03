package msg

import (
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

//msg包放置与客户端交互的所有消息结构

//封装发送给客户端协议
func PkgS2CMsg(subCmd int32, buff []byte) []byte {
	msgStruct := &S2CMsg{MainCmd: 2, SubCmd: subCmd, Buff: buff}

	b, err := proto.Marshal(msgStruct)
	if err != nil {
		log.Traceln("PkgS2CMsg err : ", err)
	}
	return b
}
