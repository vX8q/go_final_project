package api

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidRepeatFormat = errors.New("invalid repeat format")
	ErrInvalidDate         = errors.New("invalid date")
	ErrInvalidDays         = errors.New("invalid days (1-400)")
	ErrUnsupportedRule     = errors.New("unsupported rule")
)

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", ErrInvalidRepeatFormat
	}
	start, err := time.Parse(DateLayout, dstart)
	if err != nil {
		return "", ErrInvalidDate
	}
	parts := strings.Fields(repeat)
	if len(parts) < 1 {
		return "", ErrInvalidRepeatFormat
	}
	switch parts[0] {
	case "d":
		return handleDailyRule(now, start, parts)
	case "y":
		return handleYearlyRule(now, start)
	default:
		return "", ErrUnsupportedRule
	}
}

func handleDailyRule(now, start time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", ErrInvalidRepeatFormat
	}
	days, err := strconv.Atoi(strings.Trim(parts[1], "+ "))
	if err != nil || days < 1 || days > 400 {
		return "", ErrInvalidDays
	}
	date := start
	for {
		date = date.AddDate(0, 0, days)
		if isAfter(date, now) {
			return date.Format(DateLayout), nil
		}
	}
}

func handleYearlyRule(now, start time.Time) (string, error) {
	date := start
	for {
		var next time.Time

		if date.Month() == time.February && date.Day() == 29 {
			y := date.Year() + 1
			if isLeap(y) {
				next = time.Date(y, time.February, 29,
					date.Hour(), date.Minute(), date.Second(), date.Nanosecond(),
					date.Location())
			} else {
				next = time.Date(y, time.March, 1,
					date.Hour(), date.Minute(), date.Second(), date.Nanosecond(),
					date.Location())
			}
		} else {

			next = date.AddDate(1, 0, 0)
		}
		date = next
		if isAfter(date, now) {
			break
		}
	}
	return date.Format(DateLayout), nil
}

func isLeap(year int) bool {
	return year%400 == 0 || (year%4 == 0 && year%100 != 0)
}

func isAfter(date, now time.Time) bool {
	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	n := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return !d.Before(n)
}
