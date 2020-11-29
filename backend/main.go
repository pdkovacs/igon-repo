package main

import (
	"github.com/pdkovacs/igo-repo/pkg/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

	server.SetupAndStartServer(func(port int) {
	})
}
