package tasks

import (
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// parseSchedule parses and validates the cron expression and timezone.
func parseSchedule(schedule, timezone string) (cron.Schedule, *time.Location, error) {
	spec, err := cronParser.Parse(schedule)
	if err != nil {
		return nil, nil, mcp.ErrRPCInvalidParams.WithMessage("invalid schedule: %s", err)
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, nil, mcp.ErrRPCInvalidParams.WithMessage("invalid timezone: %s", timezone)
	}
	return spec, loc, nil
}

// parseExpiration parses a YYYY-MM-DD expiration string into end-of-day in the
// given location. Returns nil if expiration is empty.
func parseExpiration(expiration string, loc *time.Location) (*time.Time, error) {
	expiration = strings.TrimSpace(expiration)
	if expiration == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation(time.DateOnly, expiration, loc)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("expiration must be YYYY-MM-DD")
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), loc)
	return &t, nil
}

func validateSchedule(schedule string, hasExpiration bool) error {
	fields := strings.Fields(schedule)
	if len(fields) != 5 {
		return mcp.ErrRPCInvalidParams.WithMessage("schedule must be a five-field cron expression")
	}
	dayOfMonth, month, dayOfWeek := fields[2], fields[3], fields[4]
	switch {
	case dayOfMonth == "*" && month == "*" && dayOfWeek == "*":
	case dayOfMonth == "*" && month == "*" && dayOfWeek != "*":
	case dayOfMonth != "*" && month == "*" && dayOfWeek == "*":
	case dayOfMonth != "*" && month != "*" && dayOfWeek == "*" && hasExpiration:
	default:
		return mcp.ErrRPCInvalidParams.WithMessage("schedule must be daily, weekly, monthly, or a single date with expiration")
	}
	return nil
}

// nextRunAt computes the next fire time from a pre-parsed cron spec in the
// given location.
func nextRunAt(spec cron.Schedule, loc *time.Location, expiresAt *time.Time, after time.Time) *time.Time {
	next := spec.Next(after.In(loc))
	if next.IsZero() || (expiresAt != nil && next.After(*expiresAt)) {
		return nil
	}
	return &next
}
