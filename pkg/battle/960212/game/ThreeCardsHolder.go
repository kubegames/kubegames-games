package game

import (
	"sync"
)

type ThreeCardsHolder struct {
	cards [3]byte
	m     sync.Mutex

	cardsValue       map[int]int
	cardsValue3Cards map[int]int
}

func (this *ThreeCardsHolder) Set(cards []byte) {
	this.m.Lock()
	defer this.m.Unlock()
	copy(this.cards[:], cards)
}
func (this *ThreeCardsHolder) Get() [3]byte {
	this.m.Lock()
	defer this.m.Unlock()
	return this.cards
}

func (this *ThreeCardsHolder) SetCardsValue(seatId int, v int) {
	this.m.Lock()
	defer this.m.Unlock()
	this.cardsValue[seatId] = v
}
func (this *ThreeCardsHolder) SetCardsValue3(seatId int, v int) {
	this.m.Lock()
	defer this.m.Unlock()
	this.cardsValue3Cards[seatId] = v
}
func (this *ThreeCardsHolder) GetCardsValue3(seatId int) int {
	this.m.Lock()
	defer this.m.Unlock()
	return this.cardsValue3Cards[seatId]
}
func (this *ThreeCardsHolder) GetCardsValue(seatId int) int {
	this.m.Lock()
	defer this.m.Unlock()
	return this.cardsValue[seatId]
}

var (
	instance *ThreeCardsHolder
	once     sync.Once
)

func GetThreeCardsHolderInstance() *ThreeCardsHolder {
	once.Do(func() {
		instance = &ThreeCardsHolder{cardsValue: make(map[int]int, 3), cardsValue3Cards: make(map[int]int, 3)}
	})
	return instance
}
