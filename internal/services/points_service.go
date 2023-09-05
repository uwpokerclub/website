package services

import "math"

const (
	SizeFactor float64 = 50.0
)

func getPayout(placement int) int {
	payouts := map[int]int{
		1:  32,
		2:  28,
		3:  24,
		4:  21,
		5:  18,
		6:  16,
		7:  14,
		8:  12,
		9:  11,
		10: 10,
		11: 9,
		12: 9,
		13: 8,
		14: 8,
		15: 7,
		16: 7,
		17: 6,
		18: 6,
		19: 5,
		20: 5,
		21: 4,
		22: 4,
		23: 4,
		24: 4,
		25: 4,
		26: 3,
		27: 3,
		28: 3,
		29: 3,
		30: 3,
		31: 2,
		32: 2,
		33: 2,
		34: 2,
		35: 2,
		36: 2,
		37: 2,
		38: 2,
		39: 2,
		40: 2,
	}

	if placement > 40 {
		return 1
	}

	return payouts[placement]
}

func CalculatePoints(eventSize int, placement int, pointsMultiplier float32) int {
	return int(math.Ceil(float64((getPayout(placement)*eventSize))/SizeFactor) * float64(pointsMultiplier))
}
