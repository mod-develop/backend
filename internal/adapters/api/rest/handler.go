package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mod-develop/backend/internal/adapters/apperr"
	"github.com/mod-develop/backend/internal/adapters/types"
	"github.com/mod-develop/backend/internal/models"
)

func (s *Server) handlerRegistrationPage(c *gin.Context) {
	err := s.ui.RegistrationPage(c.Writer)
	if err != nil {
		s.log.Error("failed create registration page", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}
}

func (s *Server) handlerApiRegistration(c *gin.Context) {
	bBody, _ := s.readBody(c)
	jBody := tRequestRegistration{}
	err := json.Unmarshal(bBody, &jBody)
	if err != nil {
		s.log.Error("failed parse body", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.disc.Registration(c.Request.Context(), jBody.Login, jBody.Password)
	if err != nil {
		s.log.Error("failed register user", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": true,
	})
}

func (s *Server) handlerAuthorization(c *gin.Context) {
	err := s.ui.AuthorizationPage(c.Writer)
	if err != nil {
		s.log.Error("failed create registration page", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}
}

func (s *Server) handlerAPIAuthorization(c *gin.Context) {
	bBody, _ := s.readBody(c)
	jBody := tRequestAuthorization{}
	err := json.Unmarshal(bBody, &jBody)
	if err != nil {
		s.log.Error("failed parse body", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.authorization(c, jBody.Login, jBody.Password)
	if err != nil {
		if errors.Is(err, apperr.ErrDataNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Не удалось авторизоваться",
			})
			return
		}
		s.log.Error("failed register user", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func (s *Server) handlerAPIManageQuestConfirmation(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	bBody, _ := s.readBody(c)
	jBody := tRequestApiManageQuestConfirmation{}
	err = json.Unmarshal(bBody, &jBody)
	if err != nil {
		s.log.Error("failed parse body", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.disc.ManageQuestConfirmation(c.Request.Context(), jBody.ID, user.ID, jBody.Action == actionConfirmationAccpet)
	if err != nil {
		s.log.Error("failed confirmation quest", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func (s *Server) handlerAPIManageCreateMaster(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	userM, err := s.disc.ManageCreateSelfMaster(c.Request.Context(), user.ID)
	if err != nil {
		s.log.Error("failed confirmation quest", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.sess.SaveUser(c, userM)
	if err != nil {
		s.log.Error("failed save user session", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"masterID":   userM.QuestMaster.ID,
		"masterCode": userM.QuestMaster.UniqueCode,
	})
}

func (s *Server) handlerAPIPlayerAddMaster(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	bBody, _ := s.readBody(c)
	jBody := tRequestAPISettingsAddMaster{}
	err = json.Unmarshal(bBody, &jBody)
	if err != nil {
		s.log.Error("failed parse body", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.disc.AddMaster(c.Request.Context(), jBody.Code, user.ID)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func (s *Server) handleUserLogout(c *gin.Context) {
	unauthorize(c)

	c.Redirect(http.StatusTemporaryRedirect, "/authorization")
}

func (s *Server) handlerAPIUserInfo(c *gin.Context) {
	user, statusCode, err := s.checkUser(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if statusCode > 0 {
		c.Writer.WriteHeader(statusCode)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"userID": user.ID,
		"login":  user.Login,
	})
}

func (s *Server) handlerUserProfile(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}

	err = s.ui.ProfilePage(c.Writer, &types.TProfilePage{User: *user})
	if err != nil {
		s.log.Error("page", zap.Error(err))
	}
}

func (s *Server) handlerQuestGiverPage(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}

	quests, err := s.disc.GetQuests(c.Request.Context(), user.ID)
	if err != nil && !errors.Is(err, apperr.ErrDataNotFound) {
		s.log.Error("failed get user from session", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}

	err = s.ui.QuestGiverPage(c.Writer, &types.TQuestGiverPage{
		User:   *user,
		Quests: *quests,
	})
	if err != nil {
		s.log.Error("page", zap.Error(err))
	}
}

func (s *Server) handlerQuestEdit(c *gin.Context) {
	sID := c.Param("id")
	questID, err := strconv.Atoi(sID)
	if err != nil {
		s.log.Error("invalid id quest", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}

	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}

	page := &types.TQuestEdit{
		User: *user,
		Quest: types.TQuest{
			ID:    uint(questID),
			Types: types.QuestTypes,
		},
	}

	players, err := s.disc.GetPlayers(c.Request.Context(), user.Master.ID)
	if err != nil {
		s.log.Error("failed get players", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}
	page.Quest.Players = *players

	if c.Request.Method == http.MethodPost {
		page.Quest.Title = c.PostForm("title")
		page.Quest.Description = c.PostForm("description")
		for i := range page.Quest.Types {
			page.Quest.Types[i].Selected = page.Quest.Types[i].Value == models.QuestType(c.PostForm("type"))
		}
		page.Quest.IsActive = c.PostForm("active") == "on"
		page.Quest.IsAllPlayers = c.PostForm("players_all") == "on"
		for _, p := range c.PostFormArray("players") {
			for i := range page.Quest.Players {
				if strconv.Itoa(int(page.Quest.Players[i].ID)) == p || c.PostForm("players_all") == "on" {
					page.Quest.Players[i].Selected = true
					break
				}
			}
		}
		price, _ := strconv.Atoi(c.PostForm("price"))
		page.Quest.Price = uint(price)
		page.Quest.DateStart = c.PostForm("date_start")
		page.Quest.DateEnd = c.PostForm("date_end")
		page.Quest.TimeZoneOffset, _ = strconv.Atoi(c.PostForm("timezoneoffset"))

		s.log.Debug("form",
			zap.String("title", c.PostForm("title")),
			zap.String("description", c.PostForm("description")),
			zap.String("active", c.PostForm("active")),
			zap.String("players_all", c.PostForm("players_all")),
			zap.String("price", c.PostForm("price")),
			zap.Strings("players", c.PostFormArray("players")),
			zap.String("date_start", c.PostForm("date_start")),
			zap.String("date_end", c.PostForm("date_end")),
			zap.String("type", c.PostForm("type")),
			zap.String("timezoneoffset", c.PostForm("timezoneoffset")),
		)

		quest, err := s.disc.EditQuest(c.Request.Context(), &page.Quest, user.ID)
		if err != nil {
			s.log.Error("create quest", zap.Error(err))
			page.Error = "Не удалось создать квест"
		}
		page.Quest = *quest
		page.Success = "Квест обновлен"

	} else {
		quest, err := s.disc.GetQuest(c.Request.Context(), uint(questID))
		if err != nil {
			s.log.Error("failed get quest", zap.Error(err), zap.Int("quest_id", questID))
			_ = s.ui.Error500Page(c.Writer)
			return
		}
		page.Quest = *quest
	}

	err = s.ui.QuestEdit(c.Writer, page)
	if err != nil {
		s.log.Error("page", zap.Error(err))
	}
}

func (s *Server) handlerQuestNew(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}

	players, err := s.disc.GetPlayers(c.Request.Context(), user.Master.ID)
	if err != nil {
		s.log.Error("failed get players", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}
	page := &types.TQuestNew{
		User: *user,
		Quest: types.TQuest{
			Types:        types.QuestTypes,
			IsAllPlayers: true,
			Players:      *players,
		},
	}

	if c.Request.Method == http.MethodPost {
		page.Quest.Title = c.PostForm("title")
		page.Quest.Description = c.PostForm("description")
		for i := range page.Quest.Types {
			page.Quest.Types[i].Selected = page.Quest.Types[i].Value == models.QuestType(c.PostForm("type"))
		}
		page.Quest.IsActive = c.PostForm("active") == "on"
		page.Quest.IsAllPlayers = c.PostForm("players_all") == "on"
		for _, p := range c.PostFormArray("players") {
			for i := range page.Quest.Players {
				if strconv.Itoa(int(page.Quest.Players[i].ID)) == p || c.PostForm("players_all") == "on" {
					page.Quest.Players[i].Selected = true
					break
				}
			}
		}
		price, _ := strconv.Atoi(c.PostForm("price"))
		page.Quest.Price = uint(price)
		page.Quest.DateStart = c.PostForm("date_start")
		page.Quest.DateEnd = c.PostForm("date_end")
		page.Quest.TimeZoneOffset, _ = strconv.Atoi(c.PostForm("timezoneoffset"))

		s.log.Debug("form",
			zap.String("title", c.PostForm("title")),
			zap.String("description", c.PostForm("description")),
			zap.String("active", c.PostForm("active")),
			zap.String("players_all", c.PostForm("players_all")),
			zap.String("price", c.PostForm("price")),
			zap.Strings("players", c.PostFormArray("players")),
			zap.String("date_start", c.PostForm("date_start")),
			zap.String("date_end", c.PostForm("date_end")),
			zap.String("type", c.PostForm("type")),
			zap.String("timezoneoffset", c.PostForm("timezoneoffset")),
		)

		quest, err := s.disc.NewQuest(c.Request.Context(), &page.Quest, user.ID)
		if err != nil {
			s.log.Error("create quest", zap.Error(err))
			page.Error = "Не удалось создать квест"
		}

		if err == nil {
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/manager/quests/%v", quest.ID))
			return
		}
	}

	err = s.ui.QuestNew(c.Writer, page)
	if err != nil {
		s.log.Error("page QuestNew", zap.Error(err))
	}
}

func (s *Server) handlerQuestAwait(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	awaits, err := s.disc.GetAwaitQuests(c.Request.Context(), user.ID)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.ui.QuestAwait(c.Writer, &types.TQuestAwaitPage{
		User:        *user,
		AwaitQuests: *awaits,
	})
	if err != nil {
		s.log.Error("page handlerQuestAwait", zap.Error(err))
	}
}

func (s *Server) handlerPlayer(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	wallets, err := s.disc.GetWallets(c.Request.Context(), user.ID)
	if err != nil && !errors.Is(err, apperr.ErrDataNotFound) {
		s.log.Error("failed get wallet", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	score := 0
	if wallets != nil {
		for _, w := range *wallets {
			score += w.Score
		}
	}

	err = s.ui.PlayerProfile(c.Writer, &types.TPlayerProfilePage{
		User: *user,
		Profile: types.TProfile{
			Score: score,
		},
	})
	if err != nil {
		s.log.Error("page handlerPlayer", zap.Error(err))
	}
}

func (s *Server) handlerPlayerQuests(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	quests, err := s.disc.GetQuestsPlayer(c.Request.Context(), user.ID)
	if err != nil {
		s.log.Error("failed get quests", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.ui.PlayerQuests(c.Writer, &types.TPlayerQuestsPage{
		User:   *user,
		Quests: *quests,
	})
	if err != nil {
		s.log.Error("page GetQuestsPlayer", zap.Error(err))
	}
}

func (s *Server) handlerPlayerQuest(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	sID := c.Param("id")
	questID, err := strconv.Atoi(sID)
	if err != nil {
		s.log.Error("invalid id quest", zap.Error(err))
		_ = s.ui.Error500Page(c.Writer)
		return
	}

	page := &types.TPlayerQuestPage{
		User: *user,
	}

	if c.Request.Method == http.MethodPost {
		if c.PostForm("action") == "send" {
			_, err = s.disc.SendQuestPlayer(c.Request.Context(), uint(questID), user.ID)
			if err != nil && !errors.Is(err, apperr.ErrPlayerQuestStatusExists) {
				if errors.Is(err, apperr.ErrDataNotFound) {
					c.Writer.WriteHeader(http.StatusNotFound)
					return
				}
				s.log.Error("failed get quests", zap.Error(err))
				c.Writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err != nil && errors.Is(err, apperr.ErrPlayerQuestStatusExists) {
				page.Error = "Квест уже был отправлен"
			} else {
				page.Success = "Квест отправлен"
			}
		}
	}

	quest, err := s.disc.GetQuestPlayer(c.Request.Context(), uint(questID), user.ID)
	if err != nil {
		if errors.Is(err, apperr.ErrDataNotFound) {
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		s.log.Error("failed get quests", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	page.Quest = *quest

	err = s.ui.PlayerQuest(c.Writer, page)
	if err != nil {
		s.log.Error("page QuestNew", zap.Error(err))
	}
}

func (s *Server) handlerPlayerSettings(c *gin.Context) {
	user, err := s.sess.GetUser(c)
	if err != nil {
		s.log.Error("failed get user from session", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	masters, err := s.disc.GetPlayerMasters(c.Request.Context(), user.ID)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	playerMasters := []types.TMaster{}
	for _, m := range *masters {
		playerMasters = append(playerMasters, types.TMaster{
			ID: m.ID,
		})
	}

	err = s.ui.PlayerSettings(c.Writer, &types.TPlayerSettingsPage{
		User:    *user,
		Masters: playerMasters,
	})
	if err != nil {
		s.log.Error("page PlayerSettings", zap.Error(err))
	}
}
