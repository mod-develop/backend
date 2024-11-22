package types

import "github.com/mod-develop/backend/internal/models"

type TAction struct {
	IsQuestCreater bool
}

type TMaster struct {
	ID   uint
	Code string
}

type TPlayer struct {
	MasterIDs []uint
}

type TUser struct {
	ID            uint
	Login         string
	IsAdmin       bool
	IsQuestMaster bool
	Action        TAction
	Master        TMaster
	Player        TPlayer
}

type TMainPage struct {
	User TUser
}

type TProfilePage struct {
	User TUser
}

type TQuestGiverPage struct {
	User   TUser
	Quests []TQuest
}

type TQuestPlayer struct {
	ID       uint
	Name     string
	Selected bool
}

type TQuestType struct {
	Title    string
	Value    models.QuestType
	Selected bool
}

var QuestTypes = []TQuestType{
	{
		Title: "Разовый",
		Value: models.OneTime,
	},
	{
		Title: "Ежедневный",
		Value: models.Daily,
	},
}

type TQuest struct {
	ID             uint
	Title          string
	Description    string
	Price          uint
	IsActive       bool
	Players        []TQuestPlayer
	Types          []TQuestType
	IsAllPlayers   bool
	DateStart      string
	DateEnd        string
	TimeZoneOffset int
}

type TQuestEdit struct {
	User    TUser
	Quest   TQuest
	Error   string
	Success string
}

type TQuestNew struct {
	User  TUser
	Quest TQuest
	Error string
}

type TQuestAwait struct {
	ID          uint
	Title       string
	Description string
	PlayerName  string
	Price       uint
}

type TQuestAwaitPage struct {
	User        TUser
	AwaitQuests []TQuestAwait
}

type TProfile struct {
	Score int
}

type TPlayerProfilePage struct {
	User    TUser
	Profile TProfile
}

type TPlayerQuest struct {
	ID          uint
	Title       string
	Description string
	Price       uint
	IsSended    bool
	IsRejected  bool
	IsConfirmed bool
}

type TPlayerQuestsPage struct {
	User   TUser
	Quests []TPlayerQuest
}

type TPlayerQuestPage struct {
	User    TUser
	Quest   TPlayerQuest
	Error   string
	Success string
}

type TPlayerSettingsPage struct {
	User    TUser
	Masters []TMaster
}

type TPlayerWaller struct {
	Score    int
	PlayerID uint
}
