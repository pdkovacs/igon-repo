package http

import (
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"
	"igo-repo/web"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
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
	Configuration config.Options
	API           api
	logger        zerolog.Logger
}

func CreateServer(configuration config.Options, api api, logger zerolog.Logger) server {
	return server{
		Configuration: configuration,
		API:           api,
		logger:        logger,
	}
}

// Start starts the service
func (s *server) Start(portRequested int, r http.Handler, ready func(port int)) {
	logger := s.logger.With().Str("prefix", "StartServer").Logger()
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
		ready(portAsInt)
	}

	http.Serve(s.listener, r)
}

// SetupAndStart sets up and starts server.
func (s *server) SetupAndStart(options config.Options, ready func(port int)) {
	r := s.initEndpoints(options)
	s.Start(options.ServerPort, r, ready)
}

func (s *server) initEndpoints(options config.Options) *gin.Engine {
	logger := s.logger.With().Str("prefix", "server:initEndpoints").Logger()
	authorizationService := services.NewAuthorizationService(options)
	userService := services.NewUserService(&authorizationService)

	gob.Register(SessionData{})

	rootEngine := gin.Default()

	store := memstore.NewStore([]byte("secret"))
	store.Options(sessions.Options{MaxAge: 60 * 60 * 24})
	rootEngine.Use(sessions.Sessions("mysession", store))

	rootEngine.NoRoute(Authentication(options, &userService), gin.WrapH(web.AssetHandler("/", "dist")))

	logger.Debug().Msgf("Creating login end-point with authentication type: %v...", options.AuthenticationType)
	rootEngine.GET("/login", Authentication(options, &userService))

	rootEngine.GET("/app-info", func(c *gin.Context) {
		c.JSON(200, config.GetBuildInfo())
	})

	logger.Debug().Msg("Creating authorized group....")

	authorizedGroup := rootEngine.Group("/")
	{
		logger.Debug().Msgf("Setting up authorized group with authentication type: %v...", options.AuthenticationType)
		authorizedGroup.Use(AuthenticationCheck(options, &userService))

		authorizedGroup.GET("/user", UserInfoHandler(userService))

		if options.EnableBackdoors {
			authorizedGroup.PUT("/backdoor/authentication", HandlePutIntoBackdoorRequest)
			authorizedGroup.GET("/backdoor/authentication", HandleGetIntoBackdoorRequest)
		}

		authorizedGroup.GET("/icon", describeAllIconsHanler(s.API.iconService.DescribeAllIcons, logging.CreateMethodLogger(s.logger, "describeAllIconsHanler")))
		authorizedGroup.GET("/icon/:name", describeIconHandler(s.API.iconService.DescribeIcon, logging.CreateMethodLogger(s.logger, "describeIconHandler")))
		authorizedGroup.POST("/icon", createIconHandler(s.API.iconService.CreateIcon, logging.CreateMethodLogger(s.logger, "createIconHandler")))
		authorizedGroup.DELETE("/icon/:name", deleteIconHandler(s.API.iconService.DeleteIcon, logging.CreateMethodLogger(s.logger, "deleteIconHandler")))

		authorizedGroup.POST("/icon/:name", addIconfileHandler(s.API.iconService.AddIconfile, logging.CreateMethodLogger(s.logger, "addIconfileHandler")))
		authorizedGroup.GET("/icon/:name/format/:format/size/:size", getIconfileHandler(s.API.iconService.GetIconfile, logging.CreateMethodLogger(s.logger, "getIconfileHandler")))
		authorizedGroup.DELETE("/icon/:name/format/:format/size/:size", deleteIconfileHandler(s.API.iconService.DeleteIconfile, logging.CreateMethodLogger(s.logger, "deleteIconfileHandler")))

		authorizedGroup.GET("/tag", getTagsHandler(s.API.iconService.GetTags, logging.CreateMethodLogger(s.logger, "getTagsHandler")))
		authorizedGroup.POST("/icon/:name/tag", addTagHandler(s.API.iconService.AddTag, logging.CreateMethodLogger(s.logger, "addTagHandler")))
		authorizedGroup.DELETE("/icon/:name/tag/:tag", removeTagHandler(s.API.iconService.RemoveTag, logging.CreateMethodLogger(s.logger, "removeTagHandler")))
	}

	return rootEngine
}

// Close kills the listener
func (s *server) Close() {
	logger := s.logger.With().Str("prefix", "ListenerKiller").Logger()
	logger.Info().Msgf("listener: %v", s.listener)
	error := s.listener.Close()
	if error != nil {
		logger.Error().Msgf("Error while closing listener: %v", error)
	} else {
		logger.Info().Msg("Listener closed successfully")
	}
}
