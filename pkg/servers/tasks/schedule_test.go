package tasks

import (
	"testing"
	"time"
)

func TestValidateSchedule(t *testing.T) {
	tests := []struct {
		name          string
		schedule      string
		hasExpiration bool
		wantErr       bool
	}{
		{name: "daily", schedule: "0 9 * * *"},
		{name: "weekly", schedule: "0 9 * * 1,3,5"},
		{name: "monthly", schedule: "0 9 1,15 * *"},
		{name: "one time", schedule: "0 9 20 3 *", hasExpiration: true},
		{name: "one time missing expiration", schedule: "0 9 20 3 *", wantErr: true},
		{name: "unsupported yearly-style rule", schedule: "0 9 * 3 *", wantErr: true},
		{name: "invalid cron", schedule: "not-a-cron", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchedule(tt.schedule, tt.hasExpiration)
			if tt.wantErr && err == nil {
				t.Fatal("validateSchedule() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateSchedule() error = %v", err)
			}
		})
	}
}

func TestNextRunAtRespectsTimezoneAndExpiration(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation() failed: %v", err)
	}

	t.Run("timezone", func(t *testing.T) {
		spec, err := cronParser.Parse("0 9 * * *")
		if err != nil {
			t.Fatalf("Parse() failed: %v", err)
		}

		from := time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC)
		got := nextRunAt(spec, location, nil, from)
		if got == nil {
			t.Fatal("nextRunAt() = nil, want next run")
		}

		want := time.Date(2026, 3, 20, 9, 0, 0, 0, location)
		if !got.Equal(want) {
			t.Fatalf("nextRunAt() = %v, want %v", got, want)
		}
	})

	t.Run("expiration", func(t *testing.T) {
		spec, err := cronParser.Parse("0 9 20 3 *")
		if err != nil {
			t.Fatalf("Parse() failed: %v", err)
		}

		expiresAt := time.Date(2026, 3, 20, 23, 59, 59, 0, location)

		before := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
		got := nextRunAt(spec, location, &expiresAt, before)
		if got == nil {
			t.Fatal("nextRunAt() = nil before expiration, want scheduled run")
		}

		after := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
		got = nextRunAt(spec, location, &expiresAt, after)
		if got != nil {
			t.Fatalf("nextRunAt() after expiration = %v, want nil", got)
		}
	})
}
