package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"grafana-alert-to-telegram/src/config"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	var wg sync.WaitGroup
	for localErrCh := range localErrChans {
		wg.Go(func() {
			select {
			case err, ok := <-localErrCh:
				if !ok {
					delete(localErrChans, localErrCh)
				} else {
					if ctx.Err() == nil {
						errCh <- err
					}
				}
			case <-ctx.Done():
				return
			}
		})
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

func SendNotif(ctx context.Context, botToken string, chatId int, data map[string]any, templatePath string) error {
	errCh := make(chan error)

	var msgTemplate *template.Template
	go func() {
		var err error
		msgTemplate, err = template.ParseFiles(templatePath)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	var message string
	go func() {
		var builder strings.Builder
		err := msgTemplate.Execute(&builder, data)
		message = builder.String()
		if err == nil && len(message) == 0 {
			err = fmt.Errorf("empty message after template execution")
		}
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	var jsonPayload []byte
	go func() {
		var err error
		jsonPayload, err = json.Marshal(map[string]any{"chat_id": chatId, "text": message})
		if err == nil && len(jsonPayload) == 0 {
			err = fmt.Errorf("empty JSON payload after marshalling")
		}
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", botToken)
	var response *http.Response
	go func() {
		// request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
		// request.ContentLength = int64(len(jsonPayload))
		// response, err = http.DefaultClient.Do(request)
		var err error
		response, err = http.Post(url, "application/json", bytes.NewReader(jsonPayload))
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}
	defer response.Body.Close()

	var body []byte
	go func() {
		var err error
		body, err = io.ReadAll(response.Body)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	go func() {
		err := response.Body.Close()
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	bodyData := make(map[string]any)
	go func() {
		err := json.Unmarshal(body, &bodyData)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("SendNotif: %v", err)
		}
	}

	responseOkAny, ok := bodyData["ok"]
	if !ok {
		return fmt.Errorf("SendNotif: ok field expected in response")
	}

	responseOk, ok := responseOkAny.(bool)
	if !ok {
		return fmt.Errorf("SendNotif: can't convert ok field to string")
	}

	if !responseOk {
		responseDescAny, ok := bodyData["description"]
		if !ok {
			return fmt.Errorf("SendNotif: Telegram API response failed due to unknown error")
		} else {
			responseDesc, ok := responseDescAny.(string)
			if !ok {
				return fmt.Errorf("SendNotif: Telgram API response failed due to error that can't be converted to string")
			} else {
				return fmt.Errorf("SendNotif: Telegram API response failed with description: %v", responseDesc)
			}
		}
	}

	return nil
}
