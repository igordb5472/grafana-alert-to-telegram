package service

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"unicode/utf8"
)

func SendAllAlertNotifs(ctx context.Context, data map[string]any) error {
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SendAllAlertNotifs: request timeout exceeded")
		} else {
			return fmt.Errorf("SendAllAlertNotifs: request cancelled")
		}
	}

	baseDir, ok := os.LookupEnv("BASE_DIR")
	if !ok {
		return fmt.Errorf("SendAllAlertNotifs: failed to get BASE_DIR from .env")
	}
	settingsDir, err := filepath.Abs(baseDir + "/settings")
	if err != nil {
		return fmt.Errorf("SendAllAlertNotifs: %v", err)
	}

	errCh := make(chan error)
	var wg sync.WaitGroup
	for _, bot := range Bots {
		if ctx.Err() != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("SendAllAlertNotifs: request timeout exceeded")
			} else {
				return fmt.Errorf("SendAllAlertNotifs: request cancelled")
			}
		}
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

			wg.Go(func() {
				err := SendNotif(ctx, bot.Token, chat.Id, data, templatePath)
				if ctx.Err() == nil {
					errCh <- err
				}
			})
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

func SendNotif(ctx context.Context, botToken string, chatId int, data map[string]any, templatePath string) error {
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SendNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("SendNotif: request cancelled")
		}
	}

	msgTemplate, err := template.ParseFiles(templatePath)
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SendNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("SendNotif: request cancelled")
		}
	}
	if err != nil {
		return fmt.Errorf("SendNotif: %v", err)
	}

	var builder strings.Builder
	err = msgTemplate.Execute(&builder, data)
	message := builder.String()
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SendNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("SendNotif: request cancelled")
		}
	}
	if err != nil {
		return fmt.Errorf("SendNotif: %v", err)
	}
	if len(message) == 0 {
		return nil
	}
	if utf8.RuneCountInString(message) > 4096 {
		return fmt.Errorf("SendNotif: too long message received (>4096 characters), change your template or Grafana configuration")
	}

	jsonBytes, err := json.Marshal(map[string]any{"chat_id": chatId, "text": message})
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SendNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("SendNotif: request cancelled")
		}
	}
	if err == nil && len(jsonBytes) == 0 {
		err = fmt.Errorf("empty JSON payload after marshalling")
	}
	if err != nil {
		return fmt.Errorf("SendNotif: %v", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", botToken)
	response, err := http.Post(url, "application/json", bytes.NewReader(jsonBytes))
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SendNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("SendNotif: request cancelled")
		}
	}
	if err != nil {
		return fmt.Errorf("SendNotif: %v", err)
	}

	body, err := io.ReadAll(response.Body)
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SendNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("SendNotif: request cancelled")
		}
	}
	if err != nil {
		return fmt.Errorf("SendNotif: %v", err)
	}

	bodyData := make(map[string]any)
	err = json.Unmarshal(body, &bodyData)
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SendNotif: request timeout exceeded")
		} else {
			return fmt.Errorf("SendNotif: request cancelled")
		}
	}
	if err != nil {
		return fmt.Errorf("SendNotif: %v", err)
	}

	responseOkAny, ok := bodyData["ok"]
	if !ok {
		return fmt.Errorf("SendNotif: ok field expected in response")
	}

	responseOk, ok := responseOkAny.(bool)
	if !ok {
		return fmt.Errorf("SendNotif: can't convert ok field to string")
	}

	if responseOk {
		return nil
	} else {
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
}
