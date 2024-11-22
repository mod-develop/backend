package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mod-develop/backend/internal/models"
	"github.com/mod-develop/backend/pkg/jwt"
)

// Logger middleware логирования.
func (s *Server) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		s.log.Info(
			"Request information",
			zap.String("uri", c.Request.RequestURI),
			zap.Duration("duration", time.Since(start)),
			zap.String("method", c.Request.Method),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}

func (s *Server) middlewareErrorPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() == 401 {
			_ = s.ui.AuthorizationPage(c.Writer)
			return
		}
		if c.Writer.Status() == 403 {
			_ = s.ui.Error403Page(c.Writer)
			return
		}
		if c.Writer.Status() == 404 {
			_ = s.ui.Error404Page(c.Writer)
			return
		}
		if c.Writer.Status() == 500 {
			_ = s.ui.Error500Page(c.Writer)
			return
		}
	}
}

func (s *Server) middlewareAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := s.checkAuth(c)
		if err != nil {
			if !errors.Is(err, errUnauthorize) {
				c.Writer.WriteHeader(http.StatusInternalServerError)
			} else {
				s.ui.AuthorizationPage(c.Writer)
			}
			c.Abort()
		}

		if err == nil && userID == 0 {
			s.ui.AuthorizationPage(c.Writer)
			c.Abort()
		}

		c.Next()
	}
}

func (s *Server) middlewareManagerRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := s.sess.GetUser(c)
		if err != nil {
			s.log.Error("failed get user from session", zap.Error(err))
			c.Abort()
		}
		if !user.IsAdmin && !user.IsQuestMaster {
			// _ = s.ui.Error403Page(c.Writer)
			c.Writer.WriteHeader(http.StatusForbidden)
			c.Abort()
		}

		c.Next()
	}
}

func (s *Server) middlewareAuthenticationAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := s.checkAuth(c)
		if err != nil {
			if !errors.Is(err, errUnauthorize) {
				c.Writer.WriteHeader(http.StatusInternalServerError)
			} else {
				c.Writer.WriteHeader(http.StatusUnauthorized)
			}
			c.Abort()
		}

		if err == nil && userID == 0 {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Abort()
		}

		c.Next()
	}
}

func (s *Server) checkAuth(c *gin.Context) (userID uint, err error) {
	var ok bool
	var userIDS string
	cookieUserID, err := c.Request.Cookie(cookieName)
	if err != nil {
		return 0, fmt.Errorf("failed reade user cookie: %w %w", err, errUnauthorize)
	}

	jwtRest := jwt.New([]byte(s.secretKey))
	userIDS, ok, err = jwtRest.Verify(cookieUserID.Value, cookieKey)
	if err != nil {
		return 0, fmt.Errorf("failed verify token: %w %w", err, errUnauthorize)
	}

	if !ok {
		return 0, fmt.Errorf("unverify usercookie: %w", errUnauthorize)
	}

	userID64, err := strconv.ParseUint(userIDS, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("can't convert string userID to uint: %w", err)
	}

	return uint(userID64), nil
}

func (s *Server) checkUser(c *gin.Context) (user *models.User, statusCode int, err error) {
	userID, err := s.checkAuth(c)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("failed check cookie: %w", err)
	}
	user, err = s.disc.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed get user: %w", err)
	}
	if user.ID == 0 {
		return nil, http.StatusUnauthorized, errors.New("not found user")
	}
	return user, 0, nil
}
