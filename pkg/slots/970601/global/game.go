package global

//一些时间、金额常量
const (
	WAN_RATE_TOTAL  = 10000
	RED_TOTAL_ROUTE = 15
)

//牌桌当前状态
const (
	TABLE_CUR_STATUS_WAIT            = 0 //准备开赛
	TABLE_CUR_STATUS_WAIT_SEND_CARDS = 1 //等待客户端发牌完毕
	TABLE_CUR_STATUS_ING             = 2 //正在游戏
	TABLE_CUR_STATUS_END             = 3 //游戏结束
)

const (
	SCORE_KIND_PEACE = 0 // 积分变动类型-和
	SCORE_KIND_BET   = 1 // 积分变动类型-下注
	SCORE_KIND_WIN   = 2 // 积分变动类型-赢
	SCORE_KIND_BACK  = 2 // 积分变动类型-退还
)

const (
	USER_STATUS_WAIT = 0
	USER_STATUS_ING  = 1
	USER_STATUS_END  = 2
)

const (
	COMPARE_LOSE = -1
	COMPARE_EQ   = 0
	COMPARE_WIN  = 1
)

const (
	TOTAL_BOX_COUNT = 15
)

const (
	//1:投注 2:结算
	SET_SCORE_BET    = 1 //下注
	SET_SCORE_SETTLE = 2 //结算
)


const (
	//这儿本来byte就可以，但是想着rand随机函数方便，就定义了int32
	//第一关
	ICON_BAIYU = 1 //白玉
	ICON_BIYU  = 2 //碧玉
	ICON_MOYU  = 3 //墨玉
	ICON_MANAO = 4 //玛瑙
	ICON_HUPO  = 5 //琥珀
	//第二关
	ICON_ZUMULV     = 11 //祖母绿
	ICON_MAOYANSHI  = 12 //猫眼石
	ICON_ZISHUIJING = 13 //紫水晶
	ICON_FEICUISHI  = 14 //翡翠石
	ICON_BAIZHENZHU = 15 //白珍珠
	//第三关
	ICON_HONGBAOSHI  = 21 //红宝石
	ICON_LVBAOSHI    = 22 //绿宝石
	ICON_HUANGBAOSHI = 23 //黄宝石
	ICON_LANBAOSHI   = 24 //蓝宝石
	ICON_BAIZUANSHI  = 25 //白钻石

	//钥匙
	ICON_KEY = 100
)
