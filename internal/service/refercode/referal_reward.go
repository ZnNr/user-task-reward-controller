package refercode

const rewardProcent = 10

func Reward(price int) int {
	if price < 10 {
		return 1
	}

	return price/rewardProcent + 1
}
