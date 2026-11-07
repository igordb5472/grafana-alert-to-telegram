package config

import (
	"encoding/json"
	"fmt"
	"grafana-alert-to-telegram/src/domain"
	"grafana-alert-to-telegram/src/util"
	"os"
	"time"
)

func getDutySchedule(path string) ([]*domain.DutyNotif, error) {
	rawJSONData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var scheduleData []any
	if err = json.Unmarshal(rawJSONData, &scheduleData); err != nil {
		return nil, err
	}

	var schedule []*domain.DutyNotif
	for _, notifAny := range scheduleData {
		notifData, ok := notifAny.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("getDutySchedule: duty notif must be object with string keys")
		}

		var (
			newNotif domain.DutyNotif
			valid    bool
		)

		dateAny, ok := notifData["date"]
		if !ok {
			return nil, fmt.Errorf("getDutySchedule: date must be configured in duty notif")
		}
		date, ok := dateAny.(string)
		if !ok {
			return nil, fmt.Errorf("getDutySchedule: date must be string in duty notif")
		}
		valid, err = util.ValidDate(date)
		if err != nil {
			return nil, fmt.Errorf("getDutySchedule: validDate: %v", err)
		}
		if !valid {
			return nil, fmt.Errorf("getDutySchedule: invalid date in duty notif")
		}

		timeAny, ok := notifData["time"]
		var notifTime string
		if !ok {
			notifTime = "00:00"
		} else {
			notifTime, ok = timeAny.(string)
			if !ok {
				return nil, fmt.Errorf("getDutySchedule: time must be string in duty notif")
			}
			valid, err = util.ValidNotifTime(notifTime)
			if err != nil {
				return nil, fmt.Errorf("getDutySchedule: validTime: %v", err)
			}
			if !valid {
				return nil, fmt.Errorf("getDutySchedule: invalid time in duty notif")
			}
		}

		timezoneAny, ok := notifData["timezone"]
		var timezone string
		if !ok {
			timezone = "UTC+00:00"
		} else {
			timezone, ok = timezoneAny.(string)
			if !ok {
				return nil, fmt.Errorf("getDutySchedule: timezone must be string in duty notif")
			}
			valid, err = util.ValidTimezone(timezone)
			if err != nil {
				return nil, fmt.Errorf("getDutySchedule: validTimezone: %v", err)
			}
			if !valid {
				return nil, fmt.Errorf("getDutySchedule: invalid timezone in duty notif")
			}
		}

		timeParsed, err := time.Parse("02.01.2006 15:04 UTC-07:00", date+" "+notifTime+" "+timezone)
		if err != nil {
			return nil, fmt.Errorf("getDutySchedule: failed to parse time from string: %v %v %v", date, notifTime, timezone)
		}
		newNotif.Date = timeParsed

		repeatAny, ok := notifData["repeat"]
		var repeat string
		if ok {
			repeat, ok = repeatAny.(string)
			if !ok {
				return nil, fmt.Errorf("getDutySchedule: repeat must be string in duty notif")
			}
			valid, err = util.ValidRepeat(repeat)
			if err != nil {
				return nil, fmt.Errorf("getDutySchedule: validRepeat: %v", err)
			}
			if !valid {
				return nil, fmt.Errorf("getDutySchedule: invalid repeat in duty notif")
			}
		}
		newNotif.Repeat = repeat

		usersAny, ok := notifData["users"]
		if !ok {
			return nil, fmt.Errorf("getDutySchedule: users must be configured in duty notif")
		}
		usersAnyArray, ok := usersAny.([]any)
		if !ok {
			return nil, fmt.Errorf("getDutySchedule: users must be array")
		}
		var users []string
		for _, userAny := range usersAnyArray {
			user, ok := userAny.(string)
			if !ok {
				return nil, fmt.Errorf("getDutySchedule: user must be string in duty notif users")
			}
			users = append(users, user)
		}
		newNotif.Users = users

		schedule = append(schedule, &newNotif)
	}

	return schedule, nil
}
