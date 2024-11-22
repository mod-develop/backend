package discipline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/mod-develop/backend/internal/adapters/apperr"
	"github.com/mod-develop/backend/internal/adapters/types"
	"github.com/mod-develop/backend/internal/models"
	"github.com/mod-develop/backend/pkg/tools"
)

var (
	defaultFormDateTimeFormat = "2006-01-02T15:04"
)

type Store interface {
	NewUser(ctx context.Context, login, passwordHash string) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	GetUsers(ctx context.Context) (*[]models.User, error)
	GetMasterByID(ctx context.Context, masterID uint) (*models.UserMaster, error)
	NewQuest(ctx context.Context, quest *models.Quest) (*models.Quest, error)
	GetQuest(ctx context.Context, questID uint) (*models.Quest, error)
	GetQuests(ctx context.Context, userID uint) (*[]models.Quest, error)
	UpdQuest(ctx context.Context, quest *models.Quest) (*models.Quest, error)
	GetAwaitQuests(ctx context.Context, userID uint) (*[]models.QuestPlayerStatus, error)
	GetAwaitQUest(ctx context.Context, awaitID uint) (*models.QuestPlayerStatus, error)
	UpdAwaitQuest(ctx context.Context, status *models.QuestPlayerStatus) (*models.QuestPlayerStatus, error)
	NewMaster(ctx context.Context, master *models.User) (*models.UserMaster, error)
	AddPlayerForMaster(ctx context.Context, masterID, playerID uint) error

	GetPlayerQuests(ctx context.Context, playerID uint) (*[]models.Quest, error)
	GetPlayerQuest(ctx context.Context, questID, playerID uint) (*models.Quest, error)
	GetPlayerQuestStatus(ctx context.Context, questID, playerID uint) (*models.QuestPlayerStatus, error)
	SendPlayerQuest(ctx context.Context, questID, playerID uint) (*models.QuestPlayerStatus, error)
	UpdSendStatusPlayerQuest(ctx context.Context, status *models.QuestPlayerStatus) (*models.QuestPlayerStatus, error)
	GetMasterByCode(ctx context.Context, code string) (*models.UserMaster, error)
	GetMastersByPlayerID(ctx context.Context, playerID uint) (*[]models.UserMaster, error)
	GetWallets(ctx context.Context, playerID uint) (*[]models.PlayerWallet, error)

	GetQuestNotPayed(ctx context.Context) (*[]models.QuestPlayerStatus, error)
	PayQuest(ctx context.Context, quest *models.QuestPlayerStatus) error
}

var (
	lengthMasterCode uint = 10
)

type Discipline struct {
	store Store
	log   *zap.Logger
}

type option func(*Discipline)

func SetLogger(log *zap.Logger) option {
	return func(d *Discipline) {
		d.log = log
	}
}

func New(ctx context.Context, store Store, options ...option) (*Discipline, error) {
	d := &Discipline{
		store: store,
	}

	for _, opt := range options {
		opt(d)
	}

	go d.workerQuestPayer(ctx)

	return d, nil
}

func (s *Discipline) workerQuestPayer(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			quests, err := s.store.GetQuestNotPayed(ctx)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				s.log.Error("failed get quest not payed", zap.Error(err))
				continue
			}

			for _, quest := range *quests {
				if err := s.store.PayQuest(ctx, &quest); err != nil {
					s.log.Error("failed pay quest", zap.Error(err))
				}
			}
		}
	}
}

func (s *Discipline) Registration(ctx context.Context, login, password string) error {
	passwordHash, err := tools.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed hashing password: %w", err)
	}
	_, err = s.store.NewUser(ctx, login, passwordHash)
	if err != nil {
		return fmt.Errorf("failed create user: %w", err)
	}
	return nil
}

func (s *Discipline) Login(ctx context.Context, login, password string) (*models.User, error) {
	user, err := s.store.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, apperr.ErrDataNotFound)
		}
		return nil, fmt.Errorf("failed get user: %w", err)
	}
	if !tools.CheckPasswordHash(password, user.PasswordHash) {
		return nil, apperr.ErrDataNotFound
	}
	return user, nil
}

func (s *Discipline) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, apperr.ErrDataNotFound)
		}
		return nil, fmt.Errorf("failed get user: %w", err)
	}
	return user, nil
}

func (s *Discipline) GetPlayers(ctx context.Context, masterID uint) (*[]types.TQuestPlayer, error) {
	master, err := s.store.GetMasterByID(ctx, masterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, apperr.ErrDataNotFound)
		}
		return nil, fmt.Errorf("failed get players: %w", err)
	}

	p := []types.TQuestPlayer{}
	for _, player := range master.Players {
		p = append(p, types.TQuestPlayer{
			ID:   player.ID,
			Name: player.Login,
		})
	}
	return &p, nil
}

func (s *Discipline) NewQuest(ctx context.Context, quest *types.TQuest, userID uint) (*models.Quest, error) {
	q := &models.Quest{
		Title:       quest.Title,
		Description: quest.Description,
		Type:        models.OneTime,
		UserID:      userID,
		Price:       quest.Price,
		IsActive:    quest.IsActive,
	}
	if quest.DateStart != "" {
		if date, err := time.Parse(defaultFormDateTimeFormat, quest.DateStart); err == nil {
			date = date.Add(time.Duration(quest.TimeZoneOffset) * time.Minute)
			q.StartTime = &date
		}
	}
	if quest.DateEnd != "" {
		if date, err := time.Parse(defaultFormDateTimeFormat, quest.DateEnd); err == nil {
			date = date.Add(time.Duration(quest.TimeZoneOffset) * time.Minute)
			q.EndTime = &date
		}
	}

	for _, t := range quest.Types {
		if t.Selected {
			q.Type = t.Value
		}
	}
	for _, t := range quest.Players {
		if t.Selected {
			q.Players = append(q.Players, models.User{ID: t.ID})
		}
	}
	q, err := s.store.NewQuest(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed create quest: %w", err)
	}
	return q, nil
}

func (s *Discipline) GetQuest(ctx context.Context, questID uint) (*types.TQuest, error) {
	quest, err := s.store.GetQuest(ctx, questID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, apperr.ErrDataNotFound)
		}
		return nil, fmt.Errorf("failed get quest: %w", err)
	}

	players, err := s.GetPlayers(ctx, quest.User.QuestMaster.ID)
	if err != nil {
		return nil, fmt.Errorf("failed get players:  %w", err)
	}

	q := &types.TQuest{
		ID:          quest.ID,
		Title:       quest.Title,
		Description: quest.Description,
		IsActive:    quest.IsActive,
		Price:       quest.Price,
		Types:       types.QuestTypes,
		// Players:     *players,
	}
	if quest.StartTime != nil {
		q.DateStart = quest.StartTime.UTC().Format(defaultFormDateTimeFormat)
	}
	if quest.EndTime != nil {
		q.DateEnd = quest.EndTime.UTC().Format(defaultFormDateTimeFormat)
	}

	q.IsAllPlayers = len(quest.Players) == 0
	for _, p := range *players {
		if !q.IsAllPlayers {
			for _, q := range quest.Players {
				if q.ID == p.ID {
					p.Selected = true
					break
				}
			}
		}
		q.Players = append(q.Players, p)
	}
	for i := range q.Types {
		q.Types[i].Selected = quest.Type == q.Types[i].Value
	}

	return q, nil
}

func (s *Discipline) GetQuests(ctx context.Context, userID uint) (*[]types.TQuest, error) {
	qs, err := s.store.GetQuests(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &[]types.TQuest{}, nil
		}
		return nil, fmt.Errorf("failed get quests: %w", err)
	}
	quests := []types.TQuest{}
	for _, q := range *qs {
		upQ := types.TQuest{
			ID:          q.ID,
			Title:       q.Title,
			Description: q.Description,
			Price:       q.Price,
			IsActive:    q.IsActive,
			Players:     []types.TQuestPlayer{},
		}
		if q.StartTime != nil {
			upQ.DateStart = q.StartTime.UTC().Format(defaultFormDateTimeFormat)
		}
		if q.EndTime != nil {
			upQ.DateEnd = q.EndTime.UTC().Format(defaultFormDateTimeFormat)
		}

		if len(q.Players) == 0 {
			upQ.IsAllPlayers = true
		} else {
			for _, p := range q.Players {
				upQ.Players = append(upQ.Players, types.TQuestPlayer{
					ID:       p.ID,
					Name:     p.Login,
					Selected: true,
				})
			}
		}
		for _, t := range types.QuestTypes {
			t.Selected = t.Value == q.Type
			upQ.Types = append(upQ.Types, t)
		}
		quests = append(quests, upQ)
	}
	return &quests, apperr.ErrDataNotFound
}

func (s *Discipline) EditQuest(ctx context.Context, quest *types.TQuest, userID uint) (*types.TQuest, error) {
	q := &models.Quest{
		Model:       gorm.Model{ID: quest.ID},
		Title:       quest.Title,
		Description: quest.Description,
		UserID:      userID,
		Price:       quest.Price,
		IsActive:    quest.IsActive,
	}
	if quest.DateStart != "" {
		if date, err := time.Parse(defaultFormDateTimeFormat, quest.DateStart); err == nil {
			date = date.Add(time.Duration(quest.TimeZoneOffset) * time.Minute)
			q.StartTime = &date
		}
	} else {
		q.StartTime = nil
	}
	if quest.DateEnd != "" {
		if date, err := time.Parse(defaultFormDateTimeFormat, quest.DateEnd); err == nil {
			date = date.Add(time.Duration(quest.TimeZoneOffset) * time.Minute)
			q.EndTime = &date
		}
	} else {
		q.EndTime = nil
	}
	for _, t := range quest.Types {
		if t.Selected {
			q.Type = t.Value
			break
		}
	}
	for _, p := range quest.Players {
		if p.Selected {
			q.Players = append(q.Players, models.User{ID: p.ID})
		}
	}

	q, err := s.store.UpdQuest(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed update quest: %w", err)
	}

	quest.Title = q.Title

	return quest, nil
}

func (s *Discipline) GetAwaitQuests(ctx context.Context, userID uint) (*[]types.TQuestAwait, error) {
	result := []types.TQuestAwait{}
	awaits, err := s.store.GetAwaitQuests(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed get await quests: %w", err)
	}
	for _, s := range *awaits {
		result = append(result, types.TQuestAwait{
			ID:          s.ID,
			Title:       s.Quest.Title,
			Description: s.Quest.Description,
			PlayerName:  s.Player.Login,
			Price:       s.Quest.Price,
		})
	}

	return &result, nil
}

func (s *Discipline) ManageQuestConfirmation(ctx context.Context, statusID, userID uint, confirm bool) error {
	status, err := s.store.GetAwaitQUest(ctx, statusID)
	if err != nil {
		return fmt.Errorf("failed get await quest: %w", err)
	}

	if status.Quest.UserID != userID {
		return apperr.ErrDataNotFound
	}

	currentTime := time.Now().UTC()
	if confirm {
		status.ConfirmationDate = &currentTime
	} else {
		status.RejectExecuteDate = &currentTime
	}

	_, err = s.store.UpdAwaitQuest(ctx, status)
	if err != nil {
		return fmt.Errorf("failed update status quest: %w", err)
	}

	return nil
}

func (s *Discipline) ManageCreateSelfMaster(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed get user by id: %w", err)
	}
	user.Roles = append(user.Roles, models.RoleQuestMasterObject)
	code := tools.RandomString(lengthMasterCode)
	user.QuestMaster = &models.UserMaster{
		UserID:     userID,
		UniqueCode: code,
	}
	master, err := s.store.NewMaster(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed create quest master: %w", err)
	}

	user, err = s.store.GetUserByID(ctx, master.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed get user: %w", err)
	}

	return user, nil
}

func (s *Discipline) GetQuestsPlayer(ctx context.Context, playerID uint) (*[]types.TPlayerQuest, error) {
	quests := []types.TPlayerQuest{}
	data, err := s.store.GetPlayerQuests(ctx, playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &quests, nil
		}
		return nil, fmt.Errorf("failed get quests by player `%v`: %w", playerID, err)
	}

	for _, q := range *data {
		quest := types.TPlayerQuest{
			ID:          q.ID,
			Title:       q.Title,
			Description: q.Description,
			Price:       q.Price,
		}
		status, err := s.store.GetPlayerQuestStatus(ctx, q.ID, playerID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed get quest player: %w", err)
		}
		if status != nil {
			quest.IsSended = status.RequestExecuteDate != nil
			quest.IsRejected = status.RejectExecuteDate != nil
			quest.IsConfirmed = status.ConfirmationDate != nil

			if status.RejectExecuteDate != nil && status.RequestExecuteDate != nil {
				quest.IsSended = status.RequestExecuteDate.After(*status.RejectExecuteDate)
				quest.IsRejected = status.RejectExecuteDate.After(*status.RequestExecuteDate)
			}
		}

		quests = append(quests, quest)
	}

	return &quests, nil
}

func (s *Discipline) AddMaster(ctx context.Context, masterCode string, playerID uint) error {
	master, err := s.store.GetMasterByCode(ctx, masterCode)
	if err != nil {
		return fmt.Errorf("failed get master by code `%s`: %w", masterCode, err)
	}

	err = s.store.AddPlayerForMaster(ctx, master.ID, playerID)
	if err != nil {
		return fmt.Errorf("failed add player for master: %w", err)
	}

	return nil
}

func (s *Discipline) GetQuestPlayer(ctx context.Context, questID uint, playerID uint) (*types.TPlayerQuest, error) {
	quest, err := s.store.GetPlayerQuest(ctx, questID, playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, apperr.ErrDataNotFound)
		}
		return nil, fmt.Errorf("failed get quest player: %w", err)
	}

	playerQuests := &types.TPlayerQuest{
		ID:          quest.ID,
		Title:       quest.Title,
		Description: quest.Description,
		Price:       quest.Price,
	}

	status, err := s.store.GetPlayerQuestStatus(ctx, questID, playerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed get quest player: %w", err)
	}
	if status != nil {
		playerQuests.IsSended = status.RequestExecuteDate != nil
		playerQuests.IsRejected = status.RejectExecuteDate != nil
		playerQuests.IsConfirmed = status.ConfirmationDate != nil

		if status.RejectExecuteDate != nil && status.RequestExecuteDate != nil {
			playerQuests.IsSended = status.RequestExecuteDate.After(*status.RejectExecuteDate)
			playerQuests.IsRejected = status.RejectExecuteDate.After(*status.RequestExecuteDate)
		}
	}

	return playerQuests, nil
}

func (s *Discipline) SendQuestPlayer(ctx context.Context, questID, playerID uint) (*types.TPlayerQuest, error) {
	status, err := s.store.GetPlayerQuestStatus(ctx, questID, playerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed get quest player: %w", err)
	}

	if status != nil && status.ID != 0 {
		if status.RejectExecuteDate != nil && status.RequestExecuteDate != nil {
			if status.RejectExecuteDate.Before(*status.RequestExecuteDate) {
				return nil, apperr.ErrPlayerQuestStatusExists
			}
		}
		status, err = s.store.UpdSendStatusPlayerQuest(ctx, status)
		if err != nil {
			return nil, fmt.Errorf("failed update send status quest: %w", err)
		}
	} else {
		status, err = s.store.SendPlayerQuest(ctx, questID, playerID)
		if err != nil {
			return nil, fmt.Errorf("failed send quest player: %w", err)
		}
	}

	result := &types.TPlayerQuest{
		ID:          status.Quest.ID,
		Title:       status.Quest.Title,
		Description: status.Quest.Description,
		Price:       status.Quest.Price,
		IsSended:    true,
	}

	return result, nil
}

func (s *Discipline) GetPlayerMasters(ctx context.Context, playerID uint) (*[]models.UserMaster, error) {
	masters, err := s.store.GetMastersByPlayerID(ctx, playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &[]models.UserMaster{}, nil
		}
		return nil, fmt.Errorf("failed get master by player: %w", err)
	}

	return masters, nil
}

func (s *Discipline) GetWallets(ctx context.Context, playerID uint) (*[]types.TPlayerWaller, error) {
	wallets, err := s.store.GetWallets(ctx, playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &[]types.TPlayerWaller{}, errors.Join(err, apperr.ErrDataNotFound)
		}
		return nil, fmt.Errorf("failed get wallet: %w", err)
	}
	result := []types.TPlayerWaller{}
	for _, w := range *wallets {
		result = append(result, types.TPlayerWaller{
			Score:    w.Prise,
			PlayerID: w.PlayerID,
		})
	}
	return &result, nil
}
