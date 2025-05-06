package competitionmgr

import (
	"fmt"
	"os"
	"strconv"
	"time"

	timeParser "yadro_test/common"
	"yadro_test/internal/cfg"
	lh "yadro_test/internal/logger"
)

type CompetitionManager struct {
	outputFile  *os.File
	cfg         *cfg.Config
	competitors map[int]*Competitor
}

type Competitor struct {
	CompetitorId     int
	Status           string
	StartTime        time.Time
	LastLapTime      time.Time
	TotalTime        time.Duration
	LapTimes         []time.Duration
	LapSpeeds        []float64
	PenaltyTime      time.Duration
	PenaltySpeed     float64
	Hits             []bool
	LapsEnded        uint
	PenaltyLapsEnter time.Time
	FiringRangeNum   int
}

func NewCompetitionManager(outFile *os.File, cfg *cfg.Config) *CompetitionManager {
	return &CompetitionManager{
		outputFile:  outFile,
		cfg:         cfg,
		competitors: make(map[int]*Competitor, 0),
	}
}

func (cm CompetitionManager) HandleEvent(eventInfo lh.EventInfo) error {
	switch eventInfo.EventId {
	case 1:
		cm.competitors[eventInfo.CompetitorId] = &Competitor{
			CompetitorId: eventInfo.CompetitorId,
			LapTimes:     make([]time.Duration, 0),
			LapSpeeds:    make([]float64, 0),
			Hits:         make([]bool, 5*cm.cfg.FiringLines),
		}
	case 2:
		competitor := cm.competitors[eventInfo.CompetitorId]
		startTime, err := timeParser.ConvertStringToTime(eventInfo.ExtraParams)
		if err != nil {
			return err
		}

		competitor.LastLapTime = startTime
		competitor.StartTime = startTime
	case 3, 7:
		return nil
	case 4:
		competitor := cm.competitors[eventInfo.CompetitorId]
		startDeltaDur, err := timeParser.ConvertStringToDuration(cm.cfg.StartDelta)
		if err != nil {
			return err
		}
		diff := eventInfo.EventTime.Sub(competitor.LastLapTime)

		if diff > startDeltaDur || diff < 0 {
			competitor.Status = "NotStarted"
			competitor.TotalTime = startDeltaDur
		}
	case 5:
		competitor := cm.competitors[eventInfo.CompetitorId]
		rangeNum, err := strconv.Atoi(eventInfo.ExtraParams)
		if err != nil {
			return fmt.Errorf("unable to convert rangeNum to int(%s)", eventInfo.ExtraParams)
		}
		competitor.FiringRangeNum = rangeNum - 1
	case 6:
		competitor := cm.competitors[eventInfo.CompetitorId]
		targetNum, err := strconv.Atoi(eventInfo.ExtraParams)
		if err != nil {
			return fmt.Errorf("unable to convert targetNum to int(%s)", eventInfo.ExtraParams)
		}
		targetNum = 5*competitor.FiringRangeNum + targetNum - 1
		maxFiringLines := 5 * cm.cfg.FiringLines
		if targetNum > maxFiringLines {
			return fmt.Errorf("target num can`t be more than %d(got %d)", maxFiringLines, targetNum)

		}
		competitor.Hits[targetNum] = true
	case 8:
		competitor := cm.competitors[eventInfo.CompetitorId]
		competitor.PenaltyLapsEnter = eventInfo.EventTime
	case 9:
		competitor := cm.competitors[eventInfo.CompetitorId]
		time := eventInfo.EventTime.Sub(competitor.PenaltyLapsEnter)

		competitor.PenaltyTime += time
	case 10:
		competitor := cm.competitors[eventInfo.CompetitorId]
		time, speed := calculateLapStats(competitor.LastLapTime, eventInfo.EventTime, float64(cm.cfg.LapLen))
		competitor.LapTimes = append(competitor.LapTimes, time)
		competitor.LapSpeeds = append(competitor.LapSpeeds, speed)
		competitor.LastLapTime = eventInfo.EventTime

		if competitor.LapsEnded == uint(cm.cfg.Laps)-1 {
			competitor.Status = "Finished"
			competitor.TotalTime = calculateTotalTime(competitor.StartTime, competitor.LapTimes)
		} else if competitor.LapsEnded == uint(cm.cfg.Laps) {
			return fmt.Errorf("competitor ended more laps than needed")
		}
		competitor.LapsEnded += 1
	case 11:
		competitor := cm.competitors[eventInfo.CompetitorId]

		competitor.Status = "NotFinished"
		competitor.TotalTime = eventInfo.EventTime.Sub(competitor.StartTime)
	}
	return nil
}

func countMisses(hits []bool) int {
	count := 0
	for _, v := range hits {
		if !v {
			count += 1
		}
	}
	return count
}

func calculateLapStats(startTime, endTime time.Time, distanceMeters float64) (time.Duration, float64) {
	duration := endTime.Sub(startTime)
	speedMps := computeAvgSpeed(duration, distanceMeters)
	return duration, speedMps
}

func calculateTotalTime(startTime time.Time, lapTimes []time.Duration) time.Duration {
	finishTime := startTime
	for _, dur := range lapTimes {
		finishTime = finishTime.Add(dur)
	}
	return finishTime.Sub(startTime)
}
