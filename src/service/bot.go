package service

import (
	"encoding/json/v2"
	"fmt"
	"grafana-alert-to-telegram/src/domain"
	"os"
	"path/filepath"
	"strings"
)

var Bots []*domain.Bot

func GetBots() ([]*domain.Bot, error) {
	baseDir, ok := os.LookupEnv("BASE_DIR")
	if !ok {
		return nil, fmt.Errorf("GetBots: failed to load BASE_DIR from env")
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
	dutySchedDir, err := filepath.Abs(settingsDir + "/duty-scheds")
	if err != nil {
		return nil, fmt.Errorf("GetBots: %v", err)
	}

	botDirEntries, err := os.ReadDir(botConfigDir)
	if err != nil {
		return nil, fmt.Errorf("GetBots: %v", err)
	}

	bots := make([]*domain.Bot, 0, len(botDirEntries))
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

		var botInput domain.BotInput
		if err = json.Unmarshal(rawJSONData, &botInput); err != nil {
			return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
		}

		id := botInput.Id
		if !ok {
			return nil, fmt.Errorf("GetBots: bot id must be non-zero in %v", botConfigFile)
		}

		token, ok := os.LookupEnv(fmt.Sprintf("BOT%v_TOKEN", id))
		if !ok {
			return nil, fmt.Errorf("GetBots: failed to get token for bot%v from .env in %v", id, botConfigFile)
		}

		alertNotifTemplatePath := ""
		alertNotifTemplateFile := botInput.AlertNotifTemplateFile
		if len(alertNotifTemplateFile) != 0 {
			if !strings.HasSuffix(alertNotifTemplateFile, ".tmpl") {
				return nil, fmt.Errorf("GetBots: alertNotifTemplate must end with .tmpl in %v", botConfigFile)
			}
			var err error
			alertNotifTemplatePath, err = filepath.Abs(alertNotifTemplateDir + "/" + alertNotifTemplateFile)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
		}

		dutyNotifTemplatePath := ""
		alertNotifTemplateFile = botInput.DutyNotifTemplateFile
		if len(alertNotifTemplateFile) != 0 {
			if !strings.HasSuffix(alertNotifTemplateFile, ".tmpl") {
				return nil, fmt.Errorf("GetBots: dutyNotifTemplate must end with .tmpl in %v", botConfigFile)
			}
			var err error
			dutyNotifTemplatePath, err = filepath.Abs(dutyNotifTemplateDir + "/" + alertNotifTemplateFile)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
		}

		var dutySched []*domain.DutyNotif
		dutySchedFile := botInput.DutySchedFile
		if len(dutySchedFile) != 0 {
			if !strings.HasSuffix(dutySchedFile, ".json") {
				return nil, fmt.Errorf("GetBots: dutySched must end with .json in %v", botConfigFile)
			}
			dutySchedPath, err := filepath.Abs(dutySchedDir + "/" + dutySchedFile)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
			dutySched, err = getDutySched(dutySchedPath)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
		}

		var chats []*domain.Chat
		chatsInput := botInput.Chats
		if len(chatsInput) == 0 {
			return nil, fmt.Errorf("GetBots: at least one chat must be configured in %v", botConfigFile)
		} else {
			var err error
			chats, err = getChats(chatsInput)
			if err != nil {
				return nil, fmt.Errorf("GetBots: %v in %v", err, botConfigFile)
			}
		}

		bots = append(bots, &domain.Bot{
			Id:                     id,
			Token:                  token,
			AlertNotifTemplatePath: alertNotifTemplatePath,
			DutyNotifTemplatePath:  dutyNotifTemplatePath,
			DutySched:              dutySched,
			Chats:                  chats,
		})
	}

	return bots, nil
}
