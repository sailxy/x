package id

import (
	"fmt"
	"math/rand"
)

func NewRandomNumber(length int) (int, error) {
	if length <= 0 {
		return 0, fmt.Errorf("length must be greater than 0")
	}

	// Calculate the lower and upper bounds for the number of the specified length.
	lower := 1
	upper := 1
	for i := range length {
		upper *= 10
		if i > 0 {
			lower *= 10
		}
	}
	// To ensure the number has exactly the specified length.
	upper--

	// Generate the random number.
	number := lower + rand.Intn(upper-lower+1)

	return number, nil
}
