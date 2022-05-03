package config

import (
	"encoding/json"
	"fmt"
	"go-game-sdk/example/game_LaBa/970601/global"
	"io/ioutil"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/sipt/GoJsoner"
	"github.com/tidwall/gjson"
)

type config struct {
	Cheat  int32 //web socket 侦听端口
	Count  int   //场次
	Bottom int64 //数据库链接字符串
	Tax    int64 //需要制定输出的日志路径
}

var Config = new(config)

type gameFrameConfig struct {
	DbIp   string `json:"db_ip"`
	DbPwd  string `json:"db_pwd"`
	DbUser string `json:"db_user"`
	DbName string `json:"db_name"`
}

var GameFrameConfig = new(gameFrameConfig)

//连环夺宝作弊率
type IconRate struct {
	MaxTimes int64
	//最大消除次数
	MaxDisCount int
	//钻头出现概率
	KeyRate int
	//进入彩金池概率
	IntoCaijin int
	//彩金额度
	Caijin1 int
	Caijin2 int
	Caijin3 int
	Caijin4 int
	Caijin5 int
	//第一关
	BaiyuRateA int
	BiyuRateA  int
	MoyuRateA  int
	ManaoRateA int
	HupoRateA  int
	//第二关
	ZumulvRateA     int
	MaoyanshiRateA  int
	ZishuijingRateA int
	FeicuishiRateA  int
	BaizhenzhuRateA int
	//第三关
	HongbaoshiRateA  int
	LvbaoshiRateA    int
	HuangbaoshiRateA int
	LanbaoshiRateA   int
	BaizuanshiRateA  int
	//第一关
	BaiyuRateB int
	BiyuRateB  int
	MoyuRateB  int
	ManaoRateB int
	HupoRateB  int
	//第二关
	ZumulvRateB     int
	MaoyanshiRateB  int
	ZishuijingRateB int
	FeicuishiRateB  int
	BaizhenzhuRateB int
	//第三关
	HongbaoshiRateB  int
	LvbaoshiRateB    int
	HuangbaoshiRateB int
	LanbaoshiRateB   int
	BaizuanshiRateB  int
}

var IconRateArr = make([]*IconRate, 0)

type IconConfigStruct struct {
	Baiyu []int64
	Biyu  []int64
	Moyu  []int64
	Manao []int64
	Hupo  []int64

	Zumulv     []int64
	Maoyanshi  []int64
	Zishuijing []int64
	Feicuishi  []int64
	Baizhenzhu []int64

	Hongbaoshi  []int64
	Lvbaoshi    []int64
	Huangbaoshi []int64
	Lanbaoshi   []int64
	Baizuanshi  []int64

	IconIndex1  []int64
	IconIndex2  []int64
	IconIndex3  []int64
	Caijin      []int64
	CaijinScore []int64
}

var IconConfig = new(IconConfigStruct)
var IconScoreMap = make(map[int32]map[int32]int64) //种类、数量=>得分
func InitIconScoreMap() {
	log.Traceln("InitIconScoreMap CaijinScore: ", IconConfig.CaijinScore)
	//第一关 白玉
	IconScoreMap[global.ICON_BAIYU] = make(map[int32]int64)
	IconScoreMap[global.ICON_BAIYU][4] = IconConfig.Baiyu[0]
	IconScoreMap[global.ICON_BAIYU][5] = IconConfig.Baiyu[1]
	IconScoreMap[global.ICON_BAIYU][6] = IconConfig.Baiyu[2]
	IconScoreMap[global.ICON_BAIYU][7] = IconConfig.Baiyu[3]
	IconScoreMap[global.ICON_BAIYU][8] = IconConfig.Baiyu[4]
	IconScoreMap[global.ICON_BAIYU][9] = IconConfig.Baiyu[5]
	IconScoreMap[global.ICON_BAIYU][10] = IconConfig.Baiyu[6]
	IconScoreMap[global.ICON_BAIYU][11] = IconConfig.Baiyu[7]
	IconScoreMap[global.ICON_BAIYU][12] = IconConfig.Baiyu[8]
	IconScoreMap[global.ICON_BAIYU][13] = IconConfig.Baiyu[9]
	IconScoreMap[global.ICON_BAIYU][14] = IconConfig.Baiyu[10]
	IconScoreMap[global.ICON_BAIYU][15] = IconConfig.Baiyu[10]
	IconScoreMap[global.ICON_BAIYU][16] = IconConfig.Baiyu[10]
	//第一关 碧玉
	IconScoreMap[global.ICON_BIYU] = make(map[int32]int64)
	IconScoreMap[global.ICON_BIYU][4] = IconConfig.Biyu[0]
	IconScoreMap[global.ICON_BIYU][5] = IconConfig.Biyu[1]
	IconScoreMap[global.ICON_BIYU][6] = IconConfig.Biyu[2]
	IconScoreMap[global.ICON_BIYU][7] = IconConfig.Biyu[3]
	IconScoreMap[global.ICON_BIYU][8] = IconConfig.Biyu[4]
	IconScoreMap[global.ICON_BIYU][9] = IconConfig.Biyu[5]
	IconScoreMap[global.ICON_BIYU][10] = IconConfig.Biyu[6]
	IconScoreMap[global.ICON_BIYU][11] = IconConfig.Biyu[7]
	IconScoreMap[global.ICON_BIYU][12] = IconConfig.Biyu[8]
	IconScoreMap[global.ICON_BIYU][13] = IconConfig.Biyu[9]
	IconScoreMap[global.ICON_BIYU][14] = IconConfig.Biyu[10]
	IconScoreMap[global.ICON_BIYU][15] = IconConfig.Biyu[10]
	IconScoreMap[global.ICON_BIYU][16] = IconConfig.Biyu[10]
	//第一关 墨玉
	IconScoreMap[global.ICON_MOYU] = make(map[int32]int64)
	IconScoreMap[global.ICON_MOYU][4] = IconConfig.Moyu[0]
	IconScoreMap[global.ICON_MOYU][5] = IconConfig.Moyu[1]
	IconScoreMap[global.ICON_MOYU][6] = IconConfig.Moyu[2]
	IconScoreMap[global.ICON_MOYU][7] = IconConfig.Moyu[3]
	IconScoreMap[global.ICON_MOYU][8] = IconConfig.Moyu[4]
	IconScoreMap[global.ICON_MOYU][9] = IconConfig.Moyu[5]
	IconScoreMap[global.ICON_MOYU][10] = IconConfig.Moyu[6]
	IconScoreMap[global.ICON_MOYU][11] = IconConfig.Moyu[7]
	IconScoreMap[global.ICON_MOYU][12] = IconConfig.Moyu[8]
	IconScoreMap[global.ICON_MOYU][13] = IconConfig.Moyu[9]
	IconScoreMap[global.ICON_MOYU][14] = IconConfig.Moyu[10]
	IconScoreMap[global.ICON_MOYU][15] = IconConfig.Moyu[10]
	IconScoreMap[global.ICON_MOYU][16] = IconConfig.Moyu[10]
	//第一关 玛瑙
	IconScoreMap[global.ICON_MANAO] = make(map[int32]int64)
	IconScoreMap[global.ICON_MANAO][4] = IconConfig.Manao[0]
	IconScoreMap[global.ICON_MANAO][5] = IconConfig.Manao[1]
	IconScoreMap[global.ICON_MANAO][6] = IconConfig.Manao[2]
	IconScoreMap[global.ICON_MANAO][7] = IconConfig.Manao[3]
	IconScoreMap[global.ICON_MANAO][8] = IconConfig.Manao[4]
	IconScoreMap[global.ICON_MANAO][9] = IconConfig.Manao[5]
	IconScoreMap[global.ICON_MANAO][10] = IconConfig.Manao[6]
	IconScoreMap[global.ICON_MANAO][11] = IconConfig.Manao[7]
	IconScoreMap[global.ICON_MANAO][12] = IconConfig.Manao[8]
	IconScoreMap[global.ICON_MANAO][13] = IconConfig.Manao[9]
	IconScoreMap[global.ICON_MANAO][14] = IconConfig.Manao[10]
	IconScoreMap[global.ICON_MANAO][15] = IconConfig.Manao[10]
	IconScoreMap[global.ICON_MANAO][16] = IconConfig.Manao[10]
	//第一关 琥珀
	IconScoreMap[global.ICON_HUPO] = make(map[int32]int64)
	IconScoreMap[global.ICON_HUPO][4] = IconConfig.Hupo[0]
	IconScoreMap[global.ICON_HUPO][5] = IconConfig.Hupo[1]
	IconScoreMap[global.ICON_HUPO][6] = IconConfig.Hupo[2]
	IconScoreMap[global.ICON_HUPO][7] = IconConfig.Hupo[3]
	IconScoreMap[global.ICON_HUPO][8] = IconConfig.Hupo[4]
	IconScoreMap[global.ICON_HUPO][9] = IconConfig.Hupo[5]
	IconScoreMap[global.ICON_HUPO][10] = IconConfig.Hupo[6]
	IconScoreMap[global.ICON_HUPO][11] = IconConfig.Hupo[7]
	IconScoreMap[global.ICON_HUPO][12] = IconConfig.Hupo[8]
	IconScoreMap[global.ICON_HUPO][13] = IconConfig.Hupo[9]
	IconScoreMap[global.ICON_HUPO][14] = IconConfig.Hupo[10]
	IconScoreMap[global.ICON_HUPO][15] = IconConfig.Hupo[10]
	IconScoreMap[global.ICON_HUPO][16] = IconConfig.Hupo[10]

	//第二关 祖母绿
	IconScoreMap[global.ICON_ZUMULV] = make(map[int32]int64)
	IconScoreMap[global.ICON_ZUMULV][5] = IconConfig.Zumulv[0]
	IconScoreMap[global.ICON_ZUMULV][6] = IconConfig.Zumulv[1]
	IconScoreMap[global.ICON_ZUMULV][7] = IconConfig.Zumulv[2]
	IconScoreMap[global.ICON_ZUMULV][8] = IconConfig.Zumulv[3]
	IconScoreMap[global.ICON_ZUMULV][9] = IconConfig.Zumulv[4]
	IconScoreMap[global.ICON_ZUMULV][10] = IconConfig.Zumulv[5]
	IconScoreMap[global.ICON_ZUMULV][11] = IconConfig.Zumulv[6]
	IconScoreMap[global.ICON_ZUMULV][12] = IconConfig.Zumulv[7]
	IconScoreMap[global.ICON_ZUMULV][13] = IconConfig.Zumulv[8]
	IconScoreMap[global.ICON_ZUMULV][14] = IconConfig.Zumulv[9]
	IconScoreMap[global.ICON_ZUMULV][15] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][16] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][17] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][18] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][19] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][20] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][21] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][22] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][23] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][24] = IconConfig.Zumulv[10]
	IconScoreMap[global.ICON_ZUMULV][25] = IconConfig.Zumulv[10]
	//第二关 猫眼石
	IconScoreMap[global.ICON_MAOYANSHI] = make(map[int32]int64)
	IconScoreMap[global.ICON_MAOYANSHI][5] = IconConfig.Maoyanshi[0]
	IconScoreMap[global.ICON_MAOYANSHI][6] = IconConfig.Maoyanshi[1]
	IconScoreMap[global.ICON_MAOYANSHI][7] = IconConfig.Maoyanshi[2]
	IconScoreMap[global.ICON_MAOYANSHI][8] = IconConfig.Maoyanshi[3]
	IconScoreMap[global.ICON_MAOYANSHI][9] = IconConfig.Maoyanshi[4]
	IconScoreMap[global.ICON_MAOYANSHI][10] = IconConfig.Maoyanshi[5]
	IconScoreMap[global.ICON_MAOYANSHI][11] = IconConfig.Maoyanshi[6]
	IconScoreMap[global.ICON_MAOYANSHI][12] = IconConfig.Maoyanshi[7]
	IconScoreMap[global.ICON_MAOYANSHI][13] = IconConfig.Maoyanshi[8]
	IconScoreMap[global.ICON_MAOYANSHI][14] = IconConfig.Maoyanshi[9]
	IconScoreMap[global.ICON_MAOYANSHI][15] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][15] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][16] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][17] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][18] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][19] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][20] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][21] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][22] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][23] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][24] = IconConfig.Maoyanshi[10]
	IconScoreMap[global.ICON_MAOYANSHI][25] = IconConfig.Maoyanshi[10]
	//第二关 紫水晶
	IconScoreMap[global.ICON_ZISHUIJING] = make(map[int32]int64)
	IconScoreMap[global.ICON_ZISHUIJING][5] = IconConfig.Zishuijing[0]
	IconScoreMap[global.ICON_ZISHUIJING][6] = IconConfig.Zishuijing[1]
	IconScoreMap[global.ICON_ZISHUIJING][7] = IconConfig.Zishuijing[2]
	IconScoreMap[global.ICON_ZISHUIJING][8] = IconConfig.Zishuijing[3]
	IconScoreMap[global.ICON_ZISHUIJING][9] = IconConfig.Zishuijing[4]
	IconScoreMap[global.ICON_ZISHUIJING][10] = IconConfig.Zishuijing[5]
	IconScoreMap[global.ICON_ZISHUIJING][11] = IconConfig.Zishuijing[6]
	IconScoreMap[global.ICON_ZISHUIJING][12] = IconConfig.Zishuijing[7]
	IconScoreMap[global.ICON_ZISHUIJING][13] = IconConfig.Zishuijing[8]
	IconScoreMap[global.ICON_ZISHUIJING][14] = IconConfig.Zishuijing[9]
	IconScoreMap[global.ICON_ZISHUIJING][15] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][16] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][17] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][18] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][19] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][20] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][21] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][22] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][23] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][24] = IconConfig.Zishuijing[10]
	IconScoreMap[global.ICON_ZISHUIJING][25] = IconConfig.Zishuijing[10]
	//第二关 翡翠石
	IconScoreMap[global.ICON_FEICUISHI] = make(map[int32]int64)
	IconScoreMap[global.ICON_FEICUISHI][5] = IconConfig.Feicuishi[0]
	IconScoreMap[global.ICON_FEICUISHI][6] = IconConfig.Feicuishi[1]
	IconScoreMap[global.ICON_FEICUISHI][7] = IconConfig.Feicuishi[2]
	IconScoreMap[global.ICON_FEICUISHI][8] = IconConfig.Feicuishi[3]
	IconScoreMap[global.ICON_FEICUISHI][9] = IconConfig.Feicuishi[4]
	IconScoreMap[global.ICON_FEICUISHI][10] = IconConfig.Feicuishi[5]
	IconScoreMap[global.ICON_FEICUISHI][11] = IconConfig.Feicuishi[6]
	IconScoreMap[global.ICON_FEICUISHI][12] = IconConfig.Feicuishi[7]
	IconScoreMap[global.ICON_FEICUISHI][13] = IconConfig.Feicuishi[8]
	IconScoreMap[global.ICON_FEICUISHI][14] = IconConfig.Feicuishi[9]
	IconScoreMap[global.ICON_FEICUISHI][15] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][16] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][17] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][18] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][19] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][20] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][21] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][22] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][23] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][24] = IconConfig.Feicuishi[10]
	IconScoreMap[global.ICON_FEICUISHI][25] = IconConfig.Feicuishi[10]
	//第二关 白珍珠
	IconScoreMap[global.ICON_BAIZHENZHU] = make(map[int32]int64)
	IconScoreMap[global.ICON_BAIZHENZHU][5] = IconConfig.Baizhenzhu[0]
	IconScoreMap[global.ICON_BAIZHENZHU][6] = IconConfig.Baizhenzhu[1]
	IconScoreMap[global.ICON_BAIZHENZHU][7] = IconConfig.Baizhenzhu[2]
	IconScoreMap[global.ICON_BAIZHENZHU][8] = IconConfig.Baizhenzhu[3]
	IconScoreMap[global.ICON_BAIZHENZHU][9] = IconConfig.Baizhenzhu[4]
	IconScoreMap[global.ICON_BAIZHENZHU][10] = IconConfig.Baizhenzhu[5]
	IconScoreMap[global.ICON_BAIZHENZHU][11] = IconConfig.Baizhenzhu[6]
	IconScoreMap[global.ICON_BAIZHENZHU][12] = IconConfig.Baizhenzhu[7]
	IconScoreMap[global.ICON_BAIZHENZHU][13] = IconConfig.Baizhenzhu[8]
	IconScoreMap[global.ICON_BAIZHENZHU][14] = IconConfig.Baizhenzhu[9]
	IconScoreMap[global.ICON_BAIZHENZHU][15] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][16] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][17] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][18] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][19] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][20] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][21] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][22] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][23] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][24] = IconConfig.Baizhenzhu[10]
	IconScoreMap[global.ICON_BAIZHENZHU][25] = IconConfig.Baizhenzhu[10]

	//第三关 红宝石
	IconScoreMap[global.ICON_HONGBAOSHI] = make(map[int32]int64)
	IconScoreMap[global.ICON_HONGBAOSHI][6] = IconConfig.Hongbaoshi[0]
	IconScoreMap[global.ICON_HONGBAOSHI][7] = IconConfig.Hongbaoshi[1]
	IconScoreMap[global.ICON_HONGBAOSHI][8] = IconConfig.Hongbaoshi[2]
	IconScoreMap[global.ICON_HONGBAOSHI][9] = IconConfig.Hongbaoshi[3]
	IconScoreMap[global.ICON_HONGBAOSHI][10] = IconConfig.Hongbaoshi[4]
	IconScoreMap[global.ICON_HONGBAOSHI][11] = IconConfig.Hongbaoshi[5]
	IconScoreMap[global.ICON_HONGBAOSHI][12] = IconConfig.Hongbaoshi[6]
	IconScoreMap[global.ICON_HONGBAOSHI][13] = IconConfig.Hongbaoshi[7]
	IconScoreMap[global.ICON_HONGBAOSHI][14] = IconConfig.Hongbaoshi[8]
	IconScoreMap[global.ICON_HONGBAOSHI][15] = IconConfig.Hongbaoshi[9]
	IconScoreMap[global.ICON_HONGBAOSHI][16] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][17] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][18] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][19] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][20] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][21] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][22] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][23] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][24] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][25] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][26] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][27] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][28] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][29] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][30] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][31] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][32] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][33] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][34] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][35] = IconConfig.Hongbaoshi[10]
	IconScoreMap[global.ICON_HONGBAOSHI][36] = IconConfig.Hongbaoshi[10]

	//第三关 绿宝石
	IconScoreMap[global.ICON_LVBAOSHI] = make(map[int32]int64)
	IconScoreMap[global.ICON_LVBAOSHI][6] = IconConfig.Lvbaoshi[0]
	IconScoreMap[global.ICON_LVBAOSHI][7] = IconConfig.Lvbaoshi[1]
	IconScoreMap[global.ICON_LVBAOSHI][8] = IconConfig.Lvbaoshi[2]
	IconScoreMap[global.ICON_LVBAOSHI][9] = IconConfig.Lvbaoshi[3]
	IconScoreMap[global.ICON_LVBAOSHI][10] = IconConfig.Lvbaoshi[4]
	IconScoreMap[global.ICON_LVBAOSHI][11] = IconConfig.Lvbaoshi[5]
	IconScoreMap[global.ICON_LVBAOSHI][12] = IconConfig.Lvbaoshi[6]
	IconScoreMap[global.ICON_LVBAOSHI][13] = IconConfig.Lvbaoshi[7]
	IconScoreMap[global.ICON_LVBAOSHI][14] = IconConfig.Lvbaoshi[8]
	IconScoreMap[global.ICON_LVBAOSHI][15] = IconConfig.Lvbaoshi[9]
	IconScoreMap[global.ICON_LVBAOSHI][16] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][17] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][18] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][19] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][20] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][21] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][22] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][23] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][24] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][25] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][26] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][27] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][28] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][29] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][30] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][31] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][32] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][33] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][34] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][35] = IconConfig.Lvbaoshi[10]
	IconScoreMap[global.ICON_LVBAOSHI][36] = IconConfig.Lvbaoshi[10]
	//第三关 黄宝石
	IconScoreMap[global.ICON_HUANGBAOSHI] = make(map[int32]int64)
	IconScoreMap[global.ICON_HUANGBAOSHI][6] = IconConfig.Huangbaoshi[0]
	IconScoreMap[global.ICON_HUANGBAOSHI][7] = IconConfig.Huangbaoshi[1]
	IconScoreMap[global.ICON_HUANGBAOSHI][8] = IconConfig.Huangbaoshi[2]
	IconScoreMap[global.ICON_HUANGBAOSHI][9] = IconConfig.Huangbaoshi[3]
	IconScoreMap[global.ICON_HUANGBAOSHI][10] = IconConfig.Huangbaoshi[4]
	IconScoreMap[global.ICON_HUANGBAOSHI][11] = IconConfig.Huangbaoshi[5]
	IconScoreMap[global.ICON_HUANGBAOSHI][12] = IconConfig.Huangbaoshi[6]
	IconScoreMap[global.ICON_HUANGBAOSHI][13] = IconConfig.Huangbaoshi[7]
	IconScoreMap[global.ICON_HUANGBAOSHI][14] = IconConfig.Huangbaoshi[8]
	IconScoreMap[global.ICON_HUANGBAOSHI][15] = IconConfig.Huangbaoshi[9]
	IconScoreMap[global.ICON_HUANGBAOSHI][16] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][17] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][18] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][19] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][20] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][21] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][22] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][23] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][24] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][25] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][26] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][27] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][28] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][29] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][30] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][31] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][32] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][33] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][34] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][35] = IconConfig.Huangbaoshi[10]
	IconScoreMap[global.ICON_HUANGBAOSHI][36] = IconConfig.Huangbaoshi[10]
	//第三关 蓝宝石
	IconScoreMap[global.ICON_LANBAOSHI] = make(map[int32]int64)
	IconScoreMap[global.ICON_LANBAOSHI][6] = IconConfig.Lanbaoshi[0]
	IconScoreMap[global.ICON_LANBAOSHI][7] = IconConfig.Lanbaoshi[1]
	IconScoreMap[global.ICON_LANBAOSHI][8] = IconConfig.Lanbaoshi[2]
	IconScoreMap[global.ICON_LANBAOSHI][9] = IconConfig.Lanbaoshi[3]
	IconScoreMap[global.ICON_LANBAOSHI][10] = IconConfig.Lanbaoshi[4]
	IconScoreMap[global.ICON_LANBAOSHI][11] = IconConfig.Lanbaoshi[5]
	IconScoreMap[global.ICON_LANBAOSHI][12] = IconConfig.Lanbaoshi[6]
	IconScoreMap[global.ICON_LANBAOSHI][13] = IconConfig.Lanbaoshi[7]
	IconScoreMap[global.ICON_LANBAOSHI][14] = IconConfig.Lanbaoshi[8]
	IconScoreMap[global.ICON_LANBAOSHI][15] = IconConfig.Lanbaoshi[9]
	IconScoreMap[global.ICON_LANBAOSHI][16] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][17] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][18] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][19] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][20] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][21] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][22] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][23] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][24] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][25] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][26] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][27] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][28] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][29] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][30] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][31] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][32] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][33] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][34] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][35] = IconConfig.Lanbaoshi[10]
	IconScoreMap[global.ICON_LANBAOSHI][36] = IconConfig.Lanbaoshi[10]
	//第三关 白钻石
	IconScoreMap[global.ICON_BAIZUANSHI] = make(map[int32]int64)
	IconScoreMap[global.ICON_BAIZUANSHI][6] = IconConfig.Baizuanshi[0]
	IconScoreMap[global.ICON_BAIZUANSHI][7] = IconConfig.Baizuanshi[1]
	IconScoreMap[global.ICON_BAIZUANSHI][8] = IconConfig.Baizuanshi[2]
	IconScoreMap[global.ICON_BAIZUANSHI][9] = IconConfig.Baizuanshi[3]
	IconScoreMap[global.ICON_BAIZUANSHI][10] = IconConfig.Baizuanshi[4]
	IconScoreMap[global.ICON_BAIZUANSHI][11] = IconConfig.Baizuanshi[5]
	IconScoreMap[global.ICON_BAIZUANSHI][12] = IconConfig.Baizuanshi[6]
	IconScoreMap[global.ICON_BAIZUANSHI][13] = IconConfig.Baizuanshi[7]
	IconScoreMap[global.ICON_BAIZUANSHI][14] = IconConfig.Baizuanshi[8]
	IconScoreMap[global.ICON_BAIZUANSHI][15] = IconConfig.Baizuanshi[9]
	IconScoreMap[global.ICON_BAIZUANSHI][16] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][17] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][18] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][19] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][20] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][21] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][22] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][23] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][24] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][25] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][26] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][27] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][28] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][29] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][30] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][31] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][32] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][33] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][34] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][35] = IconConfig.Baizuanshi[10]
	IconScoreMap[global.ICON_BAIZUANSHI][36] = IconConfig.Baizuanshi[10]
}

//初始化连环夺宝config
func LoadLhdbConfig() {
	data, err := ioutil.ReadFile("./config/cheat.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result := gjson.Parse(string(data))
	Analysiscfg(result, 3000)
	Analysiscfg(result, 2000)
	Analysiscfg(result, 1000)
	Analysiscfg(result, -1000)
	Analysiscfg(result, -2000)
	Analysiscfg(result, -3000)
	//读取图标赔付倍数
	//作弊率配置文件
	iconConfigData, err := ioutil.ReadFile("./config/icon.json")
	if err != nil {
		log.Traceln("gameConfigData reading error", err)
		panic("")
		return
	}
	iConConfigResult, _ := GoJsoner.Discard(string(iconConfigData))
	err = json.Unmarshal([]byte(iConConfigResult), &IconConfig)
	if err != nil {
		log.Traceln("Load cheat_config.go file err: ", err.Error())
		panic("")
		return
	}
	InitIconScoreMap()

	configData, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		log.Traceln("configData reading error", err)
		panic("")
	}
	configResult, _ := GoJsoner.Discard(string(configData))
	err = json.Unmarshal([]byte(configResult), &Config)
	if err != nil {
		log.Traceln("configResult cheat_config.go file err: ", err.Error())
		panic("")
	}
}

func GetLhdbConfig(cheat int32) *IconRate {
	index := 2
	switch cheat {
	case 3000:
		index = 0
	case 2000:
		index = 1
	case 1000:
		index = 2
	case -1000:
		index = 3
	case -2000:
		index = 4
	case -3000:
		index = 5
	}
	return IconRateArr[index]
}

func Analysiscfg(cfg gjson.Result, cheatvalue int) {
	tmp := new(IconRate)
	str := fmt.Sprintf("%v.MaxTimes", cheatvalue)
	tmp.MaxTimes = int64(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.MaxDisCount", cheatvalue)
	//log.Traceln("tmp.MaxTimes ",tmp.MaxTimes)
	tmp.MaxDisCount = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.KeyRate", cheatvalue)
	tmp.KeyRate = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.IntoCaijin", cheatvalue)
	tmp.IntoCaijin = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.Caijin1", cheatvalue)
	tmp.Caijin1 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.Caijin2", cheatvalue)
	tmp.Caijin2 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.Caijin3", cheatvalue)
	tmp.Caijin3 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.Caijin4", cheatvalue)
	tmp.Caijin4 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.Caijin5", cheatvalue)
	tmp.Caijin5 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.BaiyuRateA", cheatvalue)
	tmp.BaiyuRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.BaiyuRateB", cheatvalue)
	tmp.BaiyuRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.BiyuRateA", cheatvalue)
	tmp.BiyuRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.BiyuRateB", cheatvalue)
	tmp.BiyuRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.MoyuRateA", cheatvalue)
	tmp.MoyuRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.MoyuRateB", cheatvalue)
	tmp.MoyuRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.ManaoRateA", cheatvalue)
	tmp.ManaoRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.ManaoRateB", cheatvalue)
	tmp.ManaoRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.HupoRateA", cheatvalue)
	tmp.HupoRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.HupoRateB", cheatvalue)
	tmp.HupoRateB = int(cfg.Get(str).Int())
	//第二关
	str = fmt.Sprintf("%v.ZumulvRateA", cheatvalue)
	tmp.ZumulvRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.ZumulvRateB", cheatvalue)
	tmp.ZumulvRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.MaoyanshiRateA", cheatvalue)
	tmp.MaoyanshiRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.MaoyanshiRateB", cheatvalue)
	tmp.MaoyanshiRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.ZishuijingRateA", cheatvalue)
	tmp.ZishuijingRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.ZishuijingRateB", cheatvalue)
	tmp.ZishuijingRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.FeicuishiRateA", cheatvalue)
	tmp.FeicuishiRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.FeicuishiRateB", cheatvalue)
	tmp.FeicuishiRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.BaizhenzhuRateA", cheatvalue)
	tmp.BaizhenzhuRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.BaizhenzhuRateB", cheatvalue)
	tmp.BaizhenzhuRateB = int(cfg.Get(str).Int())
	//第三关
	str = fmt.Sprintf("%v.HongbaoshiRateA", cheatvalue)
	tmp.HongbaoshiRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.HongbaoshiRateB", cheatvalue)
	tmp.HongbaoshiRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.LvbaoshiRateA", cheatvalue)
	tmp.LvbaoshiRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.LvbaoshiRateB", cheatvalue)
	tmp.LvbaoshiRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.HuangbaoshiRateA", cheatvalue)
	tmp.HuangbaoshiRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.HuangbaoshiRateB", cheatvalue)
	tmp.HuangbaoshiRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.LanbaoshiRateA", cheatvalue)
	tmp.LanbaoshiRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.LanbaoshiRateB", cheatvalue)
	tmp.LanbaoshiRateB = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.BaizuanshiRateA", cheatvalue)
	tmp.BaizuanshiRateA = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.BaizuanshiRateB", cheatvalue)
	tmp.BaizuanshiRateB = int(cfg.Get(str).Int())

	IconRateArr = append(IconRateArr, tmp)
}
