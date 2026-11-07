package domain

type BotInput struct {
	Id                     int    `json:"id"`
	Token                  string `json:"token"`
	AlertNotifTemplatePath string `json:"alertNotifTemplatePath"`
	DutyNotifTemplatePath  string `json:"dutyNotifTemplatePath"`
	DutySchedule           []*DutyNotif
	Chats                  []*Chat
}

type ChatInput struct {
	Id                     int    `json:"id"`
	AlertNotifTemplatePath string `json:"alertNotifTemplatePath"`
	DutyNotifTemplatePath  string `json:"dutyNotifTemplatePath"`
	DutySchedule           []*DutyNotif
}

type DutyNotifInput struct {
	Date     string   `json:"date"`
	Time     string   `json:"time"`
	Timezone string   `json:"timezone"`
	Repeat   string   `json:"repeat"`
	Users    []string `json:"users"`
}
