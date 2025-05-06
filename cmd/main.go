package main

import (
	"bufio"
	"log"
	"os"

	"yadro_test/internal/cfg"
	cmptmgr "yadro_test/internal/competitionMgr"
	cl "yadro_test/internal/logger"
)

func main() {
	logFile, err := os.OpenFile("output.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	outFile, err := os.OpenFile("output.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	inputFile, err := os.Open("events")
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	cfg := cfg.MustLoad()
	l := cl.NewCustomLogger(logFile)
	cmptMgr := cmptmgr.NewCompetitionManager(outFile, cfg)

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		eventInfo, err := l.ProcessLine(line)
		if err != nil {
			log.Fatalf("LogHandler(ProcessLine) error: %v", err)
		}
		err = cmptMgr.HandleEvent(eventInfo)
		if err != nil {
			log.Fatalf("CompetitorManager(HandleEvent) error: %v", err)
		}
	}
	err = cmptMgr.GenerateReport()
	if err != nil {
		log.Fatalf("CompetitorManager(GenerateReport) error: %v", err)
	}
}
