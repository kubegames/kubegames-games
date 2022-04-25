package global

//一些时间、金额常量
const (
	START_GAME_TIME               = 3  //游戏倒计时开始时间
	USER_ACTION_TIME              = 10 //用户考虑的时间
	GAME_TOTAL_ROUND              = 20 //总轮数
	GAME_RESTART_TIME             = 4  //重开时间
	FOLLOW_ALL_THE_WAY_DELAY_TIME = 2  //跟到底动画延时时间
	COMPARE_CARDS_DELAY           = 4  //比完牌到结束有个延时

	TABLE_MIN_ACTION_AMOUNT = 100       //最小的入场（下注）金额
	PROFIT_RATE             = 5         //抽佣比率，1表示1%，0.1表示 0.1%
	MAX_AMOUNT              = 300000000 //牌桌最大注限制

	WAN_RATE_TOTAL = 10000

	TICKER_TIME_READY_ROB = 5  //倒计时10s发包，准备开抢
	TICKER_TIME_START_ROB = 10 //10s的抢包时间
	TICKER_TIME_END_ROB   = 5  //5s的清场时间
)

//玩家操作的选项
const (
	USER_OPTION_GIVE_UP                   = "give_up"
	USER_OPTION_COMPARE                   = "compare"
	USER_OPTION_RAISE                     = "raise"
	USER_OPTION_SEE_CARDS                 = "see_cards"
	USER_OPTION_ALL_IN                    = "all_in"
	USER_OPTION_FOLLOW                    = "follow"             //跟注
	USER_OPTION_FOLLOW_ALL_THE_WAY        = "follow_all_the_way" //跟到底
	USER_OPTION_CANCEL_FOLLOW_ALL_THE_WAY = "cancel_follow"
)

//定时器相关事件
const (
	TRIGGER_EVENT_GAME_START = "start_game" //开始游戏
	TRIGGER_EVENT_ACTION     = "action"     //玩家发言
)

const (
	//1:投注 2:结算
	SET_SCORE_BET    = 1 //下注
	SET_SCORE_SETTLE = 2 //结算
)

//游戏事件信号
const (
	GAME_SIGNAL_START = "start_game"
	GAME_SIGNAL_END   = "end_game"
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
	TABLE_CUR_STATUS_READY_ROB = 1 // 此时正在进行发包前的倒计时
	TABLE_CUR_STATUS_START_ROB = 2 //发包倒计时结束，开始抢包，一共有10s时间抢包
	TABLE_CUR_STATUS_END_ROB   = 3 //抢包结束，清场
)

//红包状态
const (
	RED_CUR_STATUS_READY = 0 //红包发送出来准备抢
	RED_CUR_STATUS_ING   = 1 //正在抢
	RED_CUR_STATUS_OVER  = 2 //已抢完
)

const (
	EVENT_SIG_READY_RED = 0 //红包即将出现倒计时
	EVENT_SIG_START_ROB = 1 //红包出现，开始抢包倒计时
	EVENT_SIG_ROB_END   = 2 //抢包结束
)
