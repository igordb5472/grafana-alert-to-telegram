package domain

type BotInput struct {
	Id                     int    `json:"id"`
	Token                  string `json:"token"`
	AlertNotifTemplateFile string `json:"alertNotifTemplate"`
	DutyNotifTemplateFile  string `json:"dutyNotifTemplate"`
	DutySchedFile          string `json:"dutySched"`
	Chats                  []ChatInput
}

type ChatInput struct {
	Id                     int    `json:"id"`
	AlertNotifTemplateFile string `json:"alertNotifTemplate"`
	DutyNotifTemplateFile  string `json:"dutyNotifTemplate"`
	DutySchedFile          string `json:"dutySched"`
}

type DutyNotifInput struct {
	Date      string   `json:"date"`
	Time      string   `json:"time"`
	Timezone  string   `json:"timezone"`
	Repeat    string   `json:"repeat"`
	Usernames []string `json:"users"`
}
