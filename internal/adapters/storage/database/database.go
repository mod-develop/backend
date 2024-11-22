package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/mod-develop/backend/internal/models"
)

type Storage struct {
	db  *gorm.DB
	log *zap.Logger
}

type Config struct {
	DSN string `env:"DATABASE_URI"`
}

type option func(s *Storage)

func New(ctx context.Context, dsn string, options ...option) (*Storage, error) {
	var err error
	s := &Storage{
		log: zap.NewNop(),
	}
	lgr := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  true,
		},
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: lgr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed connect to database: %w", err)
	}

	s.db = db.WithContext(ctx)

	for _, opt := range options {
		opt(s)
	}

	err = s.db.AutoMigrate(
		&models.Action{},
		&models.Role{},
		&models.User{},
		&models.UserMaster{},
		&models.PlayerWallet{},
		&models.Quest{},
		&models.QuestPlayerStatus{},
	)

	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	err = s.defaultActions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed create default actions: %w", err)
	}
	err = s.defaultRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed create default roles: %w", err)
	}

	return s, nil
}

func (s *Storage) defaultRoles(ctx context.Context) error {
	err := s.db.WithContext(ctx).Save(&models.DefaultRoles).Error
	if err != nil {
		return fmt.Errorf("failed create default roles: %w", err)
	}
	return nil
}

func (s *Storage) defaultActions(ctx context.Context) error {
	err := s.db.WithContext(ctx).Save(&models.DefaultActions).Error
	if err != nil {
		return fmt.Errorf("failed create default actions: %w", err)
	}
	return nil
}

func (s *Storage) GetQuestNotPayed(ctx context.Context) (*[]models.QuestPlayerStatus, error) {
	quests := []models.QuestPlayerStatus{}
	err := s.db.WithContext(ctx).
		Where("accrual_date is NULL").
		Where("confirmation_date is not NULL").
		Preload("Quest").
		Preload("Quest.User").Preload("Quest.User.QuestMaster").Limit(100).Find(&quests).Error
	if err != nil {
		return nil, fmt.Errorf("failed get quests status: %w", err)
	}

	return &quests, nil
}
func (s *Storage) PayQuest(ctx context.Context, quest *models.QuestPlayerStatus) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		masterID := quest.Quest.User.QuestMaster.ID
		playerID := quest.PlayerID

		wallet := &models.PlayerWallet{UserMasterID: masterID, PlayerID: playerID}
		_ = tx.Where("user_master_id = ? and player_id = ?", masterID, playerID).First(wallet).Error
		wallet.Prise += int(quest.Quest.Price)
		err := tx.Save(wallet).Error
		if err != nil {
			return fmt.Errorf("failed save wallet: %w", err)
		}
		currentTime := time.Now().UTC()
		quest.AccrualDate = &currentTime
		//сбрасываем данные, что бы орм не пыталась их обновить
		quest.Quest = models.Quest{}
		quest.Player = models.User{}

		err = tx.Save(quest).Error
		if err != nil {
			return fmt.Errorf("failed save quest status as payed: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed pay by quest: %w", err)
	}

	return nil
}

func (s *Storage) NewUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	user := &models.User{
		Login:        login,
		PasswordHash: passwordHash,
	}
	err := s.db.WithContext(ctx).Save(user).Error
	if err != nil {
		return nil, fmt.Errorf("failed create user: %w", err)
	}
	return user, nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}
	err := s.db.WithContext(ctx).Where("login = ?", login).Preload("Roles").Preload("Roles.Actions").Preload("QuestMaster").First(user).Error
	if err != nil {
		return nil, fmt.Errorf("failed find user: %w", err)
	}
	return user, nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	user := &models.User{}
	err := s.db.WithContext(ctx).Where("id = ?", userID).Preload("Roles").Preload("Roles.Actions").Preload("QuestMaster").First(user).Error
	if err != nil {
		return nil, fmt.Errorf("failed find user: %w", err)
	}
	return user, nil
}

func (s *Storage) GetUsers(ctx context.Context) (*[]models.User, error) {
	users := &[]models.User{}
	err := s.db.WithContext(ctx).Find(users).Error
	if err != nil {
		return nil, fmt.Errorf("failed find users: %w", err)
	}
	return users, nil
}

func (s *Storage) NewQuest(ctx context.Context, quest *models.Quest) (*models.Quest, error) {
	err := s.db.WithContext(ctx).Save(quest).Error
	if err != nil {
		return nil, fmt.Errorf("failed create quest: %w", err)
	}
	return quest, nil
}

func (s *Storage) GetQuest(ctx context.Context, questID uint) (*models.Quest, error) {
	quest := &models.Quest{}
	err := s.db.WithContext(ctx).Where("id = ?", questID).Preload("Players").Preload("User").Preload("User.QuestMaster").First(quest).Error
	if err != nil {
		return nil, fmt.Errorf("failed get quest: %w", err)
	}
	return quest, nil
}

func (s *Storage) GetQuests(ctx context.Context, userID uint) (*[]models.Quest, error) {
	quests := []models.Quest{}
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at desc").Preload("Players").Find(&quests).Error
	if err != nil {
		return nil, fmt.Errorf("failed get quests: %w", err)
	}
	return &quests, nil
}

func (s *Storage) UpdQuest(ctx context.Context, quest *models.Quest) (*models.Quest, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("id = ?", quest.ID).Save(quest).Error
		if err != nil {
			return fmt.Errorf("failed update quest: %w", err)
		}
		err = tx.Model(quest).Association("Players").Replace(quest.Players)
		if err != nil {
			return fmt.Errorf("failed update players by quest: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed update quest: %s", err)
	}

	return quest, nil
}

func (s *Storage) GetAwaitQuests(ctx context.Context, userID uint) (*[]models.QuestPlayerStatus, error) {
	awaits := []models.QuestPlayerStatus{}
	err := s.db.WithContext(ctx).
		Joins("join quests on quests.id = quest_player_statuses.quest_id and quests.user_id = ?", userID).
		Where("confirmation_date is null").
		Where("request_execute_date > reject_execute_date or reject_execute_date is NULL").
		Preload("Player").Preload("Quest").
		Order("request_execute_date").
		Find(&awaits).Error
	if err != nil {
		return nil, fmt.Errorf("failed get await quests: %w", err)
	}

	return &awaits, nil
}

func (s *Storage) GetAwaitQUest(ctx context.Context, awaitID uint) (*models.QuestPlayerStatus, error) {
	await := &models.QuestPlayerStatus{}
	err := s.db.WithContext(ctx).Where("id = ?", awaitID).Preload("Quest").First(await).Error
	if err != nil {
		return nil, fmt.Errorf("failed get quest status: %w", err)
	}

	return await, nil
}

func (s *Storage) UpdAwaitQuest(ctx context.Context, status *models.QuestPlayerStatus) (*models.QuestPlayerStatus, error) {
	err := s.db.WithContext(ctx).Updates(status).Error
	if err != nil {
		return nil, fmt.Errorf("failed update await quest: %w", err)
	}

	return status, nil
}

func (s *Storage) NewMaster(ctx context.Context, user *models.User) (*models.UserMaster, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Save(user).Error
		if err != nil {
			return fmt.Errorf("failed create master: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed create quest master: %w", err)
	}

	return user.QuestMaster, nil
}

func (s *Storage) GetMasterByCode(ctx context.Context, code string) (*models.UserMaster, error) {
	master := &models.UserMaster{}
	err := s.db.WithContext(ctx).Where("unique_code = ?", code).First(master).Error
	if err != nil {
		return nil, fmt.Errorf("failed find master by code: %w", err)
	}
	return master, nil
}

func (s *Storage) GetMastersByPlayerID(ctx context.Context, playerID uint) (*[]models.UserMaster, error) {
	masters := []models.UserMaster{}
	err := s.db.WithContext(ctx).
		Joins("join master_players on master_players.user_master_id = user_masters.user_id").
		Where("master_players.user_id = ?", playerID).
		Find(&masters).Error
	if err != nil {
		return nil, fmt.Errorf("failed find master by code: %w", err)
	}
	return &masters, nil

}

func (s *Storage) AddPlayerForMaster(ctx context.Context, masterID, playerID uint) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		master := &models.UserMaster{}
		err := tx.Where("id = ?", masterID).First(master).Error
		if err != nil {
			return fmt.Errorf("failed get master: %w", err)
		}

		player := &models.User{}
		err = tx.Where("id = ?", playerID).First(player).Error
		if err != nil {
			return fmt.Errorf("failed get player: %w", err)
		}

		err = tx.Model(master).Association("Players").Append([]models.User{*player})
		if err != nil {
			return fmt.Errorf("failed add player for master: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed adding player for master: %w", err)
	}
	return nil
}

func (s *Storage) GetMasterByID(ctx context.Context, masterID uint) (*models.UserMaster, error) {
	master := &models.UserMaster{}
	err := s.db.WithContext(ctx).Where("id = ?", masterID).Preload("Players").First(&master).Error
	if err != nil {
		return nil, fmt.Errorf("failed get quest master: %w", err)
	}
	return master, nil
}
