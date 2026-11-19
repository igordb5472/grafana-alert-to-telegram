package service

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"grafana-alert-to-telegram/src/domain"
	"grafana-alert-to-telegram/src/util"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func ScheduleAllDutyNotifs() error {
	baseDir, ok := os.LookupEnv("BASE_DIR")
	if !ok {
		return fmt.Errorf("SchedAllDutyNotifs: failed to get BASE_DIR from .env")
	}
	settingsDir, err := filepath.Abs(baseDir + "/settings")
	if err != nil {
		return fmt.Errorf("SchedAllDutyNotifs: %v", err)
	}

	bots, err := GetBots()
	if err != nil {
		return fmt.Errorf("SchedAllDutyNotifs: %v", err)
	}

	errCh := make(chan error)
	var wg sync.WaitGroup
	for _, bot := range bots {
		for _, chat := range bot.Chats {
			for _, notif := range chat.DutySched {
				var templatePath string
				if len(chat.DutyNotifTemplatePath) > 0 {
					templatePath = chat.DutyNotifTemplatePath
				} else if len(bot.DutyNotifTemplatePath) > 0 {
					templatePath = bot.DutyNotifTemplatePath
				} else {
					templatePath = settingsDir + "/default-templates/dutyNotif.default.tmpl"
				}

				wg.Go(func() {
					err := ScheduleDutyNotif(bot.Token, chat.Id, notif, templatePath)
					errCh <- err
				})
			}
		}
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errFull error
	for err := range errCh {
		if err != nil {
			errFull = errors.Join(errFull, err)
		}
	}
	return errFull
}

func ScheduleDutyNotif(botToken string, chatId int, notif *domain.DutyNotif, templatePath string) error {
	var interval time.Duration
	if len(notif.Repeat) > 0 {
		base, err := strconv.Atoi(notif.Repeat[:len(notif.Repeat)-1])
		if err != nil {
			return fmt.Errorf("SchedDutyNotif: %v", err)
		}
		interval = time.Duration(base)
		switch notif.Repeat[len(notif.Repeat)-1] {
		case 'h':
			interval *= time.Hour
		case 'd':
			interval *= 24 * time.Hour
		case 'w':
			interval *= 7 * 24 * time.Hour
		}
	}

	nextNotifDate := notif.Date

	go SendDutyNotif(nextNotifDate, interval, botToken, chatId, map[string]any{"usernames": notif.Usernames}, templatePath)

	return nil
}

func SendDutyNotif(nextNotifDate time.Time, interval time.Duration, botToken string, chatId int, data map[string]any, templatePath string) {
	for nextNotifDate.Before(time.Now()) {
		nextNotifDate = nextNotifDate.Add(interval)
	}

	for nextNotifDate.After(time.Now()) {
		time.Sleep(time.Until(nextNotifDate))

		err := SendNotif(context.Background(), botToken, chatId, data, templatePath)
		if err != nil {
			log.Printf("failed to send duty notif: %v", err)
		}

		nextNotifDate = nextNotifDate.Add(interval)
	}
}

func getDutySched(path string) ([]*domain.DutyNotif, error) {
	rawJSONData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var schedData []domain.DutyNotifInput
	if err = json.Unmarshal(rawJSONData, &schedData); err != nil {
		return nil, err
	}

	var sched []*domain.DutyNotif
	for _, notifInput := range schedData {
		date := notifInput.Date
		if len(date) == 0 {
			return nil, fmt.Errorf("getDutySched: date must be non-empty in duty notif")
		}
		valid, err := util.ValidDate(date)
		if err != nil {
			return nil, fmt.Errorf("getDutySched: %v", err)
		}
		if !valid {
			return nil, fmt.Errorf("getDutySched: invalid date in duty notif")
		}

		notifTime := notifInput.Time
		if len(notifTime) == 0 {
			notifTime = "00:00"
		} else {
			valid, err := util.ValidNotifTime(notifTime)
			if err != nil {
				return nil, fmt.Errorf("getDutySched: %v", err)
			}
			if !valid {
				return nil, fmt.Errorf("getDutySched: invalid time in duty notif")
			}
		}

		timezone := notifInput.Timezone
		if len(timezone) == 0 {
			timezone = "UTC+00:00"
		} else {
			valid, err := util.ValidTimezone(timezone)
			if err != nil {
				return nil, fmt.Errorf("getDutySched: %v", err)
			}
			if !valid {
				return nil, fmt.Errorf("getDutySched: invalid timezone in duty notif")
			}
		}

		timeParsed, err := time.Parse("02.01.2006 15:04 UTC-07:00", date+" "+notifTime+" "+timezone)
		if err != nil {
			return nil, fmt.Errorf("getDutySched: failed to parse time from string: %v %v %v", date, notifTime, timezone)
		}

		repeat := notifInput.Repeat
		if len(repeat) != 0 {
			valid, err := util.ValidRepeat(repeat)
			if err != nil {
				return nil, fmt.Errorf("getDutySched: %v", err)
			}
			if !valid {
				return nil, fmt.Errorf("getDutySched: invalid repeat in duty notif")
			}
		}

		usernames := notifInput.Usernames
		if len(usernames) == 0 {
			return nil, fmt.Errorf("getDutySched: users must be non-empty in duty notif")
		}
		valid, err = util.ValidUsers(usernames)
		if err != nil {
			return nil, fmt.Errorf("getDutySchedule: %v", err)
		}
		if !valid {
			return nil, fmt.Errorf("getDutySched: invalid usernames in duty notif")
		}

		sched = append(sched, &domain.DutyNotif{
			Date:      timeParsed,
			Repeat:    repeat,
			Usernames: usernames,
		})
	}

	return sched, nil
}
