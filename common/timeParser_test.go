package timeParser

import (
	"testing"
	"time"
)

func TestConvertStringToTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "valid time",
			input:   "12:34:56.789",
			want:    time.Date(0, 1, 1, 12, 34, 56, 789000000, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "12:34:56.1000",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertStringToTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertStringToTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) && !tt.wantErr {
				t.Errorf("ConvertStringToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertStringToDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "valid duration",
			input:   "12:34:56",
			want:    12*time.Hour + 34*time.Minute + 56*time.Second,
			wantErr: false,
		},
		{
			name:    "zero duration",
			input:   "00:00:00",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "12:34",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "12.34.56",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertStringToDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertStringToDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want && !tt.wantErr {
				t.Errorf("ConvertStringToDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertDurationToString(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{
			name:  "hours, minutes, seconds, milliseconds",
			input: 12*time.Hour + 34*time.Minute + 56*time.Second + 789*time.Millisecond,
			want:  "12:34:56.789",
		},
		{
			name:  "zero duration",
			input: 0,
			want:  "00:00:00.000",
		},
		{
			name:  "only milliseconds",
			input: 250 * time.Millisecond,
			want:  "00:00:00.250",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertDurationToString(tt.input)
			if got != tt.want {
				t.Errorf("ConvertDurationToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
