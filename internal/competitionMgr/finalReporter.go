package competitionmgr

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	timeParser "yadro_test/common"
)

type ReportInfo struct {
	CompetitorId   int
	Status         string
	TotalTime      time.Duration
	LapTimes       []time.Duration
	LapSpeeds      []float64
	PenaltyTime    time.Duration
	Hits           []bool
	FiringRangeNum int
}

func (cm CompetitionManager) GenerateReport() error {
	compSlice := cm.sortedCompetitors() //Отсортируем наших получившихся участников по времени

	for _, c := range compSlice { //Для каждого участника запишем report
		err := cm.writeCompetitorReport(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cm CompetitionManager) writeCompetitorReport(c *Competitor) error {
	shots := TargetsPerFiringLine * c.FiringRangeNum //Здесь просто считаются все метрики по очереди
	penaltyHits := countHits(c.Hits)
	penaltyMisses := shots - penaltyHits
	penaltySpeed := computeAvgSpeed(c.PenaltyTime, float64(cm.cfg.PenaltyLen*penaltyMisses))

	totalTimeStr := formatStatus(c.Status, c.TotalTime) //Здесь наши метрики форматируются в строки для вывода
	lapsInfo := formatLapsInfo(c.LapTimes, c.LapSpeeds, cm.cfg.Laps)
	penaltyInfo := formatLapInfo(c.PenaltyTime, penaltySpeed)
	hitsInfo := fmt.Sprintf("%d/%d", penaltyHits, shots)

	line := fmt.Sprintf("%s %d %s %s %s\n", //Составляем одну общую строку
		totalTimeStr,
		c.CompetitorId,
		lapsInfo,
		penaltyInfo,
		hitsInfo,
	)
	_, err := cm.outputFile.WriteString(line) //Записываем эту строку в одну строчку
	if err != nil {
		return fmt.Errorf("unable to write report to output file: %v", err)
	}
	return nil
}

func formatStatus(status string, duration time.Duration) string { //Если статус !Finished - выводим его, иначе - выводим время
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
	for i := 0; i != lapsCount; i += 1 { //Идём по каждому кругу и форматируем его в строку, если этот круг не был пройден - он форматируется в строку {,}
		if i < timesLen {
			laps = append(laps, formatLapInfo(times[i], speeds[i]))
		} else {
			laps = append(laps, formatLapInfo(0, 0))
		}

	}
	lapsInfo := strings.Join(laps, ", ") //Соединяем инфу о всех наших кругах через запятую
	return fmt.Sprintf("[%s]", lapsInfo)
}

func formatLapInfo(dur time.Duration, speed float64) string { //То же, что и функцией выше, но для одного круга, а не для всех
	if dur == 0 && speed == 0 {
		return "{,}"
	}
	return fmt.Sprintf("{%s, %.3f}", timeParser.ConvertDurationToString(dur), truncateFloatWithoutRounding(speed, 3))
}

func truncateFloatWithoutRounding(num float64, precision int) float64 { //Для вывода, как в примере в тз
	factor := math.Pow(10, float64(precision))
	return math.Trunc(num*factor) / factor
}

func computeAvgSpeed(dur time.Duration, distanceMeters float64) float64 { //Вычисляет скорость
	if dur > 0 {
		speedMps := distanceMeters / dur.Seconds()
		return speedMps
	}
	return 0
}

func (cm CompetitionManager) sortedCompetitors() []*Competitor { //Сортирует наших участников по времени
	competitors := make([]*Competitor, 0, len(cm.competitors))
	for _, c := range cm.competitors {
		competitors = append(competitors, c)
	}
	sort.Slice(competitors, func(i, j int) bool {
		return competitors[i].TotalTime < competitors[j].TotalTime
	})
	return competitors
}
