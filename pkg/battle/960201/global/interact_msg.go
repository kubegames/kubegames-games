package global

//框架和逻辑消息协议
const (
	FRAME_MSG_PROTOCOL = 1
	LOGIC_MSG_PROTOCOL = 2
)

//客户端发送给服务器的消息
const (
	C2S_START_GAME           = 0 //玩家点击开始游戏
	C2S_USER_ACTION          = 1 //玩家发言
	C2S_COMPARE_CARDS        = 2 //比牌
	C2S_LEAVE_TABLE          = 3 //离开牌桌
	C2S_RESTART_GAME         = 4 //重开比赛
	C2S_GET_CAN_COMPARE_LIST = 5 //获取可比牌的用户列表
	C2S_SENDCARD_OVER        = 6 //客户端发牌动画结束，通知服务器开始倒计时
	C2S_GET_ROOM_LIST        = 7 //获取房间列表
	C2S_IS_CAN_INTO_ROOM     = 8 //是否有资格进入房间
	C2S_SET_CARD_TYPE        = 9 //为玩家设置指定牌型
	C2S_SEE_OTHTER_CARDS        = 10 //
)

//服务器发送给客户端的消息
const (
	S2C_INTO_ROOM              = 0  //进入房间
	S2C_OTHER_INTO_ROOM        = 1  //其他人进入房间
	S2C_WAIT_START             = 2  //当前房间只有1人，等待开赛
	S2C_TICKER_START           = 3  //房间有2人，进入开赛倒计时
	S2C_SEND_CARDS_OK          = 4  //发牌完毕，玩家开始发言
	S2C_USER_ACTION            = 5  //玩家发言
	S2C_COMPARE_CARDS          = 6  //比牌
	S2C_LEAVE_TABLE            = 7  //离开牌桌
	S2C_CUR_ACTION_USER        = 8  //当前发言玩家
	S2C_FINISH_GAME            = 9  //比赛结束
	S2C_USER_SEE_CARDS         = 10 //玩家看牌
	S2C_FINISH_CUR_ROUND       = 11 //结束当前轮，返回牌桌信息等
	S2C_TABLE_INFO             = 12 //返回给客户端当钱牌桌信息
	S2C_START_SEND_CARDS       = 13 //开始发牌
	S2C_GET_CAN_COMPARE_LIST   = 14 //获取可比牌的用户列表
	S2C_GIVE_UP_CARDS_FOR_SELF = 15 //弃牌给自己看
	S2C_GET_ROOM_LIST          = 16 //获取房间列表
	S2C_PUB_COMPARE            = 17 //发起比牌
	S2C_KEEP_MATCH             = 18 //房间内开始匹配之后有玩家离开，房间只剩一个人，继续匹配
	S2C_SEE_OTHER_CARDS             = 19 //
)

const (
	ERROR_CODE_OK                   = 20000 //没有错误消息，主要用于程序内部判断
	ERROR_CODE_NOT_INTOROOM         = 10000 //玩家还未进入房间或没连上线
	ERROR_CODE_NOT_ENOUGH           = 10001 //玩家资金不足
	ERROR_CODE_LESS_MIN             = 10002 //玩家下注金额小于当前最小金额
	ERROR_CODE_NIL                  = 10003 //非法错误,一些空指针异常等
	ERROR_CODE_CANNOT_LEAVE         = 10004 //非法离开牌桌
	ERROR_CODE_ARG                  = 10005 //参数错误，比如输入负数、或者金额 <= 0等情况
	ERROR_CODE_GAME_NOT_ING         = 10006 //游戏没在进行中
	ERROR_CODE_NOT_CUR_USER_ACTION  = 10007 //还没到该玩家发言
	ERROR_CODE_USER_NOT_FOUND       = 10008 //没找到该玩家
	ERROR_CODE_CAN_NOT_ALL_IN       = 10009 //大于等于3局才能all in
	ERROR_CODE_LESS_MIN_ACTION      = 10010 //小于最低入场金额
	ERROR_CODE_CAN_NOT_RAISE        = 10011 //已经有人allin不能加注
	ERROR_CODE_OVER_USER_MIN_AMOUNT = 10012 //加注大于了玩家的金额
	ERROR_CODE_COMPARE_NOT_ENOUGH   = 10013 //钱不够不能比牌

)
