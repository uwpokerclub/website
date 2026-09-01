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
