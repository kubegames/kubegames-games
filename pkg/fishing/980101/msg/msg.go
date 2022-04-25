package msg

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

//msg包放置与客户端交互的所有消息结构

//封装发送给客户端协议
func PkgS2CMsg(subCmd int32, buff []byte) []byte {
	msgStruct := &Msg{MainCmd: 2, SubCmd: subCmd, Buff: buff}

	b, err := proto.Marshal(msgStruct)
	if err != nil {
		fmt.Println("PkgS2CMsg err : ", err)
	}
	return b
}
