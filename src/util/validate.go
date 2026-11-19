package util

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
)

func ValidDate(date string) (bool, error) {
	valid, err := regexp.MatchString(`^\d\d.\d\d.\d\d\d\d$`, date)
	if err != nil {
		return false, fmt.Errorf("ValidDate: %v", err)
	}
	if !valid {
		return false, nil
	}
	dd, err := strconv.Atoi(date[:2])
	if err != nil {
		return false, fmt.Errorf("ValidDate: %v", err)
	}
	mm, err := strconv.Atoi(date[3:5])
	if err != nil {
		return false, fmt.Errorf("ValidDate: %v", err)
	}
	yyyy, err := strconv.Atoi(date[6:])
	if err != nil {
		return false, fmt.Errorf("ValidDate: %v", err)
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
	valid, err := regexp.MatchString(`^\d\d:\d\d$`, notifTime)
	if err != nil {
		return false, fmt.Errorf("ValidNotifTime: %v", err)
	}
	if !valid {
		return false, nil
	}
	hh, err := strconv.Atoi(notifTime[:2])
	if err != nil {
		return false, fmt.Errorf("ValidNotifTime: %v", err)
	}
	mm, err := strconv.Atoi(notifTime[3:])
	if err != nil {
		return false, fmt.Errorf("ValidNotifTime: %v", err)
	}
	return hh < 24 && mm < 60, nil
}

func ValidTimezone(timezone string) (bool, error) {
	valid, err := regexp.MatchString(`^UTC[+-]\d\d:\d\d$`, timezone)
	if err != nil {
		return false, fmt.Errorf("ValidTimezone: %v", err)
	}
	if !valid {
		return false, nil
	}
	hh, err := strconv.Atoi(timezone[4:6])
	if err != nil {
		return false, fmt.Errorf("ValidTimezone: %v", err)
	}
	mm, err := strconv.Atoi(timezone[7:])
	if err != nil {
		return false, fmt.Errorf("ValidTimezone: %v", err)
	}
	return hh < 24 && mm < 60, nil
}

func ValidRepeat(repeat string) (bool, error) {
	valid, err := regexp.MatchString(`^[1-9]\d*[mhdw]$`, repeat)
	if err != nil {
		return false, fmt.Errorf("ValidRepeat: %v", err)
	}
	if !valid {
		return false, nil
	}
	return true, nil
}

func ValidUsers(usernames []string) (bool, error) {
	for _, username := range usernames {
		valid, err := regexp.MatchString(`^@[A-Za-z0-9_]{5,32}$`, username)
		if err != nil {
			return false, fmt.Errorf("ValidUsers: %v", err)
		}
		if !valid {
			return false, nil
		}
	}
	return true, nil
}
