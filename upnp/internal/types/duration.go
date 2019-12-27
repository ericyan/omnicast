package types

import (
	"fmt"
	"time"
)

// FormatDuration returns a string representation of the duration in the
// form of h:mm:ss.
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second
	d -= s * time.Second

	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}

// ParseDuration parses the string in the form of h:mm:ss.
func ParseDuration(str string) (time.Duration, error) {
	var h, m, s time.Duration
	_, err := fmt.Sscanf(str, "%d:%02d:%02d", &h, &m, &s)
	if err != nil {
		return 0, err
	}

	return h*time.Hour + m*time.Minute + s*time.Second, nil
}
