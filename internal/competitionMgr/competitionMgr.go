package competitionmgr

import (
	"fmt"
	"os"
	"time"
	timeParser "yadro_test/common"
	"yadro_test/internal/cfg"
	lh "yadro_test/internal/logger"
)

type CompetitionManager struct {
	outputFile  *os.File
	cfg         *cfg.Config
	competitors map[int]Competitor
}

type Competitor struct {
	Status           string
	StartTime        time.Time
	LastLapTime      time.Time
	TotalTime        time.Duration
	lapTimes         []time.Duration
	lapSpeeds        []float64
	penaltyTime      time.Duration
	penaltySpeed     float64
	hits             map[byte]bool
	lapsEnded        uint
	penaltyLapsEnter time.Time
}

func NewCompetitionManager(outFile *os.File, cfg *cfg.Config) *CompetitionManager {
	return &CompetitionManager{
		outputFile:  outFile,
		cfg:         cfg,
		competitors: make(map[int]Competitor, 0),
	}
}

func (cm CompetitionManager) HandleEvent(eventInfo lh.EventInfo) error {
	switch eventInfo.EventId {
	case 1:
		cm.competitors[eventInfo.CompetitorId] = Competitor{
			lapTimes:  make([]time.Duration, 0),
			lapSpeeds: make([]float64, 0),
			hits:      map[byte]bool{},
		}
	case 2:
		competitor := cm.competitors[eventInfo.CompetitorId]
		startTime, err := timeParser.ConvertStringToTime(eventInfo.ExtraParams)
		if err != nil {
			return err
		}

		competitor.LastLapTime = startTime
		competitor.StartTime = startTime
	case 3, 5, 7:
		return nil
	case 4:
		competitor := cm.competitors[eventInfo.CompetitorId]
		startDeltaDur, err := timeParser.ConvertStringToDuration(cm.cfg.StartDelta)
		if err != nil {
			return err
		}
		diff := eventInfo.EventTime.Sub(competitor.LastLapTime)

		if diff > startDeltaDur {
			competitor.Status = "NotStarted"
			competitor.TotalTime = startDeltaDur
		}
	case 6:
		competitor := cm.competitors[eventInfo.CompetitorId]
		competitor.hits[eventInfo.ExtraParams[0]] = true
	case 8:
		competitor := cm.competitors[eventInfo.CompetitorId]
		competitor.penaltyLapsEnter = eventInfo.EventTime
	case 9:
		competitor := cm.competitors[eventInfo.CompetitorId]
		penaltyLaps := countPenaltyLaps(competitor.hits)
		time, speed := calculateLapStats(competitor.penaltyLapsEnter, eventInfo.EventTime, float64(cm.cfg.PenaltyLen*penaltyLaps))

		competitor.penaltyTime = time
		competitor.penaltySpeed = speed
	case 10:
		competitor := cm.competitors[eventInfo.CompetitorId]
		time, speed := calculateLapStats(competitor.LastLapTime, eventInfo.EventTime, float64(cm.cfg.LapLen))
		competitor.lapTimes = append(competitor.lapTimes, time)
		competitor.lapSpeeds = append(competitor.lapSpeeds, speed)

		if competitor.lapsEnded == uint(cm.cfg.Laps)-1 {
			competitor.Status = "Finished"
			competitor.TotalTime = calculateTotalTime(competitor.StartTime, competitor.lapTimes)
			competitor.lapsEnded += 1
		} else if competitor.lapsEnded == uint(cm.cfg.Laps) {
			return fmt.Errorf("competitor ended more laps than needed")
		}
	case 11:
		competitor := cm.competitors[eventInfo.CompetitorId]

		competitor.Status = "NotFinished"
		competitor.TotalTime = eventInfo.EventTime.Sub(competitor.StartTime)
	}
	return nil
}

func calculateLapStats(startTime, endTime time.Time, distanceMeters float64) (time.Duration, float64) {
	var duration time.Duration
	var speedMps float64
	duration = endTime.Sub(startTime)

	if duration > 0 {
		speedMps = distanceMeters / duration.Seconds()
	}
	return duration, speedMps
}

func countPenaltyLaps(hits map[byte]bool) int {
	count := 0
	for _, v := range hits {
		if !v {
			count += 1
		}
	}
	return count
}

func calculateTotalTime(startTime time.Time, lapTimes []time.Duration) time.Duration {
	finishTime := startTime
	for _, dur := range lapTimes {
		finishTime = finishTime.Add(dur)
	}
	return finishTime.Sub(startTime)
}
