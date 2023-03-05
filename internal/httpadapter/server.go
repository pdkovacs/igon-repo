package httpadapter

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"
	"igo-repo/web"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type iconService interface {
	DescribeAllIcons() ([]domain.IconDescriptor, error)
	DescribeIcon(iconName string) (domain.IconDescriptor, error)
	CreateIcon(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.Icon, error)
	GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	AddIconfile(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.IconfileDescriptor, error)
	DeleteIcon(iconName string, modifiedBy authr.UserInfo) error
	DeleteIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor, modifiedBy authr.UserInfo) error
	GetTags() ([]string, error)
	AddTag(iconName string, tag string, userInfo authr.UserInfo) error
	RemoveTag(iconName string, tag string, userInfo authr.UserInfo) error
}

type api struct {
	iconService iconService
}

func CreateAPI(iconService iconService) api {
	return api{
		iconService: iconService,
	}
}

type server struct {
	listener      net.Listener
	configuration config.Options
	logger        zerolog.Logger
	api           api
}

func CreateServer(configuration config.Options, api api, logger zerolog.Logger) server {
	return server{
		configuration: configuration,
		api:           api,
		logger:        logger,
	}
}

type Stoppable interface {
	Stop()
}

// start starts the service
func (s *server) start(portRequested int, r http.Handler, ready func(port int, stop func())) {
	logger := logging.CreateMethodLogger(s.logger, "StartServer")
	logger.Info().Msg("Starting server on ephemeral....")
	var err error

	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", portRequested))
	if err != nil {
		panic(fmt.Sprintf("Error while starting to listen at an ephemeral port: %v", err))
	}

	_, port, err := net.SplitHostPort(s.listener.Addr().String())
	if err != nil {
		panic(fmt.Sprintf("Error while parsing the server address: %v", err))
	}

	logger.Info().Msgf("Listening on port: %v", port)

	if ready != nil {
		portAsInt, err := strconv.Atoi(port)
		if err != nil {
			panic(err)
		}
		ready(portAsInt, s.Stop)
	}

	http.Serve(s.listener, r)
}

// SetupAndStart sets up and starts server.
func (s *server) SetupAndStart(options config.Options, ready func(port int, stop func())) {
	r := s.initEndpoints(options)
	s.start(options.ServerPort, r, ready)
}

func (s *server) createSessionStore(options config.Options) (sessions.Store, error) {
	var store sessions.Store
	logger := logging.CreateMethodLogger(s.logger, "create-session properties")

	if options.SessionDbName != "" {
		logger.Info().Msgf("Using DB %s for session store", options.SessionDbName)
		connProps := config.CreateDbProperties(s.configuration, s.configuration.DBSchemaName, logger)
		connStr := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			connProps.User,
			connProps.Password,
			connProps.Host,
			connProps.Port,
			options.SessionDbName,
		)
		sessionDb, openSessionDbErr := sql.Open("pgx", connStr)
		if openSessionDbErr != nil {
			return store, openSessionDbErr
		}
		sessionDb.Ping()
		var createDbSessionStoreErr error
		store, createDbSessionStoreErr = postgres.NewStore(sessionDb, []byte("secret"))
		if createDbSessionStoreErr != nil {
			return store, createDbSessionStoreErr
		}
	} else {
		logger.Info().Msgf("Using in-memory session store")
		store = memstore.NewStore([]byte("secret"))
	}

	return store, nil
}

func (s *server) initEndpoints(options config.Options) *gin.Engine {
	logger := logging.CreateMethodLogger(s.logger, "server:initEndpoints")
	authorizationService := services.NewAuthorizationService(options)
	userService := services.NewUserService(&authorizationService)

	rootEngine := gin.Default()

	rootEngine.Use(CORSMiddleware(options.ClientServerURL))

	if options.AuthenticationType != authn.SchemeOIDCProxy {
		gob.Register(SessionData{})
		store, createStoreErr := s.createSessionStore(options)
		if createStoreErr != nil {
			panic(createStoreErr)
		}
		store.Options(sessions.Options{MaxAge: options.SessionMaxAge})
		rootEngine.Use(sessions.Sessions("mysession", store))
	}

	rootEngine.NoRoute(authentication(options, &userService, s.logger.With().Logger()), gin.WrapH(web.AssetHandler("/", "dist", logger)))

	logger.Debug().Msgf("Creating login end-point with authentication type: %v...", options.AuthenticationType)
	rootEngine.GET("/login", authentication(options, &userService, s.logger.With().Logger()))

	rootEngine.GET("/app-info", func(c *gin.Context) {
		c.JSON(200, config.GetBuildInfo())
	})

	logger.Debug().Msg("Creating authorized group....")

	mustGetUserInfo := func(c *gin.Context) authr.UserInfo {
		userInfo, getUserInfoErr := getUserInfo(options.AuthenticationType)(c)
		if getUserInfoErr != nil {
			panic(fmt.Sprintf("failed to get user-info %s", c.Request.URL))
		}
		return userInfo
	}

	authorizedGroup := rootEngine.Group("/")
	{
		notifService := services.CreateNotificationService(logger)

		logger.Debug().Msgf("Setting up authorized group with authentication type: %v...", options.AuthenticationType)
		authorizedGroup.Use(authenticationCheck(options, &userService, s.logger.With().Logger()))

		rootEngine.GET("/config", func(c *gin.Context) {
			c.JSON(200, clientConfig{IdPLogoutURL: options.OIDCLogoutURL})
		})
		logger.Debug().Msg("Setting up logout handler")
		authorizedGroup.POST("/logout", logout(options, s.logger.With().Logger()))

		authorizedGroup.GET("/subscribe", subscriptionHandler(mustGetUserInfo, notifService, options.LoadBalancerAddress, logging.CreateMethodLogger(s.logger, "subscriptionHandler")))

		authorizedGroup.GET("/user", userInfoHandler(options.AuthenticationType, userService, logging.CreateMethodLogger(s.logger, "UserInfoHandler")))

		if options.EnableBackdoors {
			authorizedGroup.PUT("/backdoor/authentication", HandlePutIntoBackdoorRequest(logging.CreateMethodLogger(s.logger, "PUT /backdoor/authentication")))
			authorizedGroup.GET("/backdoor/authentication", HandleGetIntoBackdoorRequest(logging.CreateMethodLogger(s.logger, "GET /backdoor/authentication")))
		}

		authorizedGroup.GET("/icon", describeAllIconsHanler(s.api.iconService.DescribeAllIcons, logging.CreateMethodLogger(s.logger, "describeAllIconsHanler")))
		authorizedGroup.GET("/icon/:name", describeIconHandler(s.api.iconService.DescribeIcon, logging.CreateMethodLogger(s.logger, "describeIconHandler")))
		authorizedGroup.POST("/icon", createIconHandler(mustGetUserInfo, s.api.iconService.CreateIcon, notifService.Publish, logging.CreateMethodLogger(s.logger, "createIconHandler")))
		authorizedGroup.DELETE("/icon/:name", deleteIconHandler(mustGetUserInfo, s.api.iconService.DeleteIcon, notifService.Publish, logging.CreateMethodLogger(s.logger, "deleteIconHandler")))

		authorizedGroup.POST("/icon/:name", addIconfileHandler(mustGetUserInfo, s.api.iconService.AddIconfile, notifService.Publish, logging.CreateMethodLogger(s.logger, "addIconfileHandler")))
		authorizedGroup.GET("/icon/:name/format/:format/size/:size", getIconfileHandler(s.api.iconService.GetIconfile, logging.CreateMethodLogger(s.logger, "getIconfileHandler")))
		authorizedGroup.DELETE("/icon/:name/format/:format/size/:size", deleteIconfileHandler(mustGetUserInfo, s.api.iconService.DeleteIconfile, notifService.Publish, logging.CreateMethodLogger(s.logger, "deleteIconfileHandler")))

		authorizedGroup.GET("/tag", getTagsHandler(s.api.iconService.GetTags, logging.CreateMethodLogger(s.logger, "getTagsHandler")))
		authorizedGroup.POST("/icon/:name/tag", addTagHandler(mustGetUserInfo, s.api.iconService.AddTag, logging.CreateMethodLogger(s.logger, "addTagHandler")))
		authorizedGroup.DELETE("/icon/:name/tag/:tag", removeTagHandler(mustGetUserInfo, s.api.iconService.RemoveTag, logging.CreateMethodLogger(s.logger, "removeTagHandler")))
	}

	return rootEngine
}

// Stop kills the listener
func (s *server) Stop() {
	logger := logging.CreateMethodLogger(s.logger, "ListenerKiller")
	logger.Info().Msgf("listener: %v", s.listener)
	error := s.listener.Close()
	if error != nil {
		logger.Error().Msgf("Error while closing listener: %v", error)
	} else {
		logger.Info().Msg("Listener closed successfully")
	}

}

// TODO:
// Use "github.com/gin-contrib/cors" with strict, parameterized rules
func CORSMiddleware(clientServerURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", clientServerURL)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
