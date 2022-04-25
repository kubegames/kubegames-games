package model

import (
	"math/rand"
	"sort"
	// "fmt"
)

// 随机子元素id
func (bb ElemBase) RandSubId() int {
	randIndex := rand.Intn(len(bb.SubIds))
	return bb.SubIds[randIndex]
}

type ElemBases []ElemBase

// shakeProb 从大到小排序
func (eb ElemBases) Less(i, j int) bool {
	if eb[i].ShakeProb >= eb[j].ShakeProb {
		return true
	}
	return false
}

func (eb ElemBases) Swap(i, j int) {
	eb[i], eb[j] = eb[j], eb[i]
}

func (eb ElemBases) Len() int {
	return len(eb)
}

// 随机结果
// param: baseType:直接摇中传入的参数（大三元和大四喜除外）
// param: rmBig:移除大三元和大四喜
// ElemBase: 根据参数摇中的结果
func (ebs ElemBases) RandResult(baseType ElementType, rmBig bool) ElemBase {
	if !sort.IsSorted(ebs) {
		sort.Sort(ebs)
	}
	var checkBase bool
	switch baseType {
	case
		BenzRed,
		BenzGreen,
		BenzBlack,
		// 宝马
		BMWRed,
		BMWGreen,
		BMWBlack,
		// 雷克萨斯
		LexusRed,
		LexusGreen,
		LexusBlack,
		// 大众
		VWRed,
		VWGreen,
		VWBlack:
		checkBase = true
	}

LOOP:
	var allWeight int
	for _, v := range ebs {
		if checkBase && v.ElemType == baseType {
			return v
		}
		allWeight += v.ShakeProb
	}
	randWeight := rand.Intn(allWeight) + 1
	for _, v := range ebs {
		if randWeight <= v.ShakeProb {
			if (v.ElemType == BigFourElem || v.ElemType == BigThreeElem) && rmBig {
				goto LOOP
			}
			return v
		}
		randWeight -= v.ShakeProb
	}
	return ebs[0]
}

func (ebs ElemBases) GetByID(id int) ElemBase {
	if id < 0 || id > 12 {
		return ebs[rand.Intn(12)]
	}
	return ebs[id]
}

// 根据类型或颜色来筛选车
func (ebs ElemBases) FindWithType(et ElementType) (ElemBases, ElementType) {
	var result ElemBases
	for _, v := range ebs {
		if v.ElemType&et == et {
			result = append(result, v)
		}
	}
	return result, et
}
