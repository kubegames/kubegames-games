package def

// 算番牌型
const (
	// 一番
	HuaCard   = 0x1
	ZiMo      = 0x2
	DanDiao   = 0x4
	KanZhuang = 0x8
	BianZhang = 0x10
	MingGang  = 0x20
	YaoJiuKe  = 0x40
	LaoShaoFu = 0x80
	Lian6     = 0x100
	YiBanGao  = 0x200
	// 二番
	TingPai      = 0x400
	DuanYao      = 0x800
	AnGang       = 0x1000
	SiGuiYi      = 0x2000
	PingHu       = 0x4000
	MengQianQing = 0x8000
	JianKe       = 0x10000
	ShuangAnKe   = 0x20000
	// 四番
	HeJueZhang     = 0x40000
	ShuangMingGang = 0x80000
	BuQiuRen       = 0x100000
	QuanDaiYao     = 0x200000
	// 六番
	ShuangJianKe = 0x400000
	QuanQiuRen   = 0x800000
	HunYiSe      = 0x1000000
	PengPengHe   = 0x2000000
	// 八番
	ShuangAnGang    = 0x4000000
	QiangGangHu     = 0x8000000
	GangShangKaiHua = 0x10000000
	HaiDiLaoYue     = 0x20000000
	MiaoShouHuiChun = 0x40000000
	// 十二番
	XiaoYuWu  = 0x80000000
	DaYuWu    = 0x100000000
	SanFengKe = 0x200000000
	// 十六番
	TianTing     = 0x400000000
	SanAnKe      = 0x800000000
	YiSeSanBuGao = 0x1000000000
	QingLong     = 0x2000000000
	// 二十四番
	YiSeSanJieGao   = 0x4000000000
	YiSeSanTongShun = 0x8000000000
	QingYiSe        = 0x10000000000
	QiDui           = 0x20000000000
	// 三十二番
	DiHu        = 0x40000000000
	HunYaoJiu   = 0x80000000000
	SanGang     = 0x100000000000
	YiSeSiBuGao = 0x200000000000
	// 四十八番
	TianHu         = 0x400000000000
	YiSeSiJieGao   = 0x800000000000
	YiSeSiTongShun = 0x1000000000000
	// 六十四番
	YiSeShuangLongHun = 0x2000000000000
	SiAnKe            = 0x4000000000000
	ZiYiSe            = 0x8000000000000
	XiaoSanYuan       = 0x10000000000000
	XiaoSiXi          = 0x20000000000000
	// 八十八番
	LianQiDui      = 0x40000000000000
	SiGang         = 0x80000000000000
	JiuBaoLianDeng = 0x100000000000000
	DaSanYuan      = 0x200000000000000
	DaSiXi         = 0x400000000000000
)

// 算番牌型
var FanTypeArray = []int64{0x1, 0x2, 0x4, 0x8, 0x10, 0x20, 0x40, 0x80, 0x100, 0x200, 0x400, 0x800, 0x1000, 0x2000, 0x4000, 0x8000, 0x10000, 0x20000, 0x40000, 0x80000, 0x100000, 0x200000, 0x400000, 0x800000, 0x1000000, 0x2000000, 0x4000000, 0x8000000, 0x10000000, 0x20000000, 0x40000000, 0x80000000, 0x100000000, 0x200000000, 0x400000000, 0x800000000, 0x1000000000, 0x2000000000, 0x4000000000, 0x8000000000, 0x10000000000, 0x20000000000, 0x40000000000, 0x80000000000, 0x100000000000, 0x200000000000, 0x400000000000, 0x800000000000, 0x1000000000000, 0x2000000000000, 0x4000000000000, 0x8000000000000, 0x10000000000000, 0x20000000000000, 0x40000000000000, 0x80000000000000, 0x100000000000000, 0x200000000000000, 0x400000000000000}
var FanDoubleArray = []int32{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 4, 4, 4, 4, 6, 6, 6, 6, 8, 8, 8, 8, 8, 12, 12, 12, 16, 16, 16, 16, 24, 24, 24, 24, 32, 32, 32, 32, 48, 48, 48, 64, 64, 64, 64, 64, 88, 88, 88, 88, 88}
var FanNameArray = []string{"花牌", "自摸", "单钓将", "坎张", "边张", "明杠", "幺九刻", "老少副", "连6", "一般高", "听牌", "断幺", "暗杠", "四归一", "平和", "门前清", "箭刻", "双暗刻", "和绝张", "双明杠", "不求人", "全带幺", "双箭刻", "全求人", "混一色", "碰碰胡", "双暗杠", "抢杠胡", "杠上开花", "海底捞月", "妙手回春", "小于5", "大于5", "三风刻", "天听", "三暗刻", "一色三步高", "清龙", "一色三节高", "一色三同顺", "清一色", "七对", "地胡", "混幺九", "三杠", "一色四步高", "天胡", "一色四节高", "一色四同顺", "一色双龙会", "四暗刻", "字一色", "小三元", "小四喜", "连七对", "四杠", "九宝莲灯", "大三元", "大四喜"}

// 初始牌型Maps

//var StartCardTypeMaps = map[string][14]int{
//	"dasixi":            majiangcom.CreateDaSiXi(),
//	"dasanyuan":         majiangcom.CreateDaSanYuan(),
//	"jiubaoliandeng":    majiangcom.CreateJiuBaoLianDeng(),
//	"lianqidui":         majiangcom.CreateLianQiDui(),
//	"xiaosixi":          majiangcom.CreateXiaoSiXi(),
//	"xiaosanyuan":       majiangcom.CreateXiaoSanYuan(),
//	"ziyise":            majiangcom.CreateZiYiSe(),
//	"sianke":            majiangcom.CreateSiAnKe(),
//	"yiseshuanglonghui": majiangcom.CreateYiSeShuangLongHui(),
//	"yisesitongshun":    majiangcom.CreateYiSeSiTongShun(),
//	"yisesibugao":       majiangcom.CreateYiSeSiBuGao(),
//	"hunyaojiu":         majiangcom.CreateHunYaoJiu(),
//	"qidui":             majiangcom.CreateQiDui(),
//	"qingyise":          majiangcom.CreateQingYiSe(),
//	"yisesantongshun":   majiangcom.CreateYiSeSanTongShun(),
//}
