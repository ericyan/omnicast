package types

import (
	"fmt"
	"strconv"
)

var allowedRats = []Rat{
	Rat{1, 10},
	Rat{1, 8},
	Rat{1, 5},
	Rat{1, 4},
	Rat{3, 10},
	Rat{1, 3},
	Rat{3, 8},
	Rat{2, 5},
	Rat{1, 2},
	Rat{3, 5},
	Rat{5, 8},
	Rat{2, 3},
	Rat{7, 10},
	Rat{3, 4},
	Rat{4, 5},
	Rat{7, 8},
	Rat{9, 10},
}

// A Rat represents a quotient a/b.
type Rat struct {
	a, b int
}

// ParseFloat32 returns the nearest Rat to x.
func ParseFloat32(x float32) *Rat {
	if x == 1.0 || x < 0 {
		return &Rat{1, 1}
	}

	if x >= 2.0 {
		return &Rat{int(x), 1}
	}

	d := 0
	if x > 1 {
		d = 1
		x--
	}

	var (
		lastDiff float32
		nearest  Rat
	)
	for i, y := range allowedRats {
		diff := x - float32(y.a)/float32(y.b)
		if diff < 0 {
			diff = -diff
		}

		if diff < 0.01 {
			nearest = y
			break
		}

		if diff > lastDiff && lastDiff > 0 {
			nearest = allowedRats[i-1]
			break
		}

		lastDiff = diff
	}

	return &Rat{nearest.a + nearest.b*d, nearest.b}
}

// String returns a string representation in the form "a/b" if b != 1,
// and in the form "a" if b == 1.
func (x *Rat) String() string {
	if x.b == 1 {
		return strconv.Itoa(x.a)
	}

	return fmt.Sprintf("%d/%d", x.a, x.b)
}
