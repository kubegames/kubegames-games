package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_buyu/980301/config"
	"go-game-sdk/example/game_buyu/980301/server"
	"math/rand"
	"time"
)

func main() {
	//defer recover_handle.RecoverHandle()
	//fmt.Println("****************************************")
	//fmt.Println("*                                      *")
	//fmt.Println("*       Poker   Jh   System !          *")
	//fmt.Println("*                                      *")
	//fmt.Println("****************************************\n")
	//
	//fmt.Println("### VER: ", "0.0.8")
	//fmt.Println("### PID: ", os.Getpid())
	//
	////系统中断捕获
	//sigs := make(chan os.Signal, 1)
	//over := make(chan int, 0)
	//signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	//go func() {
	//	sig := <-sigs
	//	fmt.Println("###SIGNAL::", sig, "   PID:", os.Getpid())
	//	over <- 1
	//}()
	//
	////开启pprof
	//go func() {
	//	fmt.Println("pprof start at :9876")
	//	fmt.Println(http.ListenAndServe(":9876", nil))
	//}()
	//
	////可修改部分---开始
	//err := conf.LoadJsonConfig("./conf/config.json", conf.Config)
	//if err != nil {
	//	log.Errorf("Load config.json file err:%s ", err.Error())
	//	return
	//}
	////初始化房间
	////game.InitRoom()
	//
	////可修改部分---结束
	//
	////启动服务
	//server.StartServer(conf.Config.ListenerPort, conf.Config.MaxConn, 0)
	//
	//<-over
	//log.Tracef("###Over。。。。。")
	config.Load()

	rand.Seed(time.Now().UnixNano())
	//roomConfig := game_frame.RoomConfig{
	//	Port:      8080, //端口号
	//	Count:     5000, //最大连接数
	//	HeartBeat: 0,    //心跳 0（表示不设置心跳）
	//	TableNum:  500,  //房间桌子数量
	//	Max:       4,    //房间桌子的最大容纳人数
	//	Min:       1,    //房间桌子的最小开赛数
	//}
	room := game_frame.NewRoom(&server.Server{})
	room.Run()
}
