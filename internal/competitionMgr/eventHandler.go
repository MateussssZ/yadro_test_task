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

const TargetsPerFiringLine = 5

type CompetitionManager struct {
	outputFile  *os.File
	cfg         *cfg.Config
	competitors map[int]*Competitor
}

type Competitor struct {
	ReportInfo                 //Тут вся важная инфа, необходимая для построения final report
	StartTime        time.Time //Менее важные поля, необходимые для вычислений
	LastLapTime      time.Time
	LapsEnded        uint
	PenaltyLapsEnter time.Time
}

func NewCompetitionManager(outFile *os.File, cfg *cfg.Config) *CompetitionManager {
	return &CompetitionManager{
		outputFile:  outFile,
		cfg:         cfg,
		competitors: make(map[int]*Competitor, 0), //В качестве key будет выступать competitorId. Можно было бы обойтись слайсом, но нет уверенности,
		// что наши id идут по порядку(и не будет разрывов в номере участников)
	}
}

func (cm CompetitionManager) HandleEvent(eventInfo lh.EventInfo) error {
	switch eventInfo.EventId { //Если участник зарегался - создаём для него структуру и закидываем её в мапу
	case 1:
		cm.competitors[eventInfo.CompetitorId] = &Competitor{
			ReportInfo: ReportInfo{
				CompetitorId: eventInfo.CompetitorId,
				LapTimes:     make([]time.Duration, 0, cm.cfg.Laps),
				LapSpeeds:    make([]float64, 0, cm.cfg.Laps),
				Hits:         make([]bool, TargetsPerFiringLine*cm.cfg.FiringLines),
			},
		}
	case 2: //Если участник получил время, то считаем его как стартовое, т.к. в тз сказано "Total time includes the difference between scheduled and actual start time"
		competitor := cm.competitors[eventInfo.CompetitorId]
		startTime, err := timeParser.ConvertStringToTime(eventInfo.ExtraParams)
		if err != nil {
			return err
		}

		competitor.LastLapTime = startTime
		competitor.StartTime = startTime
	case 3, 5: //В целом ничего не требуется в этих случаях
		return nil
	case 4: //Если участние стартанул - надо посчитать, не опоздал ли он на старт, если опоздал - NotStarted статус, пусть подумает о поведении
		competitor := cm.competitors[eventInfo.CompetitorId]
		startDeltaDur, err := timeParser.ConvertStringToDuration(cm.cfg.StartDelta)
		if err != nil {
			return err
		}
		diff := eventInfo.EventTime.Sub(competitor.LastLapTime)

		if diff > startDeltaDur || diff < 0 {
			competitor.Status = "NotStarted"
			competitor.TotalTime = startDeltaDur //Вот тут не уверен, что нужно было именно такое время, может быть между запланированным и актуальным временем, но а если
			// он в целом не пришёл на старт?
		}
	case 6: //Просто обрабатываем, в какую мишень попал и сохраняем в мапу Hits, чтобы потом считать промахи/попадания
		competitor := cm.competitors[eventInfo.CompetitorId]
		targetNum, err := strconv.Atoi(eventInfo.ExtraParams)
		if err != nil {
			return fmt.Errorf("unable to convert targetNum to int(%s)", eventInfo.ExtraParams)
		}
		targetNum = TargetsPerFiringLine*competitor.FiringRangeNum + targetNum - 1
		maxFiringLines := TargetsPerFiringLine * cm.cfg.FiringLines
		if targetNum > maxFiringLines {
			return fmt.Errorf("target num can`t be more than %d(got %d)", maxFiringLines, targetNum)

		}
		competitor.Hits[targetNum] = true
	case 7: //Закончил стрельбище - сохраним
		competitor := cm.competitors[eventInfo.CompetitorId]
		competitor.FiringRangeNum += 1
	case 8: //Забежал на штрафные - запомним
		competitor := cm.competitors[eventInfo.CompetitorId]
		competitor.PenaltyLapsEnter = eventInfo.EventTime
	case 9: //Выбежал со штрафных - посчитаем время, чтобы потом в final report отправить
		competitor := cm.competitors[eventInfo.CompetitorId]
		time := eventInfo.EventTime.Sub(competitor.PenaltyLapsEnter)

		competitor.PenaltyTime += time
	case 10: //Закончил круг - посчитаем время круга, скорость. Если круг был последним - зафиксируем итоговый результат и статус Finished
		competitor := cm.competitors[eventInfo.CompetitorId]
		time, speed := calculateLapStats(competitor.LastLapTime, eventInfo.EventTime, float64(cm.cfg.LapLen))
		competitor.LapTimes = append(competitor.LapTimes, time)
		competitor.LapSpeeds = append(competitor.LapSpeeds, speed)
		competitor.LastLapTime = eventInfo.EventTime

		if competitor.LapsEnded == uint(cm.cfg.Laps)-1 {
			competitor.Status = "Finished"
			competitor.TotalTime = eventInfo.EventTime.Sub(competitor.StartTime)
		} else if competitor.LapsEnded == uint(cm.cfg.Laps) {
			return fmt.Errorf("competitor ended more laps than needed")
		}
		competitor.LapsEnded += 1
	case 11: //Ну тут просто обрабатываем, что человек не закончил гонку(статус и общее время)
		competitor := cm.competitors[eventInfo.CompetitorId]

		competitor.Status = "NotFinished"
		competitor.TotalTime = eventInfo.EventTime.Sub(competitor.StartTime)
	}
	return nil
}

func countHits(hits []bool) int {
	count := 0
	for _, v := range hits {
		if v {
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
