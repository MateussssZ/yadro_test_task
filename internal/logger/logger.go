package loghandler

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
	timeParser "yadro_test/common"
)

var mapEvents = map[int]string{ //Для логов хардкодим строчки
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

const numReqParams = 3

type EventInfo struct {
	EventId      int
	CompetitorId int
	ExtraParams  string //Здесь будет храниться либо время либо номер стрельбища, цели и тд
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

func (cl CustomLogger) ProcessLine(line string) (EventInfo, error) {
	line = strings.TrimSpace(line)    //Обрезаем по бокам лишние пробелы на всякий случай
	parts := strings.Split(line, " ") //Разбиваем на части и обрабатываем случай, если их меньше 3(time eventId compId)
	if len(parts) < numReqParams {
		return EventInfo{}, fmt.Errorf("insufficient number of parameters in line (%s)", line)
	}

	time, eventIdStr, competitorIdStr := parts[0], parts[1], parts[2] //Собираем наши параметры и конвертим их в удобные для работы типы данных
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		return EventInfo{}, fmt.Errorf("can`t convert eventId(%s) to int", eventIdStr)
	}
	competitorId, err := strconv.Atoi(competitorIdStr)
	if err != nil {
		return EventInfo{}, fmt.Errorf("can`t convert competitorId(%s) to int", competitorIdStr)
	}
	var extraParams string //Если есть ещё какие-то параметры - забираем их как строчку
	if len(parts) > numReqParams {
		extraParams = strings.Join(parts[numReqParams:], " ")
	}

	msg := buildLogMessage(time, competitorId, eventId, extraParams) //Построение итоговой строки
	cl.l.Info(msg)                                                   //Запись в лог-файл

	eventTime, err := timeParser.ConvertStringToTime(strings.Trim(time, "[]")) //Перевод в удобный тип(time.Time) для работы, у строки по типу [12:00:00.000] обрезаем скобки парсим на время
	if err != nil {
		return EventInfo{}, err
	}
	return EventInfo{ //Полезная структура для работы менеджера в будущем
		EventId:      eventId,
		CompetitorId: competitorId,
		EventTime:    eventTime,
		ExtraParams:  extraParams,
	}, nil
}

func buildLogMessage(time string, competitorId, eventId int, extraParams string) string {
	var eventMsg string
	switch eventId { //В зависимости от типа события разные параметры передаём(они в разном порядке и количестве идут)
	case 2, 5, 11:
		eventMsg = fmt.Sprintf(mapEvents[eventId], competitorId, extraParams)
	case 6:
		eventMsg = fmt.Sprintf(mapEvents[eventId], extraParams, competitorId)
	default:
		eventMsg = fmt.Sprintf(mapEvents[eventId], competitorId)
	}
	return fmt.Sprintf("%s %s", time, eventMsg)
}
