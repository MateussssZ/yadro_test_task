package competitionmgr

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"yadro_test/internal/cfg"
	lh "yadro_test/internal/logger"
)

func TestHandleEventRegistration(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testoutput")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	cfg := &cfg.Config{FiringLines: 2}
	cm := NewCompetitionManager(tmpfile, cfg)

	// Event 1: Registration
	err = cm.HandleEvent(lh.EventInfo{
		EventId:      1,
		CompetitorId: 100,
		EventTime:    time.Now(),
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	if len(cm.competitors) != 1 {
		t.Fatalf("Expected 1 competitor, got %d", len(cm.competitors))
	}
	if cm.competitors[100] == nil {
		t.Fatal("Competitor 100 not found")
	}
	if cm.competitors[100].CompetitorId != 100 {
		t.Errorf("Expected CompetitorId 100, got %d", cm.competitors[100].CompetitorId)
	}
}

func TestHandleEventStartTime(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testoutput")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	cfg := &cfg.Config{FiringLines: 2}
	cm := NewCompetitionManager(tmpfile, cfg)

	err = cm.HandleEvent(lh.EventInfo{EventId: 1, CompetitorId: 100})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	startTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	err = cm.HandleEvent(lh.EventInfo{
		EventId:      2,
		CompetitorId: 100,
		ExtraParams:  "12:00:00.000",
		EventTime:    startTime,
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	fmt.Println(cm.competitors[100].StartTime)
	time := cm.competitors[100].StartTime.Format("15:04:05.000")
	if time != "12:00:00.000" {
		t.Errorf("HandleEvent() expected StartTime 12:00:00.000, got %v", time)
	}

	err = cm.HandleEvent(lh.EventInfo{
		EventId:      2,
		CompetitorId: 100,
		ExtraParams:  "61:61:61.000",
		EventTime:    startTime,
	})
	if err == nil {
		t.Fatalf("HandleEvent expected wrong extraParams time, but err=nil")
	}
}

func TestHandleEventLapCompletion(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testoutput")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	cfg := &cfg.Config{
		Laps:       1,
		LapLen:     1000,
		StartDelta: "00:03:00",
	}
	cm := NewCompetitionManager(tmpfile, cfg)

	startTime := time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)
	err = cm.HandleEvent(lh.EventInfo{EventId: 1, CompetitorId: 100})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}
	err = cm.HandleEvent(lh.EventInfo{
		EventId:      2,
		CompetitorId: 100,
		ExtraParams:  "12:01:00.000",
		EventTime:    startTime,
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}
	err = cm.HandleEvent(lh.EventInfo{
		EventId:      4,
		CompetitorId: 100,
		EventTime:    startTime.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	lap1End := startTime.Add(7 * time.Minute)
	err = cm.HandleEvent(lh.EventInfo{
		EventId:      10,
		CompetitorId: 100,
		EventTime:    lap1End,
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	if len(cm.competitors[100].LapTimes) != 1 {
		t.Fatalf("Expected 1 lap time, got %d", len(cm.competitors[100].LapTimes))
	}
	if cm.competitors[100].LapTimes[0] != 6*time.Minute {
		t.Errorf("Expected lap time 6m, got %v", cm.competitors[100].LapTimes[0])
	}
	if cm.competitors[100].LapsEnded != 1 {
		t.Errorf("Expected 1 lap ended, got %d", cm.competitors[100].LapsEnded)
	}
	if cm.competitors[100].Status != "Finished" {
		t.Errorf("Expected status Finished, got %s", cm.competitors[100].Status)
	}
	if cm.competitors[100].TotalTime != 6*time.Minute {
		t.Errorf("Expected total time 6m, got %v", cm.competitors[100].TotalTime)
	}
}

func TestHandleEventCompetitionFailing(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testoutput")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	cfg := &cfg.Config{
		Laps:       1,
		LapLen:     1000,
		StartDelta: "00:02:00",
	}
	cm := NewCompetitionManager(tmpfile, cfg)

	startTime := time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)
	err = cm.HandleEvent(lh.EventInfo{EventId: 1, CompetitorId: 100})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}
	err = cm.HandleEvent(lh.EventInfo{
		EventId:      2,
		CompetitorId: 100,
		ExtraParams:  "12:01:00.000",
		EventTime:    startTime,
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}
	err = cm.HandleEvent(lh.EventInfo{
		EventId:      3,
		CompetitorId: 100,
		EventTime:    startTime.Add(30 * time.Second),
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}
	err = cm.HandleEvent(lh.EventInfo{
		EventId:      4,
		CompetitorId: 100,
		EventTime:    startTime.Add(1 * time.Minute),
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	err = cm.HandleEvent(lh.EventInfo{
		EventId:      11,
		CompetitorId: 100,
		ExtraParams:  "Lost",
		EventTime:    startTime.Add(7 * time.Minute),
	})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	if len(cm.competitors[100].LapTimes) != 0 {
		t.Fatalf("Expected 0 lap times, got %d", len(cm.competitors[100].LapTimes))
	}
	if cm.competitors[100].LapsEnded != 0 {
		t.Errorf("Expected 0 lap ended, got %d", cm.competitors[100].LapsEnded)
	}
	if cm.competitors[100].Status != "NotFinished" {
		t.Errorf("Expected status NotFinished, got %s", cm.competitors[100].Status)
	}
	if cm.competitors[100].TotalTime != 6*time.Minute {
		t.Errorf("Expected total time 6m, got %v", cm.competitors[100].TotalTime)
	}
}

func TestComputeAvgSpeed(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		distance float64
		expected float64
	}{
		{
			name:     "5 minutes for 1km",
			duration: 5 * time.Minute,
			distance: 1000,
			expected: 3.333333,
		},
		{
			name:     "10 minutes for 2km",
			duration: 10 * time.Minute,
			distance: 2000,
			expected: 3.333333,
		},
		{
			name:     "10 minutes for 1.2km",
			duration: 10 * time.Minute,
			distance: 1200,
			expected: 2,
		},
		{
			name:     "zero duration",
			duration: 0,
			distance: 1000,
			expected: 0,
		},
		{
			name:     "zero distance",
			duration: 5 * time.Minute,
			distance: 0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			speed := computeAvgSpeed(tt.duration, tt.distance)
			eps := 0.01
			if math.Abs(speed-tt.expected) > eps {
				t.Errorf("ComputeAvgSpeed() expected=%f, got %f", tt.expected, speed)
			}
		})
	}
}

func TestCountHits(t *testing.T) {
	tests := []struct {
		name     string
		hits     []bool
		expected int
	}{
		{
			name:     "random slice",
			hits:     []bool{false, false, true, false, true},
			expected: 2,
		},
		{
			name:     "empty slice",
			hits:     []bool{},
			expected: 0,
		},
		{
			name:     "only true",
			hits:     []bool{true, true, true, true, true},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := countHits(tt.hits)
			if ans != tt.expected {
				t.Errorf("countMisses() expected=%d, got %d", tt.expected, ans)
			}
		})
	}
}
