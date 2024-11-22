package cookie

import (
	"encoding/gob"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/mod-develop/backend/internal/adapters/types"
	"github.com/mod-develop/backend/internal/models"
)

type Cookie struct {
	store cookie.Store
}

func init() {
	gob.Register(types.TUser{})
}

func New(secret []byte) (*Cookie, error) {
	c := &Cookie{
		store: cookie.NewStore(secret),
	}

	return c, nil
}

func (c *Cookie) Middleware() gin.HandlerFunc {
	return sessions.Sessions("medalofdescipline", c.store)
}

func (s *Cookie) SaveUser(c *gin.Context, user *models.User) error {
	sUser := types.TUser{
		ID:    user.ID,
		Login: user.Login,
	}

	session := sessions.Default(c)
	for _, role := range user.Roles {
		if role.Name == models.RoleAdmin {
			sUser.IsAdmin = true
		}
		if role.Name == models.RoleQuestMaster {
			sUser.IsQuestMaster = true
			if user.QuestMaster != nil {
				sUser.Master.ID = user.QuestMaster.ID
				sUser.Master.Code = user.QuestMaster.UniqueCode
			}
		}
		for _, action := range role.Actions {
			if action.Name == models.ActionCreateQuest {
				sUser.Action.IsQuestCreater = true
			}
		}
	}
	session.Set("user", sUser)
	err := session.Save()
	if err != nil {
		return fmt.Errorf("failed save user into session: %w", err)
	}
	return nil
}

func (s *Cookie) GetUser(c *gin.Context) (*types.TUser, error) {
	session := sessions.Default(c)
	if user, ok := session.Get("user").(types.TUser); ok {
		return &user, nil
	}
	return nil, fmt.Errorf("failed get user from session")
}
