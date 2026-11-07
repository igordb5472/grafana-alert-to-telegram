package config

import (
	"fmt"
	"grafana-alert-to-telegram/src/domain"
	"os"
	"path/filepath"
	"strings"
)

func getChats(chatsAnyArray []any) ([]*domain.Chat, error) {
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

	var chats []*domain.Chat
	for _, chatAny := range chatsAnyArray {
		chatData, ok := chatAny.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("getChats: chat must be object with string keys")
		}

		var newChat domain.Chat

		idAny, ok := chatData["id"]
		if !ok {
			return nil, fmt.Errorf("getChats: chat id must be configured")
		}
		idFloat, ok := idAny.(float64)
		if !ok {
			return nil, fmt.Errorf("getChats: chat id must be number")
		}
		id := int(idFloat)
		if float64(id) != idFloat {
			return nil, fmt.Errorf("getChats: chat id must be int")
		}
		newChat.Id = id

		var alertNotifTemplatePath string
		alertNotifTemplateFileAny, ok := chatData["alertNotifTemplate"]
		if ok {
			alertNotifTemplateFile, ok := alertNotifTemplateFileAny.(string)
			if !ok {
				return nil, fmt.Errorf("getChats: alertNotifTemplate must be string")
			}
			if !strings.HasSuffix(alertNotifTemplateFile, ".tmpl") {
				return nil, fmt.Errorf("getChats: alertNotifTemplate must end with .tmpl")
			}
			alertNotifTemplatePath, err = filepath.Abs(alertNotifTemplateDir + "/" + alertNotifTemplateFile)
			if err != nil {
				return nil, fmt.Errorf("getChats: %v", err)
			}
		}
		newChat.AlertNotifTemplatePath = alertNotifTemplatePath

		var dutyNotifTemplatePath string
		dutyNotifTemplateFileAny, ok := chatData["dutyNotifTemplate"]
		if ok {
			dutyNotifTemplateFile, ok := dutyNotifTemplateFileAny.(string)
			if !ok {
				return nil, fmt.Errorf("getChats: dutyNotifTemplate must be string")
			}
			if !strings.HasSuffix(dutyNotifTemplateFile, ".tmpl") {
				return nil, fmt.Errorf("getChats: dutyNotifTemplate must end with .tmpl")
			}
			dutyNotifTemplatePath, err = filepath.Abs(dutyNotifTemplateDir + "/" + dutyNotifTemplateFile)
			if err != nil {
				return nil, fmt.Errorf("getChats: %v", err)
			}
		}
		newChat.DutyNotifTemplatePath = dutyNotifTemplatePath

		var dutySchedule []*domain.DutyNotif
		dutyScheduleFileAny, ok := chatData["dutySchedule"]
		if ok {
			dutyScheduleFile, ok := dutyScheduleFileAny.(string)
			if !ok {
				return nil, fmt.Errorf("getChats: dutySchedule must be string")
			}
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
		newChat.DutySchedule = dutySchedule

		chats = append(chats, &newChat)
	}

	return chats, nil
}
