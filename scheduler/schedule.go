package scheduler

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

const cronRegex = `^(\*|[0-5]?\d)(\/\d+)? (\*|[01]?\d|2[0-3])(\/\d+)? (\*|0?[1-9]|[12]\d|3[01])(\/\d+)? (\*|0?[1-9]|1[0-2])(\/\d+)? (\*|[0-6])(\/\d+)?$`

type timingType int

const (
	concrete timingType = iota
	step
	wildcard
)

const (
	minMax     = 60
	hourMax    = 24
	dayMax     = 31
	monthMax   = 12
	weekdayMax = 6
)

type timing struct {
	typ timingType
	val int
}

type schedule struct {
	min     timing
	hour    timing
	day     timing
	month   timing
	weekday timing
}

func getDurationTillNextProc(s schedule) time.Duration {
	currentDate := time.Now()

	nextMin := calcNextTime(s.min, currentDate.Minute(), minMax, 1)
	nextHour := calcNextTime(s.hour, currentDate.Hour(), hourMax, 0)
	nextDay := calcNextTime(s.day, currentDate.Day(), dayMax, 0)
	nextMonth := calcNextTime(s.month, int(currentDate.Month()), monthMax, 0)

	var nextYear int = currentDate.Year()
	if nextMonth < int(currentDate.Month()) {
		nextYear += 1
	}

	nextDate := time.Date(nextYear, time.Month(nextMonth), nextDay, nextHour, nextMin, 0, 0, currentDate.Location())
	return nextDate.Sub(currentDate)
}

func calcNextTime(t timing, currentTime, maxVal, wildCardIncrement int) int {
	if t.typ == wildcard {
		return currentTime + wildCardIncrement
	}

	if t.typ == step {
		stepped := min(currentTime+t.val, maxVal)
		return stepped - (stepped % min(t.val, maxVal))
	}

	if t.val < currentTime {
		return t.val + minMax
	}

	return t.val
}

func validateSchedule(schedule string) (bool, error) {
	ok, err := regexp.MatchString(cronRegex, schedule)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func parseSchedule(s string) schedule {
	timings := strings.Split(s, " ")

	min := convCronTiming(timings[0], 0, minMax)
	hour := convCronTiming(timings[1], 0, hourMax)
	day := convCronTiming(timings[2], 1, dayMax)
	month := convCronTiming(timings[3], 1, monthMax)
	weekday := convCronTiming(timings[4], 0, weekdayMax)

	return schedule{
		min:     min,
		hour:    hour,
		day:     day,
		month:   month,
		weekday: weekday,
	}
}

func convCronTiming(timeOption string, minVal, maxVal int) timing {
	if timeOption == "*" {
		return timing{
			typ: wildcard,
			val: minVal,
		}
	}

	var typ timingType
	if ok, _ := regexp.MatchString(`^\*\/\d+$`, timeOption); ok {
		timeOption = timeOption[2:]
		typ = step
	} else {
		typ = concrete
	}

	val, err := strconv.Atoi(timeOption)
	if err != nil {
		panic("String to int conversion should not have failed for cron string")
	}

	return timing{
		typ: typ,
		val: max(min(val, maxVal), minVal),
	}

}
