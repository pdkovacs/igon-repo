package server

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/backend/pkg/build"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/sirupsen/logrus"
)

var listener net.Listener

// Start starts the server
var Start = func(portRequested int, r http.Handler, ready func(port int)) {
	logger := logrus.WithField("prefix", "StartServer")
	logger.Info("Starting server on ephemeral....")
	var err error

	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", portRequested))
	if err != nil {
		logger.Fatalf("Error while starting to listen at an ephemeral port: %v", err)
	}

	_, port, err := net.SplitHostPort(listener.Addr().String())
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

	http.Serve(listener, r)
}

// SetupAndStart sets up and starts server.
var SetupAndStart = func(port int, ready func(port int)) {
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.Use(authn.HandlerProvider(authn.Basic))

	r.GET("/info", func(c *gin.Context) {
		c.JSON(200, build.GetInfo())
	})
	Start(port, r, ready)
}

// KillListener kills the listener
var KillListener = func() {
	logger := logrus.WithField("prefix", "ListenerKiller")
	logger.Debug("listener: ", listener)
	error := listener.Close()
	if error != nil {
		logger.Errorf("Error while closing listener: %v", error)
	} else {
		logger.Info("Listener closed successfully")
	}
}
