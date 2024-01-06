package util

import (
	"time"
)

type TimeRange struct {
	Daily string `json:"daily"`
	Begin string `json:"begin"`
	End   string `json:"end"`
}

func IsWithinRange(TimeRange TimeRange) (bool, error) {
	currentTime := time.Now()
	if TimeRange.Daily == "daily" {
		begin, err := time.Parse("15:04", TimeRange.Begin)
		if err != nil {
			return false, err
		}
		end, err := time.Parse("15:04", TimeRange.End)
		if err != nil {
			return false, err
		}
		startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), begin.Hour(), begin.Minute(), 0, 0, time.Local)
		endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), end.Hour(), end.Minute(), 0, 0, time.Local)
		return currentTime.After(startTime) && currentTime.Before(endTime), nil
	} else if TimeRange.Daily == "everyWorKDay" {
		if currentTime.Weekday() == time.Saturday || currentTime.Weekday() == time.Sunday {
			return false, nil
		}
		begin, err := time.Parse("15:04", TimeRange.Begin)
		if err != nil {
			return false, err
		}
		end, err := time.Parse("15:04", TimeRange.End)
		if err != nil {
			return false, err
		}
		startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), begin.Hour(), begin.Minute(), 0, 0, time.Local)
		endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), end.Hour(), end.Minute(), 0, 0, time.Local)
		return currentTime.After(startTime) && currentTime.Before(endTime), nil
	} else if TimeRange.Daily == "OpeningtimeBroad" {
		if currentTime.Weekday() >= time.Monday && currentTime.Weekday() <= time.Friday {
			return true, nil
		} else if currentTime.Weekday() == time.Saturday && currentTime.Hour() < 5 {
			return true, nil
		} else if currentTime.Weekday() == time.Saturday && currentTime.Hour() >= 5 {
			return false, nil
		} else {
			return false, nil
		}
	}
	return false, nil
}
