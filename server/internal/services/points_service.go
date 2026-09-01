package services

import "math"

// pointsTuningConstant scales the logarithmic curve to keep integer resolution in the tail. It
// does not affect the shape of the curve, only its magnitude.
const pointsTuningConstant = 25.0

// rawPoints returns the unrounded, unmultiplied point value for finishing at position place in
// an event of eventSize scored entries. place must be in [1, eventSize].
func rawPoints(eventSize, place int) float64 {
	return pointsTuningConstant*math.Log(float64(eventSize)/float64(place)) + 1
}

// CalculateTiePoints returns the points awarded to every member of a tie group occupying
// positions from..to (1-indexed, inclusive) in an event of eventSize scored entries: the mean of
// the raw curve value across the group's positions, multiplied by multiplier, rounded once at
// the end. A singleton group (from == to) reduces to the plain per-position formula, so there is
// one code path for both tied and untied entries.
func CalculateTiePoints(eventSize, from, to int, multiplier float32) int {
	sum := 0.0
	for place := from; place <= to; place++ {
		sum += rawPoints(eventSize, place)
	}
	mean := sum / float64(to-from+1)
	return int(math.Round(mean * float64(multiplier)))
}

// DEPRECATED: CalculatePoints uses the old fixed payout table. Use CalculateTiePoints instead.
// This function remains for backward compatibility until events_service is updated.
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
