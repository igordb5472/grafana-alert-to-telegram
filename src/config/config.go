package config

import (
	"grafana-alert-to-telegram/src/domain"
)

var Bots []*domain.Bot

func LoadBots() error {
	var err error
	Bots, err = GetBots()
	return err
}
