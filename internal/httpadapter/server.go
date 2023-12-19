package httpadapter

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/app/services"
	"iconrepo/internal/config"
	"iconrepo/internal/logging"
	"iconrepo/web"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

type server struct {
	listener      net.Listener
	configuration config.Options
	logger        zerolog.Logger
	api           services.IconService
}

func CreateServer(configuration config.Options, api services.IconService) server {
	return server{
		configuration: configuration,
		api:           api,
		logger:        logging.Get().With().Str(logging.UnitLogger, "http-server").Logger(),
	}
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

	logger.Info().Str("port", port).Msg("started to listen")

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
		logger.Info().Str("database", options.SessionDbName).Msg("connecting to session store")
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
		logger.Info().Msg("Using in-memory session store")
		store = memstore.NewStore([]byte("secret"))
	}

	return store, nil
}

func (s *server) initEndpoints(options config.Options) *gin.Engine {
	logger := logging.CreateMethodLogger(s.logger, "server:initEndpoints")
	authorizationService := services.NewAuthorizationService(options)
	userService := services.NewUserService(&authorizationService)

	rootEngine := gin.Default()
	rootEngine.Use(RequestLogger)

	if config.UseCORS(options) {
		rootEngine.Use(CORSMiddleware(options.AllowedClientURLsRegex, logging.CreateMethodLogger(logger, "CORS")))
	}

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

	logger.Debug().Str("authenticationType", string(options.AuthenticationType)).Msg("Creating login end-point...")
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

		logger.Debug().Str("authn-type", string(options.AuthenticationType)).Msg("Setting up authorized group")
		authorizedGroup.Use(authenticationCheck(options, &userService, s.logger.With().Logger()))

		rootEngine.GET("/config", func(c *gin.Context) {
			c.JSON(200, clientConfig{IdPLogoutURL: options.OIDCLogoutURL})
		})
		logger.Debug().Msg("Setting up logout handler")
		authorizedGroup.POST("/logout", logout(options))

		authorizedGroup.GET("/subscribe", subscriptionHandler(mustGetUserInfo, notifService, options.LoadBalancerAddress))

		authorizedGroup.GET("/user", userInfoHandler(options.AuthenticationType, userService))

		if options.EnableBackdoors {
			authorizedGroup.PUT("/backdoor/authentication", HandlePutIntoBackdoorRequest())
			authorizedGroup.GET("/backdoor/authentication", HandleGetIntoBackdoorRequest())
		}

		authorizedGroup.GET("/icon", describeAllIcons(s.api.DescribeAllIcons))
		authorizedGroup.GET("/icon/:name", describeIcon(s.api.DescribeIcon))
		authorizedGroup.POST("/icon", createIcon(mustGetUserInfo, s.api.CreateIcon, notifService.Publish))
		authorizedGroup.DELETE("/icon/:name", deleteIcon(mustGetUserInfo, s.api.DeleteIcon, notifService.Publish))

		authorizedGroup.POST("/icon/:name", addIconfile(mustGetUserInfo, s.api.AddIconfile, notifService.Publish))
		authorizedGroup.GET("/icon/:name/format/:format/size/:size", getIconfile(s.api.GetIconfile))
		authorizedGroup.DELETE("/icon/:name/format/:format/size/:size", deleteIconfile(mustGetUserInfo, s.api.DeleteIconfile, notifService.Publish))

		authorizedGroup.GET("/tag", getTags(s.api.GetTags))
		authorizedGroup.POST("/icon/:name/tag", addTag(mustGetUserInfo, s.api.AddTag))
		authorizedGroup.DELETE("/icon/:name/tag/:tag", removeTag(mustGetUserInfo, s.api.RemoveTag))
	}

	return rootEngine
}

// Stop kills the listener
func (s *server) Stop() {
	logger := logging.CreateMethodLogger(s.logger, "ListenerKiller")
	error := s.listener.Close()
	if error != nil {
		logger.Error().Err(error).Interface("listener", s.listener).Msg("Error while closing listener")
	} else {
		logger.Info().Interface("listener", s.listener).Msg("Listener closed successfully")
	}
}

func RequestLogger(g *gin.Context) {
	start := time.Now()

	l := logging.Get().With().Str("req_xid", xid.New().String()).Logger()

	r := g.Request
	g.Request = r.WithContext(l.WithContext(r.Context()))

	lrw := newLoggingResponseWriter(g.Writer)

	defer func() {
		panicVal := recover()
		if panicVal != nil {
			lrw.statusCode = http.StatusInternalServerError // ensure that the status code is updated
			panic(panicVal)                                 // continue panicking
		}
		l.
			Info().
			Str("method", g.Request.Method).
			Str("url", g.Request.URL.RequestURI()).
			Str("user_agent", g.Request.UserAgent()).
			Int("status_code", lrw.statusCode).
			Dur("elapsed_ms", time.Since(start)).
			Msg("incoming request")
	}()

	g.Next()
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// TODO:
// Use "github.com/gin-contrib/cors" with strict, parameterized rules
func CORSMiddleware(clientURLs string, logger zerolog.Logger) gin.HandlerFunc {
	clientURLsRegexp := regexp.MustCompile(clientURLs)

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("origin")

		if origin == "" { // Not from browser or not COR
			c.Next()
			return
		}

		matchingOrigin := clientURLsRegexp.FindString(origin)

		if matchingOrigin == "" {
			logger.Debug().Str("request-method", c.Request.Method).Str("origin", origin).Interface("client-urls", clientURLs).Msg("No matching origin")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		logger.Debug().Str("request-method", c.Request.Method).Str("origin", origin).Str("matching-origin", matchingOrigin).Msg("request origin matched")

		c.Writer.Header().Set("Access-Control-Allow-Origin", matchingOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
