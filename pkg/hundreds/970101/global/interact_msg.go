package global

//框架和逻辑消息协议
const (
	FRAME_MSG_PROTOCOL = 1
	LOGIC_MSG_PROTOCOL = 2
)

//客户端发送给服务器的消息
const (
	C2S_SEND_RED         = 1 //发红包
	C2S_ROB_RED          = 2 //抢红包
	C2S_GET_SENT_RED     = 3 //获取发送过的红包列表
	C2S_GET_ROBBED_INFO  = 4 //获取抢过的红包信息列表
	C2S_GET_USER_LIST    = 5 //获取房间内用户列表
	C2S_CANCEL_SEND      = 6 //取消发送红包
	C2S_GET_CUR_RED_LIST = 7 //获取当前红包列表
)

//服务器发送给客户端的消息
const (
	S2C_INTO_ROOM        = 0  //进入房间
	S2C_OTHER_INTO_ROOM  = 1  //其他人进入房间
	S2C_SEND_RED         = 2  //有人发送红包
	S2C_READY_ROB_TICKER = 3  //准备发红包的倒计时
	S2C_START_ROB        = 4  //时间到，开始抢包
	S2C_ROB_RED          = 5  //抢红包
	S2C_LEAVE_TABLE      = 6  //离开房间
	S2C_GET_SENT_RED     = 7  //获取发送过的红包列表
	S2C_GET_ROBBED_INFO  = 8  //获取抢过的红包信息列表
	S2C_GET_USER_LIST    = 9  //获取房间内用户列表
	S2C_RED_DISAPPEAR    = 10 //红包消失
	S2C_KICK_OUT         = 11 // 玩家被踢出
	S2C_CANCEL_SEND      = 12 //玩家取消发送红包
	S2C_GET_CUR_RED_LIST = 13 //获取当前红包列表
	S2C_CUR_USER_COUNT   = 14 //当前玩家数量
)

const (
	ERROR_CODE_OK            = 20000 //没有错误消息，主要用于程序内部判断
	ERROR_CODE_NOT_INTOROOM  = 10000 //玩家还未进入房间或没连上线
	ERROR_CODE_NOT_ENOUGH    = 10001 //玩家资金不足
	ERROR_CODE_RED_OVER      = 10002 //红包已被抢完
	ERROR_CODE_RED_MINENUM   = 10003 //红包雷号不能大于9
	ERROR_CODE_NOT_START     = 10004 //红包还没开始抢
	ERROR_CODE_USER_SENT_RED = 10005 //红包列表中 玩家已经发送过红包了
	ERROR_CODE_USER_ROBBED   = 10006 //玩家已经抢过该红包
	ERROR_CODE_CANT_CANCEL   = 10007 //不能取消发送红包，当前红包在第一个位置了
	ERROR_CODE_CANT_LEAVE    = 10008 //不能离开，用户还有红包在厂商

)
