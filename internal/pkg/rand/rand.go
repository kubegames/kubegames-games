package rand

//随机数相关算法
import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Shuffle(n int, swap func(i, j int)) {
	rand.Shuffle(n, swap)
}
func Int31n(n int32) int32 {
	return rand.Int31n(n)
}

func Int63n(n int64) int64 {
	return rand.Int63n(n)
}

func Intn(n int) int {
	return rand.Intn(n)
}

/*
*产生6位随机数
 */
func Rand6() string {
	code := fmt.Sprintf("%06v", rand.Int31n(1000000))
	return code
}

/*
*产生6位随机数
 */
func Rand4() string {
	code := fmt.Sprintf("%04v", rand.Int31n(10000))
	return code
}

var Chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

// var AsciiChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+,.?/:;{}[]`~")

/*
*产生指定长度随机字符串
 */
func NewLenChars(length int) string {
	if length == 0 {
		return ""
	}
	clen := len(Chars)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("Error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue // Skip this number to avoid modulo bias.
			}
			b[i] = Chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

func RandomString(l int) string {
	var result bytes.Buffer
	for i := 0; i < l; {
		result.WriteString(fmt.Sprintf("%d", RandInt(65, 90)))
		i++
	}
	return result.String()
}

func RandInt(min, max int) int {
	if min >= max {
		return max
	}
	return rand.Intn(max-min) + min
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//return r.Intn(max-min) + min
}

//传入指定概率，然后返回是否执行  比如 rate：90 表示有90%的概率要执行
func RateToExec(rate int) bool {
	r := RandInt(0, 100)
	//fmt.Println("随机数r : ",r)
	if r <= rate {
		return true
	}
	return false
}

//传入指定概率，然后返回是否执行  比如 rate：90 表示有90%的概率要执行
func RateToExecWan(rate int) bool {
	r := RandInt(0, 10000)
	//fmt.Println("随机数r : ",r)
	if r <= rate {
		return true
	}
	return false
}

//从max中随机去一个数，看是否小于rate
func RateToExecWithIn(rate, max int) bool {
	r := RandInt(1, max)
	if r < rate {
		return true
	}
	return false
}
