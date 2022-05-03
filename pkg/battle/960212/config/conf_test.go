package config

import (
	"testing"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func TestDoudizhuConfig_LoadDoudizhuCfg(t *testing.T) {
	// 加载牌型顺序表配置
	CardsOrderConf.LoadCardsOrderCfg()

	log.Traceln(CardsOrderConf.CardsOrderForm[0].WeightValue)
	log.Traceln(0xa)
	if CardsOrderConf.CardsOrderForm[0].WeightValue == 0xa {
		log.Traceln("equal")
	}
}
