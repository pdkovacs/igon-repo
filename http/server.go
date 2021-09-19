package http

import (
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/app/domain"
	"github.com/pdkovacs/igo-repo/app/security/authr"
	"github.com/pdkovacs/igo-repo/app/services"
	"github.com/pdkovacs/igo-repo/config"
	"github.com/pdkovacs/igo-repo/web"
	log "github.com/sirupsen/logrus"
)

type IconService interface {
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

type API struct {
	IconService IconService
}

type Server struct {
	listener      net.Listener
	Configuration config.Options
	API           API
}

// Start starts the service
func (s *Server) Start(portRequested int, r http.Handler, ready func(port int)) {
	logger := log.WithField("prefix", "StartServer")
	logger.Info("Starting server on ephemeral....")
	var err error

	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", portRequested))
	if err != nil {
		logger.Fatalf("Error while starting to listen at an ephemeral port: %v", err)
	}

	_, port, err := net.SplitHostPort(s.listener.Addr().String())
	if err != nil {
		logger.Fatalf("Error while parsing the server address: %v", err)
	}

	logger.Info("Listening on port: ", port)

	if ready != nil {
		portAsInt, err := strconv.Atoi(port)
		if err != nil {
			logger.Panic(err)
		}
		ready(portAsInt)
	}

	http.Serve(s.listener, r)
}

// SetupAndStart sets up and starts server.
func (s *Server) SetupAndStart(options config.Options, ready func(port int)) {
	r := s.initEndpoints(options)
	s.Start(options.ServerPort, r, ready)
}

func (s *Server) initEndpoints(options config.Options) *gin.Engine {
	logger := log.WithField("prefix", "server:initEndpoints")
	authorizationService := services.NewAuthorizationService(options)
	userService := services.NewUserService(&authorizationService)

	gob.Register(SessionData{})

	rootEngine := gin.Default()

	store := memstore.NewStore([]byte("secret"))
	store.Options(sessions.Options{MaxAge: 60 * 60 * 24})
	rootEngine.Use(sessions.Sessions("mysession", store))

	rootEngine.NoRoute(Authentication(options, &userService), gin.WrapH(web.AssetHandler("/", "dist")))

	rootEngine.GET("/login", Authentication(options, &userService))

	rootEngine.GET("/app-info", func(c *gin.Context) {
		c.JSON(200, config.GetBuildInfo())
	})

	logger.Debug("Creating authorized group....")

	authorizedGroup := rootEngine.Group("/")
	{
		logger.Debug("Setting up authorized group...")
		authorizedGroup.Use(AuthenticationCheck(options, &userService))

		authorizedGroup.GET("/user", UserInfoHandler(userService))

		if options.EnableBackdoors {
			authorizedGroup.PUT("/backdoor/authentication", HandlePutIntoBackdoorRequest)
			authorizedGroup.GET("/backdoor/authentication", HandleGetIntoBackdoorRequest)
		}

		authorizedGroup.GET("/icon", describeAllIconsHanler(s.API))
		authorizedGroup.GET("/icon/:name", describeIconHandler(s.API))
		authorizedGroup.POST("/icon", createIconHandler(s.API))
		authorizedGroup.DELETE("/icon/:name", deleteIconHandler(s.API))

		authorizedGroup.POST("/icon/:name", addIconfileHandler(s.API))
		authorizedGroup.GET("/icon/:name/format/:format/size/:size", getIconfileHandler(s.API))
		authorizedGroup.DELETE("/icon/:name/format/:format/size/:size", deleteIconfileHandler(s.API))

		authorizedGroup.GET("/tag", getTagsHandler(s.API))
		authorizedGroup.POST("/icon/:name/tag", addTagHandler(s.API))
		authorizedGroup.DELETE("/icon/:name/tag/:tag", removeTagHandler(s.API))
	}

	return rootEngine
}

// KillListener kills the listener
func (s *Server) KillListener() {
	logger := log.WithField("prefix", "ListenerKiller")
	logger.Infof("listener: %v", s.listener)
	error := s.listener.Close()
	if error != nil {
		logger.Errorf("Error while closing listener: %v", error)
	} else {
		logger.Info("Listener closed successfully")
	}
}
