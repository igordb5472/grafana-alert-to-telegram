package domain

import "time"

type Bot struct {
	Id                     int
	Token                  string
	AlertNotifTemplatePath string
	DutyNotifTemplatePath  string
	DutySched              []*DutyNotif
	Chats                  []*Chat
}

type Chat struct {
	Id                     int
	AlertNotifTemplatePath string
	DutyNotifTemplatePath  string
	DutySched              []*DutyNotif
}

type DutyNotif struct {
	Date      time.Time
	Repeat    string
	Usernames []string
}
