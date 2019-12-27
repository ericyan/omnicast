package types

import (
	"fmt"
	"strconv"
)

// A Rat represents a quotient a/b.
type Rat struct {
	a, b int
}

// ParseFloat32 returns a Rat that is the closest to x.
func ParseFloat32(x float32) *Rat {
	return &Rat{1, 1} // TODO
}

// String returns a string representation in the form "a/b" if b != 1,
// and in the form "a" if b == 1.
func (x *Rat) String() string {
	if x.b == 1 {
		return strconv.Itoa(x.a)
	}

	return fmt.Sprintf("%d/%d", x.a, x.b)
}
