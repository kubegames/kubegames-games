package model_test

import (
	model "game_LaBa/benzbmw/model"
	"log"
	"testing"
)

func TestElemBasesRandResult(t *testing.T) {
	ele := model.ElemShakeProbSlice.RandResult(model.BenzRed, false)
	log.Println("ele=", ele)
}

func TestElemBasesRandResultKnow(t *testing.T) {
	ele := model.ElemShakeProbSlice.RandResult(model.BenzRed, false)
	log.Println("ele=", ele)
}

func BenchmarkElemBasesRandResult(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = model.ElemShakeProbSlice.RandResult(model.ElementTypeNil, false)
		// log.Println("ele=", ele)
	}
}


// go test -v  -run TestElemBasesFindWithType
func TestElemBasesFindWithType(t *testing.T) {
	result := model.ElemShakeProbSlice.FindWithType(128)
	log.Println("result=", result)
}
