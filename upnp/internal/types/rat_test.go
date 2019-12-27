package types

import (
	"testing"
)

func TestRat(t *testing.T) {
	cases := []struct {
		f float32
		s string
	}{
		{0.1, "1/10"},
		{0.3, "3/10"},
		{0.31, "3/10"},
		{0.32, "1/3"},
		{1.5, "3/2"},
		{1.75, "7/4"},
		{2.0, "2"},
	}

	for _, c := range cases {
		if got := ParseFloat32(c.f).String(); c.s != got {
			t.Errorf("ParseFloat32(%f).String(): got %s; want %s", c.f, got, c.s)
		}
	}
}
