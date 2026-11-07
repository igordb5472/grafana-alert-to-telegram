package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"grafana-alert-to-telegram/src/config"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

func SendAllAlertNotifs(ctx context.Context, data map[string]any) error {
	baseDir, ok := os.LookupEnv("BASE_DIR")
	if !ok {
		return fmt.Errorf("SendAllAlertNotifs: failed to get BASE_DIR from .env")
	}
	settingsDir, err := filepath.Abs(baseDir + "/settings")
	if err != nil {
		return fmt.Errorf("SendAllAlertNotifs: %v", err)
	}

	var errChLen int
	localErrChans := make(map[chan error]struct{})
	for _, bot := range config.Bots {
		if ctx.Err() != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("SendAllAlertNotifs: request timeout exceeded")
			} else {
				return fmt.Errorf("SendAllAlertNotifs: request cancelled")
			}
		}

		localErrCh := make(chan error, len(bot.Chats))
		localErrChans[localErrCh] = struct{}{}
		errChLen += len(localErrCh)
		for _, chat := range bot.Chats {
			if ctx.Err() != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return fmt.Errorf("SendAllAlertNotifs: request timeout exceeded")
				} else {
					return fmt.Errorf("SendAllAlertNotifs: request cancelled")
				}
			}

			var templatePath string
			if len(chat.AlertNotifTemplatePath) > 0 {
				templatePath = chat.AlertNotifTemplatePath
			} else if len(bot.AlertNotifTemplatePath) > 0 {
				templatePath = bot.AlertNotifTemplatePath
			} else {
				templatePath = settingsDir + "/default-templates/alertNotif.default.tmpl"
			}

			var wg sync.WaitGroup
			wg.Go(func() {
				err := SendNotif(ctx, bot.Token, chat.Id, data, templatePath)
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

func SendNotif(ctx context.Context, botToken string, chatId int, data map[string]any, templatePath string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage?chat_id=%v", botToken, chatId)

	errCh := make(chan error)
	var err error

	var msgTemplate *template.Template
	go func() {
		msgTemplate, err = template.ParseFiles(templatePath)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	var buffer bytes.Buffer
	go func() {
		err = msgTemplate.Execute(&buffer, data)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	var response *http.Response
	go func() {
		response, err = http.Post(url, "application/json", &buffer)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	var body []byte
	go func() {
		_, err = response.Body.Read(body)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	bodyData := make(map[string]any)
	go func() {
		err = json.Unmarshal(body, &bodyData)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	responseOkAny, ok := bodyData["ok"]
	if !ok {
		return fmt.Errorf("SendNotif: ok field expected in response")
	}

	responseOk, ok := responseOkAny.(string)
	if !ok {
		return fmt.Errorf("SendNotif: can't convert ok field to string")
	}

	switch responseOk {
	case "True":
		return nil
	case "False":
		responseDescAny, ok := bodyData["description"]
		if !ok {
			return fmt.Errorf("SendNotif: response failed due to unknown error")
		} else {
			responseDesc, ok := responseDescAny.(string)
			if !ok {
				return fmt.Errorf("SendNotif: response failed due to error that can't be converted to string")
			} else {
				return fmt.Errorf("SendNotif: response failed: %v", responseDesc)
			}
		}
	default:
		return fmt.Errorf("response ok field neither True nor False")
	}
}
