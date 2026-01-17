package utils

import (
	"testing"
	"time"
)

func TestFormatDateTime(t *testing.T) {
	// Use a fixed time for testing
	fixedTime := time.Date(2025, 8, 22, 15, 30, 45, 0, time.Local)

	formatted := FormatDateTime(fixedTime)
	expected := "2025-08-22 15:30:45"

	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}
}

func TestParseDateTime(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		shouldOK bool
	}{
		{"valid datetime", "2025-08-22 15:30:45", true},
		{"valid datetime with zeros", "2025-01-01 00:00:00", true},
		{"invalid format missing time", "2025-08-22", false},
		{"invalid format", "2025/08/22 15:30:45", false},
		{"empty string", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseDateTime(tc.input)

			if tc.shouldOK {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				// Verify round-trip
				formatted := FormatDateTime(parsed)
				if formatted != tc.input {
					t.Errorf("Round-trip failed: expected %s, got %s", tc.input, formatted)
				}
			} else {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			}
		})
	}
}

func TestFormatCurrentTime(t *testing.T) {
	before := time.Now()
	formatted := FormatCurrentTime()
	after := time.Now()

	// Parse back
	parsed, err := ParseDateTime(formatted)
	if err != nil {
		t.Fatalf("Failed to parse formatted time: %v", err)
	}

	// Should be between before and after
	if parsed.Before(before.Add(-time.Second)) || parsed.After(after.Add(time.Second)) {
		t.Errorf("Formatted time %v is not within expected range [%v, %v]", parsed, before, after)
	}
}

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"seconds", 45 * time.Second, "45s"},
		{"minutes", 30 * time.Minute, "30m0s"},
		{"hours", 5 * time.Hour, "5h0m0s"},
		{"days only", 3 * 24 * time.Hour, "3 days"},
		{"days and hours", 3*24*time.Hour + 5*time.Hour, "3 days 5 hours"},
		{"1 day 1 hour", 25 * time.Hour, "1 days 1 hours"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatDuration(tc.duration)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestParseDurationString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Duration
		shouldOK bool
	}{
		// Standard Go duration formats
		{"hours", "2h", 2 * time.Hour, true},
		{"minutes", "30m", 30 * time.Minute, true},
		{"seconds", "45s", 45 * time.Second, true},
		{"combined", "2h30m", 2*time.Hour + 30*time.Minute, true},

		// Extended formats
		{"days", "5d", 5 * 24 * time.Hour, true},
		{"days long", "5day", 5 * 24 * time.Hour, true},
		{"days plural", "5days", 5 * 24 * time.Hour, true},
		{"weeks", "2w", 2 * 7 * 24 * time.Hour, true},
		{"weeks long", "2week", 2 * 7 * 24 * time.Hour, true},
		{"weeks plural", "2weeks", 2 * 7 * 24 * time.Hour, true},
		{"months", "3M", 3 * 30 * 24 * time.Hour, true},
		{"months long", "3month", 3 * 30 * 24 * time.Hour, true},
		{"months plural", "3months", 3 * 30 * 24 * time.Hour, true},
		{"years", "1y", 365 * 24 * time.Hour, true},
		{"years long", "1year", 365 * 24 * time.Hour, true},
		{"years plural", "2years", 2 * 365 * 24 * time.Hour, true},

		// Decimal values
		{"decimal days", "1.5d", 36 * time.Hour, true},
		{"decimal weeks", "0.5w", 84 * time.Hour, true},

		// Invalid formats
		{"empty", "", 0, false},
		{"no unit", "123", 0, false},
		{"invalid unit", "5x", 0, false},
		{"space separated", "5 days", 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDurationString(tc.input)

			if tc.shouldOK {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			} else {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			}
		})
	}
}

func TestParseDurationString_ComplexCombinations(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{"year and month", "1y3months", 365*24*time.Hour + 3*30*24*time.Hour},
		{"week and days", "1w2d", 7*24*time.Hour + 2*24*time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDurationString(tc.input)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v (diff: %v)", tc.expected, result, result-tc.expected)
			}
		})
	}
}

func TestDateTimeFormat_Constant(t *testing.T) {
	// Verify the format constant is correct for Go's time parsing
	expected := "2006-01-02 15:04:05"
	if DateTimeFormat != expected {
		t.Errorf("DateTimeFormat constant should be %s, got %s", expected, DateTimeFormat)
	}
}

func TestFormatDateTime_RoundTrip(t *testing.T) {
	// Test that formatting and parsing are inverse operations
	original := time.Date(2025, 12, 31, 23, 59, 59, 0, time.Local)

	formatted := FormatDateTime(original)
	parsed, err := ParseDateTime(formatted)
	if err != nil {
		t.Fatalf("Failed to parse formatted time: %v", err)
	}

	if !original.Equal(parsed) {
		t.Errorf("Round-trip failed: original %v != parsed %v", original, parsed)
	}
}

func TestParseDurationString_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		shouldOK bool
	}{
		{"zero value", "0s", true},
		{"very large", "999y", true},
		{"lowercase", "5d", true},
		{"uppercase M (should work as month)", "5M", true},
		{"multiple same units", "1d1d", true}, // Should add up
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDurationString(tc.input)

			if tc.shouldOK && err != nil {
				t.Errorf("Expected success, got error: %v", err)
			} else if !tc.shouldOK && err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}
