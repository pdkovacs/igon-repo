package main

import (
	"fmt"
	"os"

	"github.com/pdkovacs/igo-repo/internal/api"
	"github.com/pdkovacs/igo-repo/internal/auxiliaries"
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
			fmt.Print(auxiliaries.GetBuildInfoString())
			serverWanted = false
		}
	}

	if serverWanted {
		conf := auxiliaries.ParseCommandLineArgs(os.Args)
		server := api.Server{}
		server.SetupAndStart(conf, func(port int) {
		})
	}
}
