package loghandler

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

var mapEvents = map[int]string{
	1:  "The competitor(%d) registered",
	2:  "The start time for the competitor(%d) was set by a draw to %s",
	3:  "The competitor(%d) is on the start line",
	4:  "The competitor(%d) has started",
	5:  "The competitor(%d) is on the firing range(%s)",
	6:  "The target(%s) has been hit by competitor(%d)",
	7:  "The competitor(%d) left the firing range",
	8:  "The competitor(%d) entered the penalty laps",
	9:  "The competitor(%d) left the penalty laps",
	10: "The competitor(%d) ended the main lap",
	11: "The competitor(%d) can`t continue: %s",
}

type EventInfo struct {
	EventId      int
	CompetitorId int
	ExtraParams  string
	EventTime    time.Time
}

type CustomLogger struct {
	l *slog.Logger
}

func NewCustomLogger(logFile *os.File) *CustomLogger {
	return &CustomLogger{
		l: slog.New(&CustomHandler{logFile: logFile}),
	}
}

func (cl CustomLogger) ProcessLine(line string) EventInfo {
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		log.Fatalf("insufficient number of parameters in line (%s)", line)
	}

	time, eventIdStr, competitorIdStr := parts[0], parts[1], parts[2]
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		log.Fatalf("can`t convert eventId(%s) to int", eventIdStr)
	}
	competitorId, err := strconv.Atoi(competitorIdStr)
	if err != nil {
		log.Fatalf("can`t convert competitorId(%s) to int", competitorIdStr)
	}

	var extraParams string
	if len(parts) > 3 {
		extraParams = strings.Join(parts[3:], " ")
	}

	msg := buildLogMessage(time, competitorId, eventId, extraParams)
	cl.l.Info(msg)

	time = strings.Trim(time, "[]")
	return EventInfo{
		EventId:      eventId,
		CompetitorId: competitorId,
		EventTime:    convertStringToTime(time),
		ExtraParams:  extraParams,
	}
}

func buildLogMessage(time string, competitorId, eventId int, extraParams string) string {
	var eventMsg string
	switch eventId {
	case 2, 5, 11:
		eventMsg = fmt.Sprintf(mapEvents[eventId], competitorId, extraParams)
	case 6:
		eventMsg = fmt.Sprintf(mapEvents[eventId], extraParams, competitorId)
	default:
		eventMsg = fmt.Sprintf(mapEvents[eventId], competitorId)
	}

	return fmt.Sprintf("%s %s", time, eventMsg)
}

func convertStringToTime(t string) time.Time {
	time, err := time.Parse("15:04:05.000", t)
	if err != nil {
		log.Fatalf("unable to parse time(%s)", t)
	}
	return time
}
