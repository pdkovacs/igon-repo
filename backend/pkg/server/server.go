package server

import (
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var listener net.Listener

// StartServer starts the server
var StartServer = func(r http.Handler, ready func(port int)) {
	logger := log.WithField("prefix", "StartServer")
	logger.Info("Starting server on ephemeral....")
	var err error
	listener, err = net.Listen("tcp", ":0")
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
			log.Panic(err)
		}
		ready(portAsInt)
	}

	http.Serve(listener, r)
}

// SetupAndStartServer sets up and starts server.
var SetupAndStartServer = func(ready func(port int)) {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	StartServer(r, ready)
}

// ListenerKiller kills the listener
var ListenerKiller = func() {
	logger := log.WithField("prefix", "ListenerKiller")
	logger.Debug("listener: ", listener)
	error := listener.Close()
	if error != nil {
		logger.Errorf("Error while closing listener: %v", error)
	} else {
		logger.Info("Listener closed successfully")
	}
}
