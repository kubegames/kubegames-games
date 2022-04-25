package global

//框架和逻辑消息协议
const (
	FRAME_MSG_PROTOCOL = 1
	LOGIC_MSG_PROTOCOL = 2
)

//客户端发送给服务器的消息
const (
	C2S_SEND_RED        = 1 //发红包
	C2S_LOCK_RED        = 2 //锁定红包
	C2S_CANCEL_LOCK_RED = 3 //取消锁定红包
	C2S_ROB_RED         = 4 //抢红包
	C2S_GET_SENT_RED    = 5 //获取发送过的红包列表
	C2S_GET_ROBBED_INFO = 6 //获取抢过的红包信息列表
	C2S_GET_USER_LIST   = 7 //获取房间内用户列表
	C2S_GET_MINE_RECORD = 8 //中雷记录
	C2S_GET_HALL_RECORD = 9 //获取大厅战绩
	C2S_GET_USER_COUNT = 10 //获取玩家人数

)

//服务器发送给客户端的消息
const (
	S2C_INTO_ROOM         = 0  //进入房间
	S2C_OTHER_INTO_ROOM   = 1  //其他人进入房间
	S2C_SEND_RED          = 2  //有人发送红包
	S2C_LOCK_RED          = 3  //锁定红包
	S2C_CANCEL_LOCK_RED   = 4  //取消锁定红包
	S2C_ROB_RED           = 5  //抢红包
	S2C_LEAVE_TABLE       = 6  //离开房间
	S2C_GET_SENT_RED      = 7  //获取发送过的红包列表
	S2C_GET_ROBBED_INFO   = 8  //获取抢过的红包信息列表
	S2C_GET_USER_LIST     = 9  //获取房间内用户列表
	S2C_GET_MINE_RECORD   = 10 //中雷记录
	S2C_RED_INFO          = 11 //红包信息
	S2C_RED_DISAPPEAR     = 12 //红包消失
	S2C_RED_ROBBED_COUNT  = 13 //红包当前已被抢个数
	S2C_GET_HALL_RECORD   = 14 //获取大厅战绩
	S2C_RED_WAIT_SEND     = 15 //红包等待发送出去
	S2C_SEND_RED_SUCCESS  = 16 //红包发送成功
	S2C_USER_SCORE        = 17 //玩家金额
	S2C_CHECK_OVERDUE_RED = 18 //定期检查屏幕上存在已久的红包
	S2C_USER_COUNT = 19//返回玩家人数

)

const (
	ERROR_CODE_OK              = 20000 //没有错误消息，主要用于程序内部判断
	ERROR_CODE_NOT_INTOROOM    = 10000 //玩家还未进入房间或没连上线
	ERROR_CODE_NOT_ENOUGH      = 10001 //玩家资金不足
	ERROR_CODE_RED_OVER        = 10002 //红包已被抢完
	ERROR_CODE_RED_MINENUM     = 10003 //红包雷号不能大于9
	ERROR_ROB_NOT_ENOUGH       = 10004 //抢包金额不足
	ERROR_CODE_CANT_LEAVE   = 10008 //不能离开，用户还有红包在场上
	ERROR_CODE_CAN_NOT_ALL_IN  = 10009 //大于等于3局才能all in
	ERROR_CODE_LESS_MIN_ACTION = 10010 //小于最低入场金额
	ERROR_CODE_CAN_NOT_RAISE   = 10011 //已经有人allin不能加注
	ERROR_CODE_LEVEL_NOT_FIND  = 10012 //该等级的房间不存在
)
