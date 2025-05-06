package main

import (
	"bufio"
	"log"
	"os"

	"yadro_test/internal/cfg"
	cmptmgr "yadro_test/internal/competitionMgr"
	cl "yadro_test/internal/logger"
)

func main() { //Я не фанат комментариев и считаю, что код в go вполне себе самодокументируем, но мне посоветовали написать комментарии в тестовом, поэтому пишу
	logFile, err := os.OpenFile("output.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) //Открываем файл для логов
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	outFile, err := os.OpenFile("output.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) //Файл для финального репорта
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	inputFile, err := os.Open("events") //Инпут файл
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	cfg := cfg.MustLoad()                                  //Конфиг
	l := cl.NewCustomLogger(logFile)                       //Будет закидывать кастомные логи в файл
	cmptMgr := cmptmgr.NewCompetitionManager(outFile, cfg) //Отвечает за бизнес-логику и обработку событий(эвентов)

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() { // Идём по каждой строчке и передаём её в логгер
		line := scanner.Text()
		eventInfo, err := l.ProcessLine(line)
		if err != nil {
			log.Fatalf("LogHandler(ProcessLine) error: %v", err)
		}
		err = cmptMgr.HandleEvent(eventInfo) //Затем после обработки лога обрабатываем её менеджером
		if err != nil {
			log.Fatalf("CompetitorManager(HandleEvent) error: %v", err)
		}
	}
	err = cmptMgr.GenerateReport() //Когда мы прошли все строчки инпут файла - генерируем final report, на это работа программы закончена
	if err != nil {
		log.Fatalf("CompetitorManager(GenerateReport) error: %v", err)
	}
}
