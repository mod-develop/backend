package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleAdmin       = "admin"
	RoleQuestMaster = "quest_master"

	ActionCreateQuest = "create_quest"
)

var DefaultRoles = []Role{
	{
		Model: gorm.Model{
			ID: 1,
		},
		Name:    RoleAdmin,
		Actions: []Action{},
	},
	{
		Model: gorm.Model{
			ID: 2,
		},
		Name: RoleQuestMaster,
		Actions: []Action{
			{Model: gorm.Model{ID: 1}},
		},
	},
}

var DefaultActions = []Action{
	{
		Model: gorm.Model{
			ID: 1,
		},
		Name: ActionCreateQuest,
	},
}

type User struct {
	ID           uint   `gorm:"primarykey"`
	Login        string `gorm:"index"`
	PasswordHash string
	Roles        []Role `gorm:"many2many:user_roles;"`
	QuestMaster  *UserMaster
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role struct {
	gorm.Model
	Actions []Action `gorm:"many2many:role_actions;"`
	Name    string
}

type Action struct {
	gorm.Model
	Name string
}

type UserMaster struct {
	gorm.Model
	UserID     uint   `gorm:"unique:uq_user"`
	UniqueCode string `gorm:"unique:uq_code"`
	Players    []User `gorm:"many2many:master_players;constraint:OnDelete:CASCADE"`
}

type PlayerWallet struct {
	gorm.Model
	UserMasterID uint
	UserMaster   UserMaster
	PlayerID     uint
	Player       User
	Prise        int
}

type QuestType string

const (
	OneTime QuestType = "one_time"
	Daily   QuestType = "daily"
)

type Quest struct {
	gorm.Model
	Title       string
	Description string
	Type        QuestType `sql:"type:enum('one_time','daily')"`
	UserID      uint
	User        User
	Players     []User `gorm:"many2many:quest_players;constraint:OnDelete:CASCADE"`
	Price       uint
	StartTime   *time.Time
	EndTime     *time.Time
	IsActive    bool
}

type QuestPlayerStatus struct {
	gorm.Model
	PlayerID           uint `gorm:"index:idx_player_id"`
	Player             User
	QuestID            uint `gorm:"index:idx_quest_id"`
	Quest              Quest
	RequestExecuteDate *time.Time
	RejectExecuteDate  *time.Time
	ConfirmationDate   *time.Time
	AccrualDate        *time.Time
}
