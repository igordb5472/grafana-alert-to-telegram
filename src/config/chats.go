package config

import (
	"fmt"
	"grafana-alert-to-telegram/src/domain"
	"os"
	"path/filepath"
	"strings"
)

func getChats(chatsInput []domain.ChatInput) ([]*domain.Chat, error) {
	baseDir, ok := os.LookupEnv("BASE_DIR")
	if !ok {
		return nil, fmt.Errorf("GetBots: failed to get BASE_DIR from .env")
	}
	settingsDir, err := filepath.Abs(baseDir + "/settings")
	if err != nil {
		return nil, fmt.Errorf("getChats: %v", err)
	}
	alertNotifTemplateDir, err := filepath.Abs(settingsDir + "/alert-notif-templates")
	if err != nil {
		return nil, fmt.Errorf("getChats: %v", err)
	}
	dutyNotifTemplateDir, err := filepath.Abs(settingsDir + "/duty-notif-templates")
	if err != nil {
		return nil, fmt.Errorf("getChats: %v", err)
	}
	dutyScheduleDir, err := filepath.Abs(settingsDir + "/duty-schedules")
	if err != nil {
		return nil, fmt.Errorf("getChats: %v", err)
	}

	chats := make([]*domain.Chat, 0, len(chatsInput))
	for _, chatInput := range chatsInput {
		id := chatInput.Id
		if !ok {
			return nil, fmt.Errorf("getChats: chat id must be non-nil")
		}

		alertNotifTemplatePath := ""
		alertNotifTemplateFile := chatInput.AlertNotifTemplateFile
		if len(alertNotifTemplateFile) != 0 {
			if !strings.HasSuffix(alertNotifTemplateFile, ".tmpl") {
				return nil, fmt.Errorf("getChats: alertNotifTemplate must end with .tmpl")
			}
			var err error
			alertNotifTemplatePath, err = filepath.Abs(alertNotifTemplateDir + "/" + alertNotifTemplateFile)
			if err != nil {
				return nil, fmt.Errorf("getChats: %v", err)
			}
		}

		dutyNotifTemplatePath := ""
		dutyNotifTemplateFile := chatInput.DutyNotifTemplateFile
		if len(dutyNotifTemplateFile) != 0 {
			if !strings.HasSuffix(dutyNotifTemplateFile, ".tmpl") {
				return nil, fmt.Errorf("getChats: dutyNotifTemplate must end with .tmpl")
			}
			var err error
			dutyNotifTemplatePath, err = filepath.Abs(dutyNotifTemplateDir + "/" + dutyNotifTemplateFile)
			if err != nil {
				return nil, fmt.Errorf("getChats: %v", err)
			}
		}

		var dutySchedule []*domain.DutyNotif
		dutyScheduleFile := chatInput.DutyScheduleFile
		if len(dutyScheduleFile) != 0 {
			if !strings.HasSuffix(dutyScheduleFile, ".json") {
				return nil, fmt.Errorf("getChats: dutySchedule must end with .json")
			}
			dutySchedulePath, err := filepath.Abs(dutyScheduleDir + "/" + dutyScheduleFile)
			if err != nil {
				return nil, fmt.Errorf("getChats: %v", err)
			}
			dutySchedule, err = getDutySchedule(dutySchedulePath)
			if err != nil {
				return nil, fmt.Errorf("getChats: %v", err)
			}
		}

		chats = append(chats, &domain.Chat{
			Id:                     id,
			AlertNotifTemplatePath: alertNotifTemplatePath,
			DutyNotifTemplatePath:  dutyNotifTemplatePath,
			DutySchedule:           dutySchedule,
		})
	}

	return chats, nil
}
