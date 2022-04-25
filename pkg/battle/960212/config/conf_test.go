package config

import (
	"fmt"
	"testing"
)

func TestDoudizhuConfig_LoadDoudizhuCfg(t *testing.T) {
	// 加载牌型顺序表配置
	CardsOrderConf.LoadCardsOrderCfg()

	fmt.Println(CardsOrderConf.CardsOrderForm[0].WeightValue)
	fmt.Println(0xa)
	if CardsOrderConf.CardsOrderForm[0].WeightValue == 0xa {
		fmt.Println("equal")
	}
}
