package main

import (
	"fmt"
	"os"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

	var serverWanted bool = true

	for _, value := range os.Args {
		if value == "-v" || value == "--version" {
			fmt.Printf(auxiliaries.GetBuildInfoString())
			serverWanted = false
		}
	}

	if serverWanted {
		conf, configurationReadError := auxiliaries.ReadConfiguration("", os.Args)
		if configurationReadError != nil {
			log.Fatalf("Failed to read configuratioin: %v", configurationReadError)
		}

		server.SetupAndStart(conf.ServerPort, conf, func(port int) {
		})
	}
}
