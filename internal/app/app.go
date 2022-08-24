package app

import (
	"igo-repo/internal/config"
	httpadapter "igo-repo/internal/http"
	"igo-repo/internal/logging"
	"igo-repo/internal/repositories"
)

func Start(conf config.Options, ready func(port int, server httpadapter.Stoppable)) error {

	rootLogger := logging.CreateRootLogger(conf.LogLevel)

	connection, dbErr := repositories.NewDBConnection(conf, logging.CreateUnitLogger(rootLogger, "db-connection"))
	if dbErr != nil {
		return dbErr
	}

	_, schemaErr := repositories.OpenDBSchema(conf, connection, logging.CreateUnitLogger(rootLogger, "db-schema"))
	if schemaErr != nil {
		return schemaErr
	}

	db := repositories.NewDBRepository(connection, logging.CreateUnitLogger(rootLogger, "db-repository"))

	git := repositories.NewGitRepository(conf.IconDataLocationGit, logging.CreateUnitLogger(rootLogger, "git-repository"))
	gitErr := git.InitMaybe()
	if gitErr != nil {
		return gitErr
	}

	combinedRepo := repositories.RepoCombo{DB: db, Git: git}

	appRef := &AppCore{Repository: &combinedRepo}

	server := httpadapter.CreateServer(
		conf,
		httpadapter.CreateAPI(appRef.GetAPI(logging.CreateUnitLogger(rootLogger, "api")).IconService),
		logging.CreateUnitLogger(rootLogger, "server"),
	)

	server.SetupAndStart(conf, func(port int, app httpadapter.Stoppable) {
		ready(port, app)
	})

	return nil
}
