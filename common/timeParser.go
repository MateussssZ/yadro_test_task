package timeParser

import (
	"fmt"
	"log"
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
		log.Fatalf("unable to parse time.Duration(%s)", t)
		return 0, err
	}
	return time.Duration(parsed.Hour())*time.Hour +
		time.Duration(parsed.Minute())*time.Minute +
		time.Duration(parsed.Second())*time.Second, nil
}
