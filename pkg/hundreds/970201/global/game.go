package global

//一些时间、金额常量
const (
	WAN_RATE_TOTAL  = 10000
	RED_TOTAL_ROUTE = 15
)

//玩家当前状态
const (
	USER_CUR_STATUS_HALL        = iota //玩家登陆进来，在大厅
	USER_CUR_STATUS_WAIT_START         //进入房间，坐下，准备开赛1
	USER_CUR_STATUS_ING                //正在游戏2
	USER_CUR_STATUS_LOSE               //比牌输了3
	USER_CUR_STATUS_GIVE_UP            //弃牌4
	USER_CUR_STATUS_LOOK               //旁观5
	USER_CUR_STATUS_FINISH_GAME        //该局比赛结束6
)

//牌桌当前状态
const (
	TABLE_CUR_STATUS_WAIT            = 0 //准备开赛
	TABLE_CUR_STATUS_WAIT_SEND_CARDS = 1 //等待客户端发牌完毕
	TABLE_CUR_STATUS_ING             = 2 //正在游戏
	TABLE_CUR_STATUS_END             = 3 //游戏结束
)

const (
	//1:投注 2:结算
	SET_SCORE_BET    = 1 //下注
	SET_SCORE_SETTLE = 2 //结算
)

//红包状态
const (
	RED_CUR_STATUS_ING     = 1 //正在抢
	RED_CUR_STATUS_OVER    = 2 //已抢完
	RED_CUR_STATUS_WAITING = 3 //排队中
)

const (
	KICKOUT_SCORE_NOT_ENOUGH = 1 //余额不足
	KICKOUT_TIME_OUT         = 2 //时间超过3分钟没抢没发

)
