package global

//一些时间、金额常量
const (
	START_GAME_TIME               = 3    //游戏倒计时开始时间
	USER_ACTION_TIME              = 10   //用户考虑的时间
	GAME_TOTAL_ROUND              = 20   //总轮数
	GAME_RESTART_TIME             = 4    //重开时间
	FOLLOW_ALL_THE_WAY_DELAY_TIME = 2    //跟到底动画延时时间
	COMPARE_CARDS_DELAY           = 2000 //比完牌到结束有个延时 毫秒

	TABLE_MIN_ACTION_AMOUNT = 100       //最小的入场（下注）金额
	PROFIT_RATE             = 5         //抽佣比率，1表示1%，0.1表示 0.1%
	MAX_AMOUNT              = 300000000 //牌桌最大注限制
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

//游戏事件信号
const (
	GAME_SIGNAL_START = "start_game"
	GAME_SIGNAL_END   = "end_game"

	HORSE_LAMP_JH = 1
	HORSE_LAMP_SJ = 2
	HORSE_LAMP_BZ = 3
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
	TABLE_CUR_STATUS_WAIT            = 6 //准备开赛
	TABLE_CUR_STATUS_MATCHING        = 1 //匹配玩家
	TABLE_CUR_STATUS_START_SEND_CARD = 2 //客户端开始发牌

	TABLE_CUR_STATUS_ING            = 3 //正在游戏
	TABLE_CUR_STATUS_SYSTEM_COMPARE = 4 //系统比牌
	TABLE_CUR_STATUS_END            = 5 //游戏结束
)

const (
	//1:投注 2:结算
	SET_SCORE_BET    = 1 //下注
	SET_SCORE_SETTLE = 2 //结算
)

const (
	LOSE_REASON_GIVE_UP        = 1
	LOSE_REASON_SYSTEM_COMPARE = 2
)

//游戏的一些配置
var RaiseAmount = []int32{200, 500, 1000}

var ProfitTotal int64 = 0 //总共的抽水

const (
	CardTypeBZ     = 8 //豹子
	CardTypeSJ     = 7 //顺金
	CardTypeSJ123  = 6 //顺金123
	CardTypeJH     = 5 //金花
	CardTypeSZ     = 4 //顺子
	CardTypeSZA23  = 3 //顺子A23
	CardTypeDZ     = 2 //对子
	CardTypeSingle = 1 //单张
)
