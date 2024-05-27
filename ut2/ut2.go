package ut2

import (
	"encoding/json"
	"math/rand"
	"time"
)

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Abs[T Number](n T) T {
	if n < 0 {
		return -n
	}

	return n
}

func Clamp[T Number](value, min, max T) T {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func IndexOf[T Comparable](arr []T, target T) int {
	for i, num := range arr {
		if num == target {
			return i
		}
	}
	return -1
}

func RemoveByIdx[T Comparable](arr *[]T, i int) {
	*arr = append((*arr)[:i], (*arr)[i+1:]...)
}

func TryRemoveByValue[T Comparable](arr *[]T, v T) int {
	var tail = -1
	for i := 0; i < len(*arr); i++ {
		findit := (*arr)[i] == v
		if tail == -1 {
			if findit {
				tail = i
			}
		} else {
			if !findit {
				(*arr)[tail] = (*arr)[i]
				tail++
			}
		}
	}

	if tail == -1 {
		return 0
	}

	ret := len(*arr) - tail

	*arr = (*arr)[:tail]

	return ret
}

func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := rand.Intn(26) + 'a'
		bytes[i] = byte(b)
	}
	return string(bytes)
}

func ShuffleArray[T any](arr []T) {
	for i := len(arr) - 1; i >= 1; i-- {
		var ridx = rand.Intn(i)
		arr[i], arr[ridx] = arr[ridx], arr[i]
	}
}

// 随机 [low, high]
func RandomInt(low int, high int) int {
	if low == high {
		return low
	}
	if low > high || high == 0 {
		return 0
	}

	return rand.Intn(high-low+1) + low
}

func RandomFromArray[T any](arr []T) T {
	idx := rand.Intn(len(arr))
	return arr[idx]
}

func ToJson(v any) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}

func MidnightTimeBy(now time.Time) time.Time {
	// now = now.In(time.FixedZone("CST", 7*3600))
	banye := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return banye
}

func IsSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay()
}

func MidnightTime() time.Time {
	now := time.Now()
	return MidnightTimeBy(now)
}
