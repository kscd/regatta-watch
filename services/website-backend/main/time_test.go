package main

import (
	"testing"
	"time"
)

// MockTimeSource implements timeSource for testing
type MockTimeSource struct {
	currentTime time.Time
}

func (m *MockTimeSource) Now() time.Time {
	return m.currentTime
}

func (m *MockTimeSource) Since(t time.Time) time.Duration {
	return m.currentTime.Sub(t)
}

func (m *MockTimeSource) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

func TestClock_Now(t *testing.T) {
	referenceRealTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		speed           float64
		referenceTime   time.Time
		advanceRealTime time.Duration
		expectedTime    time.Time
	}{
		{
			name:            "No offset, normal speed",
			speed:           1.0,
			referenceTime:   referenceRealTime,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(10 * time.Second),
		},
		{
			name:            "Reference time in the past (go back in time), normal speed",
			referenceTime:   referenceRealTime.Add(-30 * time.Second),
			speed:           1.0,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(10 * time.Second).Add(-30 * time.Second),
		},
		{
			name:            "Reference time in the future (go forward in time), normal speed",
			referenceTime:   referenceRealTime.Add(30 * time.Second),
			speed:           1.0,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(10 * time.Second).Add(30 * time.Second),
		},
		{
			name:            "No offset, double speed",
			referenceTime:   referenceRealTime,
			speed:           2.0,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(20 * time.Second),
		},
		{
			name:            "No offset, half speed",
			referenceTime:   referenceRealTime,
			speed:           0.5,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(5 * time.Second),
		},
		{
			name:            "Reference time in the past, double speed",
			referenceTime:   referenceRealTime.Add(-30 * time.Second),
			speed:           2.0,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(20 * time.Second).Add(-30 * time.Second),
		},
		{
			name:            "Edge case: Zero speed",
			referenceTime:   referenceRealTime,
			speed:           0,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime, // Time should not advance
		},
		{
			name:            "Edge case: Negative speed",
			referenceTime:   referenceRealTime,
			speed:           -1.0,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(-10 * time.Second), // Time goes backward
		},
		{
			name:            "Edge case: Large offset",
			referenceTime:   referenceRealTime.Add(-8760 * time.Hour), // 1 year
			speed:           1.0,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(10 * time.Second).Add(-8760 * time.Hour),
		},
		{
			name:            "Edge case: Very large speed",
			referenceTime:   referenceRealTime,
			speed:           1000.0,
			advanceRealTime: 10 * time.Second,
			expectedTime:    referenceRealTime.Add(10000 * time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTime := &MockTimeSource{
				currentTime: referenceRealTime,
			}

			c := &clock{
				timeSource:        mockTime,
				referenceTime:     tt.referenceTime,
				referenceRealTime: referenceRealTime,
				speed:             tt.speed,
			}

			// Advance time
			mockTime.Advance(tt.advanceRealTime)

			// Get the current time from clock
			got := c.Now()

			if !got.Equal(tt.expectedTime) {
				t.Errorf("Clock.Now() = %v, want %v", got, tt.expectedTime)
			}
		})
	}
}
