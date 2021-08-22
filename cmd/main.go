package main

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pdkovacs/igo-repo/app"
	"github.com/pdkovacs/igo-repo/config"
	httpadapter "github.com/pdkovacs/igo-repo/http"
	"github.com/pdkovacs/igo-repo/repositories"
	log "github.com/sirupsen/logrus"
)

func setLogLevel(levelArg string) {
	var level log.Level
	if levelArg == "info" {
		level = log.InfoLevel
	} else if levelArg == "debug" {
		level = log.DebugLevel
	}
	fmt.Printf("Log level: %v\n", level)
	log.SetLevel(level)
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

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

		setLogLevel(conf.LogLevel)

		db, dbErr := repositories.InitDBRepo(conf)
		if dbErr != nil {
			panic(dbErr)
		}

		git := &repositories.GitRepository{Location: conf.IconDataLocationGit}
		gitErr := git.InitMaybe()
		if gitErr != nil {
			panic(gitErr)
		}

		combinedRepo := repositories.RepoCombo{DB: db, Git: git}

		app := app.App{Repository: &combinedRepo}

		server := httpadapter.Server{API: httpadapter.API{
			IconService: &app.GetAPI().IconService,
		}}

		server.SetupAndStart(conf, func(port int) {
		})

	}
}
