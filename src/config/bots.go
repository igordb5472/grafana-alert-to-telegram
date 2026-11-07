package config

import (
	"encoding/json"
	"fmt"
	"grafana-alert-to-telegram/src/domain"
	"os"
	"path/filepath"
	"strings"
)

func GetBots() ([]*domain.Bot, error) {
	baseDir, ok := os.LookupEnv("BASE_DIR")
	if !ok {
		return nil, fmt.Errorf("GetBots: failed to get BASE_DIR from .env")
	}
	settingsDir, err := filepath.Abs(baseDir + "/settings")
	if err != nil {
		return nil, fmt.Errorf("GetBots: %v", err)
	}
	botConfigDir, err := filepath.Abs(settingsDir + "/bots")
	if err != nil {
		return nil, fmt.Errorf("GetBots: %v", err)
	}
	alertNotifTemplateDir, err := filepath.Abs(settingsDir + "/alert-notif-templates")
	if err != nil {
		return nil, fmt.Errorf("GetBots: %v", err)
	}
	dutyNotifTemplateDir, err := filepath.Abs(settingsDir + "/duty-notif-templates")
	if err != nil {
		return nil, fmt.Errorf("GetBots: %v", err)
	}
	dutyScheduleDir, err := filepath.Abs(settingsDir + "/duty-schedules")
	if err != nil {
		return nil, fmt.Errorf("GetBots: %v", err)
	}

	botDirEntries, err := os.ReadDir(botConfigDir)
	if err != nil {
		return nil, fmt.Errorf("GetBots: %v", err)
	}

	bots := make([]*domain.Bot, len(botDirEntries))
	for _, dirEntry := range botDirEntries {
		botConfigFile := dirEntry.Name()
		botConfigPath, err := filepath.Abs(botConfigDir + "/" + botConfigFile)
		if err != nil {
			return nil, fmt.Errorf("GetBots: %v", err)
		}
		rawJSONData, err := os.ReadFile(botConfigPath)
		if err != nil {
			return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
		}

		botData := make(map[string]any)
		if err = json.Unmarshal(rawJSONData, &botData); err != nil {
			return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
		}

		var newBot domain.Bot

		idAny, ok := botData["id"]
		if !ok {
			return nil, fmt.Errorf("GetBots: bot id must be configured in %v", botConfigFile)
		}
		idFloat, ok := idAny.(float64)
		if !ok {
			return nil, fmt.Errorf("GetBots: bot id must be number in %v", botConfigFile)
		}
		id := int(idFloat)
		if float64(id) != idFloat {
			return nil, fmt.Errorf("GetBots: bot id must be int in %v", botConfigFile)
		}
		newBot.Id = id

		token, ok := os.LookupEnv(fmt.Sprintf("BOT%v_TOKEN", id))
		if !ok {
			return nil, fmt.Errorf("GetBots: failed to get token for bot%v from .env in %v", id, botConfigFile)
		}
		newBot.Token = token

		var alertNotifTemplatePath string
		alertNotifTemplateFileAny, ok := botData["alertNotifTemplate"]
		if ok {
			alertNotifTemplateFile, ok := alertNotifTemplateFileAny.(string)
			if !ok {
				return nil, fmt.Errorf("GetBots: alertNotifTemplate must be string in %v", botConfigFile)
			}
			if !strings.HasSuffix(alertNotifTemplateFile, ".tmpl") {
				return nil, fmt.Errorf("GetBots: alertNotifTemplate must end with .tmpl in %v", botConfigFile)
			}
			alertNotifTemplatePath, err = filepath.Abs(alertNotifTemplateDir + "/" + alertNotifTemplateFile)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
		}
		newBot.AlertNotifTemplatePath = alertNotifTemplatePath

		var dutyNotifTemplatePath string
		alertNotifTemplateFileAny, ok = botData["dutyNotifTemplate"]
		if ok {
			alertNotifTemplateFile, ok := alertNotifTemplateFileAny.(string)
			if !ok {
				return nil, fmt.Errorf("GetBots: dutyNotifTemplate must be string in %v", botConfigFile)
			}
			if !strings.HasSuffix(alertNotifTemplateFile, ".tmpl") {
				return nil, fmt.Errorf("GetBots: dutyNotifTemplate must end with .tmpl in %v", botConfigFile)
			}
			dutyNotifTemplatePath, err = filepath.Abs(dutyNotifTemplateDir + "/" + alertNotifTemplateFile)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
		}
		newBot.DutyNotifTemplatePath = dutyNotifTemplatePath

		var dutySchedule []*domain.DutyNotif
		dutyScheduleFileAny, ok := botData["dutySchedule"]
		if ok {
			dutyScheduleFile, ok := dutyScheduleFileAny.(string)
			if !ok {
				return nil, fmt.Errorf("GetBots: dutySchedule must be string in %v", botConfigFile)
			}
			if !strings.HasSuffix(dutyScheduleFile, ".json") {
				return nil, fmt.Errorf("GetBots: dutySchedule must end with .json in %v", botConfigFile)
			}
			dutySchedulePath, err := filepath.Abs(dutyScheduleDir + "/" + dutyScheduleFile)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
			dutySchedule, err = getDutySchedule(dutySchedulePath)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
		}
		newBot.DutySchedule = dutySchedule

		var chats []*domain.Chat
		chatsAny, ok := botData["chats"]
		if !ok {
			return nil, fmt.Errorf("GetBots: at least one chat must be configured in %v", botConfigFile)
		} else {
			chatsAnyArray, ok := chatsAny.([]any)
			if !ok {
				return nil, fmt.Errorf("GetBots: chats must be array in %v", botConfigFile)
			}
			chats, err = getChats(chatsAnyArray)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
		}
		newBot.Chats = chats

		bots = append(bots, &newBot)
	}

	return bots, nil
}
