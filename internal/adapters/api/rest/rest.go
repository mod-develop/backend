// Модуль rest предоставляет http сервер и методы взаимодействия с REST API.
package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playmixer/medal-of-discipline/internal/adapters/types"
	"github.com/playmixer/medal-of-discipline/internal/models"
	"github.com/playmixer/medal-of-discipline/pkg/jwt"
	"go.uber.org/zap"
)

var (
	cookieName = "token"
	cookieKey  = "UserID"

	msgErrorCloseBody = "failed close body request"

	errUnauthorize = errors.New("unauthorize")
)

// Константы сервиса.
const (
	ContentLength   string = "Content-Length"   // заголовок длины конетента
	ContentType     string = "Content-Type"     // заколовок типа контент
	ApplicationJSON string = "application/json" // json контент

	CookieNameUserID string = "token" // поле хранения токента
)

var (
	shutdownDelay = time.Second * 5
)

type discipline interface {
	Registration(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	GetPlayers(ctx context.Context, masterID uint) (*[]types.TQuestPlayer, error)
	NewQuest(ctx context.Context, quest *types.TQuest, userID uint) (*models.Quest, error)
	GetQuest(ctx context.Context, questID uint) (*types.TQuest, error)
	GetQuests(ctx context.Context, userID uint) (*[]types.TQuest, error)
	EditQuest(ctx context.Context, quest *types.TQuest, userID uint) (*types.TQuest, error)
	GetAwaitQuests(ctx context.Context, userID uint) (*[]types.TQuestAwait, error)

	ManageQuestConfirmation(ctx context.Context, statusID, userID uint, confirm bool) error

	ManageCreateSelfMaster(ctx context.Context, userID uint) (*models.User, error)

	AddMaster(ctx context.Context, masterCode string, playerID uint) error
	GetQuestsPlayer(ctx context.Context, playerID uint) (*[]types.TPlayerQuest, error)
	GetQuestPlayer(ctx context.Context, questID uint, playerID uint) (*types.TPlayerQuest, error)
	SendQuestPlayer(ctx context.Context, questID, playerID uint) (*types.TPlayerQuest, error)
	GetPlayerMasters(ctx context.Context, playerID uint) (*[]models.UserMaster, error)
	GetWallets(ctx context.Context, playerID uint) (*[]types.TPlayerWaller, error)
}

type userInterface interface {
	Error500Page(wr http.ResponseWriter) error
	Error403Page(wr http.ResponseWriter) error
	Error404Page(wr http.ResponseWriter) error
	RegistrationPage(wr http.ResponseWriter) error
	AuthorizationPage(wr http.ResponseWriter) error
	MainPage(wr http.ResponseWriter, user *types.TMainPage) error
	ProfilePage(wr http.ResponseWriter, page *types.TProfilePage) error
	QuestGiverPage(wr http.ResponseWriter, page *types.TQuestGiverPage) error
	QuestEdit(wr http.ResponseWriter, page *types.TQuestEdit) error
	QuestNew(wr http.ResponseWriter, page *types.TQuestNew) error
	QuestAwait(wr http.ResponseWriter, page *types.TQuestAwaitPage) error

	PlayerProfile(wr http.ResponseWriter, page *types.TPlayerProfilePage) error
	PlayerQuests(wr http.ResponseWriter, page *types.TPlayerQuestsPage) error
	PlayerQuest(wr http.ResponseWriter, page *types.TPlayerQuestPage) error
	PlayerSettings(wr http.ResponseWriter, page *types.TPlayerSettingsPage) error
}

type sess interface {
	Middleware() gin.HandlerFunc
	SaveUser(c *gin.Context, user *models.User) error
	GetUser(c *gin.Context) (*types.TUser, error)
}

// Server - REST API сервер.
type Server struct {
	log       *zap.Logger
	disc      discipline
	sess      sess
	ui        userInterface
	baseURL   string
	secretKey []byte
	s         http.Server
	tlsEnable bool
}

// Option - опции сервера.
type Option func(s *Server)

// BaseURL - Настройка сервера, задает полный путь для сокращенной ссылки.
func BaseURL(url string) func(*Server) {
	return func(s *Server) {
		s.baseURL = url
	}
}

// Addr - Насткройка сервера, задает адрес сервера.
func SetAddress(addr string) func(s *Server) {
	return func(s *Server) {
		s.s.Addr = addr
	}
}

// Logger - Устанавливает логер.
func SetLogger(log *zap.Logger) func(s *Server) {
	return func(s *Server) {
		s.log = log
	}
}

// SecretKey - задает секретный ключ.
func SetSecretKey(secret []byte) Option {
	return func(s *Server) {
		s.secretKey = secret
	}
}

// HTTPSEnable - включает https.
func HTTPSEnable(enable bool) Option {
	return func(s *Server) {
		s.tlsEnable = enable
	}
}

// New создает Server.
func New(disc discipline, ui userInterface, sess sess, options ...Option) *Server {
	srv := &Server{
		disc:      disc,
		ui:        ui,
		sess:      sess,
		log:       zap.NewNop(),
		secretKey: []byte("rest_secret_key"),
	}
	srv.s.Addr = "localhost:8080"

	for _, opt := range options {
		opt(srv)
	}

	return srv
}

// SetupRouter - создает маршруты.
func (s *Server) SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(
		s.Logger(),
		s.sess.Middleware(),
	)
	r.Use(s.middlewareErrorPage())

	r.GET("/registration", s.handlerRegistrationPage)
	r.GET("/authorization", s.handlerAuthorization)

	auth := r.Group("/")
	{
		auth.Use(s.middlewareAuthentication())
		auth.GET("/", s.handlerPlayer)
		auth.GET("/logout", s.handleUserLogout)
		auth.GET("/profile", s.handlerUserProfile)

		manage := auth.Group("/manager")
		manage.Use(s.middlewareManagerRole())
		{
			manage.GET("/", s.handlerQuestGiverPage)
			manage.GET("/quests/:id", s.handlerQuestEdit)
			manage.POST("/quests/:id", s.handlerQuestEdit)
			manage.GET("/quests/new", s.handlerQuestNew)
			manage.POST("/quests/new", s.handlerQuestNew)
			manage.GET("/quests/await", s.handlerQuestAwait)
		}
		player := auth.Group("/player")
		{
			player.GET("/", s.handlerPlayer)
			player.GET("/quests", s.handlerPlayerQuests)
			player.GET("/quests/:id", s.handlerPlayerQuest)
			player.POST("/quests/:id", s.handlerPlayerQuest)
			player.GET("/settings", s.handlerPlayerSettings)
		}
	}

	apiUser := r.Group("/api/v0/user")
	apiUser.Use(s.middlewareAuthenticationAPI())
	{
		apiUser.GET("/info", s.handlerAPIUserInfo)
	}

	api := r.Group("/api/v0")
	{
		apiAuth := api.Group("/auth")
		{
			apiAuth.POST("/registration", s.handlerApiRegistration)
			apiAuth.POST("/authorization", s.handlerAPIAuthorization)
		}
		apiManage := api.Group("/manage")
		apiManage.Use(s.middlewareManagerRole())
		{
			apiManage.POST("/quests/status/confirmation", s.handlerAPIManageQuestConfirmation)
		}
		apiUser := api.Group("/user")
		{
			apiUser.POST("/settings/self/master", s.handlerAPIManageCreateMaster)
			apiUser.POST("/settings/player/master", s.handlerAPIPlayerAddMaster)
		}
	}

	r.Static("/static", "static")

	return r
}

// Run - запускает сервер.
func (s *Server) Run() error {
	s.s.Handler = s.SetupRouter().Handler()
	switch s.tlsEnable {
	case false:
		if err := s.s.ListenAndServe(); err != nil {
			return fmt.Errorf("server has failed: %w", err)
		}
	case true:
		if err := s.s.ListenAndServeTLS("./cert/server.crt", "./cert/server.key"); err != nil {
			return fmt.Errorf("server has failed: %w", err)
		}
	}
	return nil
}

// Stop - остановка сервера.
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownDelay)
	defer cancel()
	err := s.s.Shutdown(ctx)
	if err != nil {
		s.log.Error("failed shutdown server", zap.Error(err))
	}
	s.log.Info("Server exiting")
}

func unauthorize(c *gin.Context) {
	userCookie := &http.Cookie{
		Name:  cookieName,
		Value: "",
		Path:  "/",
	}
	c.Request.AddCookie(userCookie)
	http.SetCookie(c.Writer, userCookie)
}

func (s *Server) authorization(c *gin.Context, login, password string) error {
	var err error
	var user *models.User
	ctx := c.Request.Context()
	if user, err = s.disc.Login(ctx, login, password); err != nil {
		return fmt.Errorf("failed authorization: %w", err)
	}

	jwtRest := jwt.New([]byte(s.secretKey))
	signedCookie, err := jwtRest.Create(cookieKey, strconv.Itoa(int(user.ID)))
	if err != nil {
		return fmt.Errorf("can't create cookie data: %w", err)
	}

	userCookie := &http.Cookie{
		Name:  cookieName,
		Value: signedCookie,
		Path:  "/",
	}
	c.Request.AddCookie(userCookie)
	http.SetCookie(c.Writer, userCookie)

	err = s.sess.SaveUser(c, user)
	if err != nil {
		s.log.Error("failed save session", zap.Error(err))
		return err
	}

	return nil
}

func (s *Server) readBody(c *gin.Context) ([]byte, int) {
	bBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.log.Error("failed read body", zap.Error(err))
		return []byte{}, http.StatusInternalServerError
	}
	defer func() {
		if err := c.Request.Body.Close(); err != nil {
			s.log.Error(msgErrorCloseBody, zap.Error(err))
		}
	}()
	return bBody, 0
}
