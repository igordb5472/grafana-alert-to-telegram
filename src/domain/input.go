package domain

type BotInput struct {
	Id                     int    `json:"id"`
	Token                  string `json:"token"`
	AlertNotifTemplateFile string `json:"alertNotifTemplate"`
	DutyNotifTemplateFile  string `json:"dutyNotifTemplate"`
	DutyScheduleFile       string `json:"dutySchedule"`
	Chats                  []ChatInput
}

type ChatInput struct {
	Id                     int    `json:"id"`
	AlertNotifTemplateFile string `json:"alertNotifTemplate"`
	DutyNotifTemplateFile  string `json:"dutyNotifTemplate"`
	DutyScheduleFile       string `json:"dutySchedule"`
}

type DutyNotifInput struct {
	Date     string   `json:"date"`
	Time     string   `json:"time"`
	Timezone string   `json:"timezone"`
	Repeat   string   `json:"repeat"`
	Users    []string `json:"users"`
}
