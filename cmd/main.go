package main

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"igo-repo/internal/app"
	"igo-repo/internal/config"
	httpadapter "igo-repo/internal/http"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var serverWanted bool = true

	for _, value := range os.Args {
		if value == "-v" || value == "--version" {
			fmt.Print(config.GetBuildInfoString())
			serverWanted = false
		}
	}

	if serverWanted {
		var confErr error

		conf, confErr := config.ReadConfiguration(config.GetConfigFilePath(), os.Args)
		if confErr != nil {
			panic(confErr)
		}

		var server httpadapter.Stoppable

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		go func() {
			s := <-sigc
			fmt.Fprintf(os.Stderr, "Caught %v, stopping...", s)
			server.Stop()
		}()

		app.Start(conf, func(port int, stoppable httpadapter.Stoppable) {
			server = stoppable
		})
	}
}
