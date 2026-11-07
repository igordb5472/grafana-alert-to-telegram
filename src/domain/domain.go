package domain

import "time"

type Bot struct {
	Id                     int
	Token                  string
	AlertNotifTemplatePath string
	DutyNotifTemplatePath  string
	DutySchedule           []*DutyNotif
	Chats                  []*Chat
}

type Chat struct {
	Id                     int
	AlertNotifTemplatePath string
	DutyNotifTemplatePath  string
	DutySchedule           []*DutyNotif
}

type DutyNotif struct {
	Date   time.Time
	Repeat string
	Users  []string
}
