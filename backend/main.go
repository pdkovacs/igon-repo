package main

import (
	"fmt"
	"os"

	"github.com/pdkovacs/igo-repo/backend/pkg/build"
	"github.com/pdkovacs/igo-repo/backend/pkg/config"
	"github.com/pdkovacs/igo-repo/backend/pkg/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

	var wantedVersion bool = false

	for _, value := range os.Args {
		if value == "-v" || value == "--version" {
			fmt.Printf(build.GetInfoString())
			wantedVersion = true
		}
	}

	if !wantedVersion {
		conf, configurationReadError := config.ReadConfiguration("", os.Args)
		if configurationReadError != nil {
			log.Fatalf("Failed to read configuratioin: %v", configurationReadError)
		}

		server.SetupAndStart(conf.ServerPort, func(port int) {
		})
	}
}
