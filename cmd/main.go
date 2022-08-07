package main

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
	"time"

	"igo-repo/internal/app"
	"igo-repo/internal/config"
	httpadapter "igo-repo/internal/http"
	"igo-repo/internal/logging"
	"igo-repo/internal/repositories"

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

		rootLogger := logging.CreateRootLogger(conf.LogLevel)

		connection, dbErr := repositories.NewDBConnection(conf, logging.CreateUnitLogger(rootLogger, "db-connection"))
		if dbErr != nil {
			panic(dbErr)
		}
		db := repositories.NewDBRepository(connection, logging.CreateUnitLogger(rootLogger, "db-repository"))

		git := repositories.NewGitRepository(conf.IconDataLocationGit, logging.CreateUnitLogger(rootLogger, "git-repository"))
		gitErr := git.InitMaybe()
		if gitErr != nil {
			panic(gitErr)
		}

		combinedRepo := repositories.RepoCombo{DB: db, Git: git}

		app := app.App{Repository: &combinedRepo}

		server := httpadapter.CreateServer(
			conf,
			httpadapter.CreateAPI(app.GetAPI(logging.CreateUnitLogger(rootLogger, "api")).IconService),
			logging.CreateUnitLogger(rootLogger, "server"),
		)

		server.SetupAndStart(conf, func(port int) {
		})
	}
}
