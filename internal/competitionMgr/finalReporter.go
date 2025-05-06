package competitionmgr

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	timeParser "yadro_test/common"
)

func (cm CompetitionManager) GenerateReport() error {
	compSlice := cm.sortedCompetitors()

	for _, c := range compSlice {
		err := cm.writeCompetitorReport(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cm CompetitionManager) writeCompetitorReport(c *Competitor) error {
	shots := TargetsPerFiringLine * c.FiringRangeNum
	penaltyHits := countHits(c.Hits)
	penaltyMisses := shots - penaltyHits
	penaltySpeed := computeAvgSpeed(c.PenaltyTime, float64(cm.cfg.PenaltyLen*penaltyMisses))
	totalTimeStr := formatStatus(c.Status, c.TotalTime)
	lapsInfo := formatLapsInfo(c.LapTimes, c.LapSpeeds, cm.cfg.Laps)
	penaltyInfo := formatLapInfo(c.PenaltyTime, penaltySpeed)
	hitsInfo := fmt.Sprintf("%d/%d", penaltyHits, shots)

	line := fmt.Sprintf("%s %d %s %s %s\n",
		totalTimeStr,
		c.CompetitorId,
		lapsInfo,
		penaltyInfo,
		hitsInfo,
	)
	_, err := cm.outputFile.WriteString(line)
	if err != nil {
		return fmt.Errorf("unable to write report to output file: %v", err)
	}
	return nil
}

func formatStatus(status string, duration time.Duration) string {
	switch status {
	case "NotStarted":
		return "[NotStarted]"
	case "NotFinished":
		return "[NotFinished]"
	default:
		return fmt.Sprintf("[%s]", timeParser.ConvertDurationToString(duration))
	}
}

func formatLapsInfo(times []time.Duration, speeds []float64, lapsCount int) string {
	var laps []string
	timesLen := len(times)
	for i := 0; i != lapsCount; i += 1 {
		if i < timesLen {
			laps = append(laps, formatLapInfo(times[i], speeds[i]))
		} else {
			laps = append(laps, formatLapInfo(0, 0))
		}

	}
	lapsInfo := strings.Join(laps, ", ")
	return fmt.Sprintf("[%s]", lapsInfo)
}

func formatLapInfo(dur time.Duration, speed float64) string {
	if dur == 0 && speed == 0 {
		return "{,}"
	}
	return fmt.Sprintf("{%s, %.3f}", timeParser.ConvertDurationToString(dur), truncateFloatWithoutRounding(speed, 3))
}

func truncateFloatWithoutRounding(num float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Trunc(num*factor) / factor
}

func computeAvgSpeed(dur time.Duration, distanceMeters float64) float64 {
	if dur > 0 {
		speedMps := distanceMeters / dur.Seconds()
		return speedMps
	}
	return 0
}

func (cm CompetitionManager) sortedCompetitors() []*Competitor {
	competitors := make([]*Competitor, 0, len(cm.competitors))
	for _, c := range cm.competitors {
		competitors = append(competitors, c)
	}
	sort.Slice(competitors, func(i, j int) bool {
		return competitors[i].TotalTime < competitors[j].TotalTime
	})
	return competitors
}
