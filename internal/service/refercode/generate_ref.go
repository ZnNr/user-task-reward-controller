package refercode

import "math/rand"

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rewardProcent = 10
)

func RandStringBytes() string {
	n := 15
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Reward(price int) int {
	if price < 10 {
		return 1
	}

	return price/rewardProcent + 1
}
