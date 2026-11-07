package util

import (
	"regexp"
	"slices"
	"strconv"
)

func ValidDate(date string) (bool, error) {
	if len(date) != 10 {
		return false, nil
	}
	valid, err := regexp.MatchString(`\d\d.\d\d.\d\d\d\d`, date)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, nil
	}
	dd, err := strconv.Atoi(date[:2])
	if err != nil {
		return false, err
	}
	mm, err := strconv.Atoi(date[3:5])
	if err != nil {
		return false, err
	}
	yyyy, err := strconv.Atoi(date[6:])
	if err != nil {
		return false, err
	}
	if dd < 1 {
		return false, nil
	}
	if slices.Contains([]int{4, 6, 9, 11}, mm) && dd > 30 {
		return false, nil
	}
	if slices.Contains([]int{1, 3, 5, 7, 8, 10, 12}, mm) && dd > 31 {
		return false, nil
	}
	if mm == 2 {
		if yyyy%400 == 0 || yyyy%4 == 0 && yyyy%100 != 0 {
			if dd > 28 {
				return false, nil
			}
		} else {
			if dd > 29 {
				return false, nil
			}
		}
	}
	return true, nil
}

func ValidNotifTime(notifTime string) (bool, error) {
	if len(notifTime) != 5 {
		return false, nil
	}
	valid, err := regexp.MatchString(`\d\d:\d\d`, notifTime)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, nil
	}
	hh, err := strconv.Atoi(notifTime[:2])
	if err != nil {
		return false, err
	}
	mm, err := strconv.Atoi(notifTime[3:])
	if err != nil {
		return false, err
	}
	return hh < 24 && mm < 60, nil
}

func ValidTimezone(timezone string) (bool, error) {
	if len(timezone) != 9 {
		return false, nil
	}
	valid, err := regexp.MatchString(`UTC[+-]\d\d:\d\d`, timezone)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, nil
	}
	hh, err := strconv.Atoi(timezone[4:6])
	if err != nil {
		return false, err
	}
	mm, err := strconv.Atoi(timezone[7:])
	if err != nil {
		return false, err
	}
	return hh < 24 && mm < 60, nil
}

func ValidRepeat(repeat string) (bool, error) {
	valid, err := regexp.MatchString(`\d+[hdw]`, repeat)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, nil
	}
	return true, nil
}
