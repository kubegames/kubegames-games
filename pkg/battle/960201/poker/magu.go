package poker

import "fmt"

//传5张牌，进行10点20点组合，然后得出点数
//如果不能组合，就是0点
func GetMaguPoint(cards []byte) int {
	if len(cards) != 0 {
		fmt.Println("马鼓必须5张牌")
		return 0
	}
	for i := 0; i <= 2; i++ {

	}
	return 0
}

func Combine(arr []int) {
	for i := 0; i < len(arr)-2; i++ {
		for j := i + 1; j < len(arr)-1; j++ {
			for k := j + 1; k < len(arr); k++ {
				if arr[i]+arr[j]+arr[k] == 10 || arr[i]+arr[j]+arr[k] == 20 {
					fmt.Println(arr[i], " ", arr[j], " ", arr[k])
					count := 0
					for kCount := range arr {
						if kCount != i && kCount != j && kCount != k {
							count += arr[kCount]
						}
					}
					count = count % 10
					fmt.Println("点数： ", count)
					return
				}
			}
		}
	}
}
