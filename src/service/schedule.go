package service

import (
	"context"
	"errors"
	"fmt"
	"grafana-alert-to-telegram/src/config"
	"grafana-alert-to-telegram/src/domain"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func ScheduleAllDutyNotifs(ctx context.Context) error {
	baseDir, ok := os.LookupEnv("BASE_DIR")
	if !ok {
		return fmt.Errorf("ScheduleAllDutyNotifs: failed to get BASE_DIR from .env")
	}
	settingsDir, err := filepath.Abs(baseDir + "/settings")
	if err != nil {
		return fmt.Errorf("ScheduleAllDutyNotifs: %v", err)
	}

	bots, err := config.GetBots()
	if err != nil {
		return fmt.Errorf("ScheduleAllDutyNotifs: %v", err)
	}

	var errChLen int
	localErrChans := make(map[chan error]struct{})
	for _, bot := range bots {
		if ctx.Err() != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("ScheduleAllDutyNotifs: request timeout exceeded")
			} else {
				return fmt.Errorf("ScheduleAllDutyNotifs: request cancelled")
			}
		}

		for _, chat := range bot.Chats {
			if ctx.Err() != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return fmt.Errorf("ScheduleAllDutyNotifs: request timeout exceeded")
				} else {
					return fmt.Errorf("ScheduleAllDutyNotifs: request cancelled")
				}
			}

			localErrCh := make(chan error, len(chat.DutySchedule))
			localErrChans[localErrCh] = struct{}{}
			errChLen += len(localErrCh)
			for _, notif := range chat.DutySchedule {
				if ctx.Err() != nil {
					if ctx.Err() == context.DeadlineExceeded {
						return fmt.Errorf("ScheduleAllDutyNotifs: request timeout exceeded")
					} else {
						return fmt.Errorf("ScheduleAllDutyNotifs: request cancelled")
					}
				}

				var templatePath string
				if len(chat.DutyNotifTemplatePath) > 0 {
					templatePath = chat.DutyNotifTemplatePath
				} else if len(bot.DutyNotifTemplatePath) > 0 {
					templatePath = bot.DutyNotifTemplatePath
				} else {
					templatePath = settingsDir + "/default-templates/dutyNotif.default.tmpl"
				}

				var wg sync.WaitGroup
				wg.Go(func() {
					err := ScheduleDutyNotif(ctx, bot.Token, chat.Id, notif, templatePath)
					if ctx.Err() == nil {
						localErrCh <- err
					}
				})
				go func() {
					wg.Wait()
					close(localErrCh)
				}()
			}
		}
	}
	errCh := make(chan error, errChLen)

	for len(localErrChans) > 0 {
		for localErrCh := range localErrChans {
			select {
			case err, ok := <-localErrCh:
				if !ok {
					delete(localErrChans, localErrCh)
				} else {
					if ctx.Err() != nil {
						if ctx.Err() == context.DeadlineExceeded {
							return fmt.Errorf("ScheduleAllDutyNotifs: request timeout exceeded")
						} else {
							return fmt.Errorf("ScheduleAllDutyNotifs: request cancelled")
						}
					}
					errCh <- err
				}
			default:
			}
		}
	}

	var errFull error
	for err := range errCh {
		if err != nil {
			errors.Join(errFull, err)
		}
	}

	return errFull
}

func ScheduleDutyNotif(ctx context.Context, botToken string, chatId int, notif *domain.DutyNotif, templatePath string) error {
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("ScheduleDutyNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("ScheduleDutyNotif: request cancelled")
		}
	}

	var interval time.Duration
	if len(notif.Repeat) > 0 {
		base, err := strconv.Atoi(notif.Repeat[:len(notif.Repeat)-1])
		if err != nil {
			return fmt.Errorf("ScheduleDutyNotif: %v", err)
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
	for ctx.Err() == nil && nextNotifDate.Before(time.Now()) {
		nextNotifDate = nextNotifDate.Add(interval)
	}
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("ScheduleDutyNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("ScheduleDutyNotif: request cancelled")
		}
	}

	go sendDutyNotif(nextNotifDate, interval, botToken, chatId, map[string]any{"Users": notif.Users}, templatePath)

	return nil
}

func sendDutyNotif(nextNotifDate time.Time, interval time.Duration, botToken string, chatId int, data map[string]any, templatePath string) {
	for nextNotifDate.Before(time.Now()) {
		nextNotifDate = nextNotifDate.Add(interval)
	}
	for nextNotifDate.After(time.Now()) {
		time.Sleep(time.Until(nextNotifDate))
		errCh := make(chan error)
		go func() {
			err := SendNotif(context.Background(), botToken, chatId, data, templatePath)
			errCh <- err
		}()
		go func() {
			err := <-errCh
			if err != nil {
				log.Printf("failed to send notif: %v", err)
			}
		}()
		nextNotifDate.Add(interval)
	}
}
