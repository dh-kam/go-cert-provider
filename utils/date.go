package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DateTimeFormat defines the standard date/time format used throughout the project
const DateTimeFormat = "2006-01-02 15:04:05"

// FormatDateTime formats the given time to standard format in local timezone
func FormatDateTime(t time.Time) string {
	return t.Local().Format(DateTimeFormat)
}

func FormatCurrentTime() string {
	return FormatDateTime(time.Now())
}

// ParseDateTime parses YYYY-MM-DD HH:mm:ss format string to time in local timezone
func ParseDateTime(timeStr string) (time.Time, error) {
	return time.ParseInLocation(DateTimeFormat, timeStr, time.Local)
}

// FormatDuration formats duration to human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Hour).String()
	}
	totalHours := d.Hours()
	days := int(totalHours / 24)
	remainingHours := int(totalHours) - (days * 24)
	if days > 0 && remainingHours > 0 {
		return fmt.Sprintf("%d days %d hours", days, remainingHours)
	}
	if days > 0 {
		return fmt.Sprintf("%d days", days)
	}
	return d.Round(time.Hour).String()
}

// ParseDurationString parses duration strings with extended support for years, months, weeks
// Supports: y/year/years, M/month/months, w/week/weeks, d/day/days, h/hour/hours, m/minute/minutes, s/second/seconds
// Examples: "2y", "3months", "5d", "10w", "2h30m"
func ParseDurationString(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Try standard Go duration first (supports h, m, s, ms, us, ns)
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Parse custom duration units
	var total time.Duration
	remaining := s

	y := 365 * 24 * time.Hour
	m := 30 * 24 * time.Hour
	w := 7 * 24 * time.Hour
	d := 24 * time.Hour

	units := map[string]time.Duration{
		"y":      y,
		"year":   y,
		"years":  y,
		"m":      m,
		"month":  m,
		"months": m,
		"w":      w,
		"week":   w,
		"weeks":  w,
		"d":      d,
		"day":    d,
		"days":   d,
	}

	// Parse number + unit combinations
	i := 0
	for i < len(remaining) {
		// Find the start of a number
		start := i
		for i < len(remaining) && (remaining[i] >= '0' && remaining[i] <= '9' || remaining[i] == '.') {
			i++
		}

		if i == start {
			return 0, fmt.Errorf("invalid duration format: %s", s)
		}

		numStr := remaining[start:i]
		var num float64
		var err error
		if strings.Contains(numStr, ".") {
			num, err = strconv.ParseFloat(numStr, 64)
		} else {
			var intNum int64
			intNum, err = strconv.ParseInt(numStr, 10, 64)
			num = float64(intNum)
		}
		if err != nil {
			return 0, fmt.Errorf("invalid number in duration: %s", numStr)
		}

		unitStart := i
		for i < len(remaining) && (remaining[i] < '0' || remaining[i] > '9') && remaining[i] != '.' {
			i++
		}

		unitStr := remaining[unitStart:i]
		if unitStr == "" {
			return 0, fmt.Errorf("missing unit after number %s", numStr)
		}

		unitDuration, exists := units[strings.ToLower(unitStr)]
		if !exists {
			return 0, fmt.Errorf("unknown duration unit: %s", unitStr)
		}

		total += time.Duration(float64(unitDuration) * num)
	}

	if total == 0 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	return total, nil
}
