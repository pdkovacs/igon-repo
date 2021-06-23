package web

import (
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/pdkovacs/igo-repo/backend/pkg/services"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	listener      net.Listener
	Configuration auxiliaries.Options
	Repositories  *repositories.Repositories
}

// Start starts the server
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
func (s *Server) SetupAndStart(options auxiliaries.Options, ready func(port int)) {
	var err error
	s.Repositories = &repositories.Repositories{}

	s.Repositories.DB, err = repositories.InitDBRepo(options)
	if err != nil {
		panic(err)
	}

	s.Repositories.Git = &repositories.GitRepository{Location: options.IconDataLocationGit}
	err = s.Repositories.Git.InitMaybe()
	if err != nil {
		panic(err)
	}
	s.Configuration = options
	r := s.initEndpoints(options)
	s.Start(options.ServerPort, r, ready)
}

func (s *Server) initEndpoints(options auxiliaries.Options) *gin.Engine {
	logger := log.WithField("prefix", "server:initEndpoints")
	authorizationService := services.NewAuthorizationService(options)
	userService := services.NewUserService(&authorizationService)

	gob.Register(SessionData{})
	r := gin.Default()
	store := memstore.NewStore([]byte("secret"))
	store.Options(sessions.Options{MaxAge: 60 * 60 * 24})
	r.Use(sessions.Sessions("mysession", store))
	if options.PasswordCredentials != nil && len(options.PasswordCredentials) > 0 {
		r.Use(HandlerProvider(BasicConfig{PasswordCredentialsList: options.PasswordCredentials}, &userService))
	}

	r.POST("/login", func(c *gin.Context) {
		session := MustGetUserSession(c)
		logger.Infof("%v logged in", session.UserInfo)
		c.JSON(200, session.UserInfo)
	})

	r.GET("/info", func(c *gin.Context) {
		c.JSON(200, auxiliaries.GetBuildInfo())
	})

	r.GET("/user", UserInfoHandler(userService))

	if options.EnableBackdoors {
		r.PUT("/backdoor/authentication", HandlePutIntoBackdoorRequest)
		r.GET("/backdoor/authentication", HandleGetIntoBackdoorRequest)
	}

	iconService := services.IconService{Repositories: s.Repositories}

	r.GET("/icon", describeAllIconsHanler(&iconService))
	r.POST("/icon", createIconHandler(&iconService))

	r.POST("/icon/:name", addIconfileHandler(&iconService))
	r.GET("/icon/:name/format/:format/size/:size", getIconfileHandler(&iconService))

	return r
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
