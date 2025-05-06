package timeParser

import (
	"fmt"
	"time"
)

func ConvertStringToTime(t string) (time.Time, error) {
	time, err := time.Parse("15:04:05.000", t)
	if err != nil {
		return time, fmt.Errorf("unable to parse time.Time(%s)", t)
	}
	return time, nil
}

func ConvertStringToDuration(t string) (time.Duration, error) {
	parsed, err := time.Parse("15:04:05", t)
	if err != nil {
		return 0, fmt.Errorf("unable to parse time.Duration(%s)", t)
	}
	return time.Duration(parsed.Hour())*time.Hour +
		time.Duration(parsed.Minute())*time.Minute +
		time.Duration(parsed.Second())*time.Second, nil
}

func ConvertDurationToString(dur time.Duration) string {
	dur = dur.Round(time.Millisecond) // Округляем до миллисекунд
	h := dur / time.Hour
	dur -= h * time.Hour
	m := dur / time.Minute
	dur -= m * time.Minute
	s := dur / time.Second
	dur -= s * time.Second
	ms := dur / time.Millisecond
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}
