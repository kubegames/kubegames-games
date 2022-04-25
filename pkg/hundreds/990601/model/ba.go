package model

// type BirdAnimalType int

// const (
// 	BirdAnimalTypeAllKill     BirdAnimalType = iota + 1 // 通杀
// 	BirdAnimalTypeAllPay                                // 通赔
// 	BirdAnimalTypeGoldShark                             // 金鲨
// 	BirdAnimalTypeSilverShark                           // 银鲨
// )

// type ProbBase struct {
// 	Odds int `json:"odds"` // 赔率

// 	OddsMax int `json:"oddsMax"` // 最小赔率
// 	OddsMin int `json:"oddsMin"` // 最大赔率

// 	Prob int `json:"prob"` // 权重
// }

// type BirdAnimalBase struct {
// 	ID         int         `json:"id"`
// 	Name       string      `json:"name"`
// 	WinConfig  []*ProbBase `json:"winConfig"`  // 赢赔率设置
// 	LossConfig []*ProbBase `json:"lossConfig"` // 输赔率设置
// 	NowOdds    int         `json:"-"`          // 当前赔率
// }

// type BirdAnimalBases []*BirdAnimalBase

// func (b BirdAnimalBases) RandOdds() {
// 	// for _, v := range b {
// 	// 	var allProb int
// 	// 	for _, val := range v.P {
// 	// 		allProb += val.Prob
// 	// 	}
// 	// 	randProb := rand.Intn(allProb)
// 	// 	// var leftProb int
// 	// 	for _, val := range v.P {
// 	// 		if val.Prob >= randProb {
// 	// 			v.NowOdds = val.Prob
// 	// 			break
// 	// 		}
// 	// 	}

// 	// }
// }

