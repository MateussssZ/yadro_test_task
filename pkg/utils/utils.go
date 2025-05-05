package utils

import (
	"log"
	"time"
)

func ConvertStringToTime(t string) time.Time {
	time, err := time.Parse("15:04:05.000", t)
	if err != nil {
		log.Fatalf("unable to parse time.Time(%s)", t)
	}
	return time
}

func ConvertStringToDuration(t string) time.Duration {
	parsed, err := time.Parse("15:04:05", t)
	if err != nil {
		log.Fatalf("unable to parse time.Duration(%s)", t)
		return 0
	}
	return time.Duration(parsed.Hour())*time.Hour +
		time.Duration(parsed.Minute())*time.Minute +
		time.Duration(parsed.Second())*time.Second
}
