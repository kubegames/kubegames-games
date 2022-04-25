package score

import "fmt"

func GetScoreStr(score int64) string {
	return fmt.Sprintf(`%d`, score)
	// yuan := score / 100
	// remain := score % 100
	// if remain < 0 {
	// 	remain = -remain
	// }
	// jiao := remain / 10
	// fen := remain % 10
	// return fmt.Sprintf(`%d.%d%d`, yuan, jiao, fen)
}
