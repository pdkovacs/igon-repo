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
	DBRepository  *repositories.DatabaseRepository
	GitRepository *repositories.GitRepository
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
	s.DBRepository, err = repositories.InitDBRepo(options)
	if err != nil {
		panic(err)
	}

	s.GitRepository = &repositories.GitRepository{Location: options.IconDataLocationGit}
	err = s.GitRepository.InitMaybe()
	if err != nil {
		panic(err)
	}
	s.Configuration = options
	r := initEndpoints(options)
	s.Start(options.ServerPort, r, ready)
}

func initEndpoints(options auxiliaries.Options) *gin.Engine {
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

	r.GET("/info", func(c *gin.Context) {
		c.JSON(200, auxiliaries.GetBuildInfo())
	})

	r.GET("/user", UserInfoHandler(userService))

	if options.EnableBackdoors {
		r.PUT("/backdoor/authentication", HandlePutIntoBackdoorRequest)
		r.GET("/backdoor/authentication", HandleGetIntoBackdoorRequest)
	}

	r.POST("/icon", createIconHandler)

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
