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

	var scheduleData []domain.DutyNotifInput
	if err = json.Unmarshal(rawJSONData, &scheduleData); err != nil {
		return nil, err
	}

	var schedule []*domain.DutyNotif
	for _, notifInput := range scheduleData {
		date := notifInput.Date
		if len(date) == 0 {
			return nil, fmt.Errorf("getDutySchedule: date must be non-empty in duty notif")
		}
		valid, err := util.ValidDate(date)
		if err != nil {
			return nil, fmt.Errorf("getDutySchedule: validDate: %v", err)
		}
		if !valid {
			return nil, fmt.Errorf("getDutySchedule: invalid date in duty notif")
		}

		notifTime := notifInput.Time
		if len(notifTime) == 0 {
			notifTime = "00:00"
		} else {
			valid, err := util.ValidNotifTime(notifTime)
			if err != nil {
				return nil, fmt.Errorf("getDutySchedule: validTime: %v", err)
			}
			if !valid {
				return nil, fmt.Errorf("getDutySchedule: invalid time in duty notif")
			}
		}

		timezone := notifInput.Timezone
		if len(timezone) == 0 {
			timezone = "UTC+00:00"
		} else {
			valid, err := util.ValidTimezone(timezone)
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

		repeat := notifInput.Repeat
		if len(repeat) != 0 {
			valid, err := util.ValidRepeat(repeat)
			if err != nil {
				return nil, fmt.Errorf("getDutySchedule: validRepeat: %v", err)
			}
			if !valid {
				return nil, fmt.Errorf("getDutySchedule: invalid repeat in duty notif")
			}
		}

		users := notifInput.Users
		if len(users) == 0 {
			return nil, fmt.Errorf("getDutySchedule: users must be non-empty in duty notif")
		}

		schedule = append(schedule, &domain.DutyNotif{
			Date:   timeParsed,
			Repeat: repeat,
			Users:  users,
		})
	}

	return schedule, nil
}
