package loghandler

import (
	"bytes"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNewCustomLogger(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testlog")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	logger := NewCustomLogger(tmpfile)
	if logger == nil {
		t.Error("NewCustomLogger() returned nil")
	}
}

func TestProcessLineValidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected EventInfo
		wantErr  bool
	}{
		{
			name:  "valid line (event 1)",
			input: "[12:34:56.789] 1 100",
			expected: EventInfo{
				EventId:      1,
				CompetitorId: 100,
				EventTime:    time.Date(0, 1, 1, 12, 34, 56, 789000000, time.UTC),
				ExtraParams:  "",
			},
			wantErr: false,
		},
		{
			name:  "valid line with extra params (event 2)",
			input: "[12:34:56.789] 2 100 13:00:00.000",
			expected: EventInfo{
				EventId:      2,
				CompetitorId: 100,
				EventTime:    time.Date(0, 1, 1, 12, 34, 56, 789000000, time.UTC),
				ExtraParams:  "13:00:00.000",
			},
			wantErr: false,
		},
		{
			name:  "valid line (event 6)",
			input: "[12:34:56.789] 6 100 1",
			expected: EventInfo{
				EventId:      6,
				CompetitorId: 100,
				EventTime:    time.Date(0, 1, 1, 12, 34, 56, 789000000, time.UTC),
				ExtraParams:  "1",
			},
			wantErr: false,
		},
	}

	buf := new(bytes.Buffer)
	logger := &CustomLogger{
		l: slog.New(slog.NewTextHandler(buf, nil)),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := logger.ProcessLine(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.EventId != tt.expected.EventId {
				t.Errorf("ProcessLine() EventId = %v, want %v", got.EventId, tt.expected.EventId)
			}
			if got.CompetitorId != tt.expected.CompetitorId {
				t.Errorf("ProcessLine() CompetitorId = %v, want %v", got.CompetitorId, tt.expected.CompetitorId)
			}
			if !got.EventTime.Equal(tt.expected.EventTime) {
				t.Errorf("ProcessLine() EventTime = %v, want %v", got.EventTime, tt.expected.EventTime)
			}
			if got.ExtraParams != tt.expected.ExtraParams {
				t.Errorf("ProcessLine() ExtraParams = %v, want %v", got.ExtraParams, tt.expected.ExtraParams)
			}
		})
	}
}

func TestProcessLineInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty line",
			input:   "",
			wantErr: "insufficient number of parameters in line ()",
		},
		{
			name:    "num of params after time less than 2",
			input:   "[12:34:56.789] 1",
			wantErr: "insufficient number of parameters in line ([12:34:56.789] 1)",
		},
		{
			name:    "missing competitorId",
			input:   "[12:34:56.789] 1",
			wantErr: "insufficient number of parameters in line ([12:34:56.789] 1)",
		},
		{
			name:    "invalid eventId",
			input:   "[12:34:56.789] abc 100",
			wantErr: "can`t convert eventId(abc) to int",
		},
		{
			name:    "invalid competitorId",
			input:   "[12:34:56.789] 1 abc",
			wantErr: "can`t convert competitorId(abc) to int",
		},
		{
			name:    "invalid time format",
			input:   "[12:34:56] 1 100",
			wantErr: "unable to parse time.Time(12:34:56)",
		},
	}

	logger := &CustomLogger{
		l: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := logger.ProcessLine(tt.input)
			if err == nil {
				t.Errorf("ProcessLine() expected error, got nil")
				return
			}
			if err.Error() != tt.wantErr {
				t.Errorf("ProcessLine() error = %v, wantErr %v", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestBuildLogMessage(t *testing.T) {
	tests := []struct {
		name         string
		time         string
		competitorId int
		eventId      int
		extraParams  string
		expectedMsg  string
	}{
		{
			name:         "event 1 (no extra params)",
			time:         "[12:34:56.789]",
			competitorId: 100,
			eventId:      1,
			extraParams:  "",
			expectedMsg:  "[12:34:56.789] The competitor(100) registered",
		},
		{
			name:         "event 2 (with extra params)",
			time:         "[12:34:56.789]",
			competitorId: 100,
			eventId:      2,
			extraParams:  "13:00:00.000",
			expectedMsg:  "[12:34:56.789] The start time for the competitor(100) was set by a draw to 13:00:00.000",
		},
		{
			name:         "event 6 (reversed params)",
			time:         "[12:34:56.789]",
			competitorId: 100,
			eventId:      6,
			extraParams:  "targetA",
			expectedMsg:  "[12:34:56.789] The target(targetA) has been hit by competitor(100)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildLogMessage(tt.time, tt.competitorId, tt.eventId, tt.extraParams)
			if got != tt.expectedMsg {
				t.Errorf("buildLogMessage() = %v, want %v", got, tt.expectedMsg)
			}
		})
	}
}
