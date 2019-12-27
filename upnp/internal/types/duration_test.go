package types

import (
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	cases := []struct {
		s string
		d time.Duration
	}{
		{"0:00:00", time.Duration(0)},
		{"0:00:10", time.Duration(10 * time.Second)},
		{"0:01:00", time.Duration(1 * time.Minute)},
		{"1:00:00", time.Duration(1 * time.Hour)},
	}

	for _, c := range cases {
		if got := FormatDuration(c.d); c.s != got {
			t.Errorf("Format(%s): got %s; want %s", c.d, got, c.s)
		}

		if got, _ := ParseDuration(c.s); c.d != got {
			t.Errorf("Parse(%s): got %s; want %s", c.s, got, c.d)
		}

	}
}
