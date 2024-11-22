package database

import (
	"context"
	"fmt"
	"time"

	"github.com/mod-develop/backend/internal/models"
)

func (s *Storage) GetPlayerQuests(ctx context.Context, playerID uint) (*[]models.Quest, error) {
	quests := []models.Quest{}
	err := s.db.WithContext(ctx).Raw("? UNION ?",
		s.db.Joins("left join quest_players qp on qp.quest_id = quests.id").
			Joins("join master_players mp on mp.user_id = ? and mp.user_id = qp.user_id ", playerID).
			Joins("join user_masters um on um.id = mp.user_master_id and um.user_id = quests.user_id ").
			Joins("left join quest_player_statuses qps on qps.quest_id = quests.id and qps.player_id = mp.user_id").
			Where("start_time is null or start_time <= ?", time.Now().UTC()).
			Where("end_time is null or end_time >= ?", time.Now().UTC()).
			Where("quests.is_active = true").
			Where("qps.accrual_date is NULL").
			Group("quests.id").
			Find(&[]models.Quest{}),
		s.db.Joins("left join quest_players qp on qp.quest_id = quests.id").
			Joins("join user_masters um on um.user_id = quests.user_id").
			Joins("join master_players mp on mp.user_master_id = um.id  and mp.user_id = ?", playerID).
			Joins("left join quest_player_statuses qps on qps.quest_id = quests.id and qps.player_id = mp.user_id").
			Where("start_time is null or start_time <= ?", time.Now().UTC()).
			Where("end_time is null or end_time >= ?", time.Now().UTC()).
			Where("quests.is_active = true").
			Where("qp.user_id is NULL").
			Where("qps.accrual_date is NULL").
			Group("quests.id").
			Find(&[]models.Quest{}),
	).Order("updated_at desc").Find(&quests).Error
	if err != nil {
		return nil, fmt.Errorf("failed find quests: %w", err)
	}
	fmt.Println(time.Now().UTC())

	return &quests, nil
}

func (s *Storage) GetPlayerQuest(ctx context.Context, questID, playerID uint) (*models.Quest, error) {
	quest := models.Quest{}
	err := s.db.WithContext(ctx).Raw("? UNION ?",
		s.db.Joins("left join quest_players qp on qp.quest_id = quests.id").
			Joins("join master_players mp on mp.user_id = ? and mp.user_id = qp.user_id ", playerID).
			Joins("join user_masters um on um.id = mp.user_master_id and um.user_id = quests.user_id ").
			Joins("left join quest_player_statuses qps on qps.quest_id = quests.id and qps.player_id = mp.user_id").
			Where("start_time is null or start_time <= ?", time.Now().UTC()).
			Where("end_time is null or end_time >= ?", time.Now().UTC()).
			Where("quests.is_active = true and quests.id = ?", questID).
			Where("qps.accrual_date is NULL").
			Group("quests.id").
			Find(&models.Quest{}),
		s.db.Joins("left join quest_players qp on qp.quest_id = quests.id").
			Joins("join user_masters um on um.user_id = quests.user_id").
			Joins("join master_players mp on mp.user_master_id = um.id and mp.user_id = ?", playerID).
			Joins("left join quest_player_statuses qps on qps.quest_id = quests.id and qps.player_id = mp.user_id").
			Where("start_time is null or start_time <= ?", time.Now().UTC()).
			Where("end_time is null or end_time >= ?", time.Now().UTC()).
			Where("quests.is_active = true and quests.id = ?", questID).
			Where("qp.user_id is NULL").
			Where("qps.accrual_date is NULL").
			Group("quests.id").
			Find(&models.Quest{}),
	).First(&quest).Error
	if err != nil {
		return nil, fmt.Errorf("failed find quest by player: %w", err)
	}

	return &quest, nil
}

func (s *Storage) GetPlayerQuestStatus(ctx context.Context, questID, playerID uint) (*models.QuestPlayerStatus, error) {
	status := &models.QuestPlayerStatus{}
	err := s.db.WithContext(ctx).
		Where("quest_id = ? and player_id = ?", questID, playerID).Preload("Quest").First(status).Error
	if err != nil {
		return nil, fmt.Errorf("failed find quest player status: %w", err)
	}

	return status, nil
}

func (s *Storage) SendPlayerQuest(ctx context.Context, questID, playerID uint) (*models.QuestPlayerStatus, error) {
	current := time.Now().UTC()
	status := &models.QuestPlayerStatus{
		QuestID:            questID,
		PlayerID:           playerID,
		RequestExecuteDate: &current,
	}
	err := s.db.WithContext(ctx).Save(status).Error
	if err != nil {
		return nil, fmt.Errorf("failed send quest status: %w", err)
	}

	status, err = s.GetPlayerQuestStatus(ctx, questID, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed get quest status: %w", err)
	}

	return status, nil
}

func (s *Storage) UpdSendStatusPlayerQuest(ctx context.Context, status *models.QuestPlayerStatus) (*models.QuestPlayerStatus, error) {
	currentTime := time.Now().UTC()
	status.RequestExecuteDate = &currentTime
	err := s.db.WithContext(ctx).Updates(status).Error
	if err != nil {
		return nil, fmt.Errorf("failed update quest send status: %w", err)
	}
	return status, nil
}

func (s *Storage) GetWallets(ctx context.Context, playerID uint) (*[]models.PlayerWallet, error) {
	wallets := []models.PlayerWallet{}
	err := s.db.WithContext(ctx).Find(&wallets, "player_id = ?", playerID).Error
	if err != nil {
		return nil, fmt.Errorf("failed get wallets: %w", err)
	}
	return &wallets, nil
}
