package app

import (
	"iconrepo/internal/config"
	"iconrepo/internal/httpadapter"
	"iconrepo/internal/repositories"
	"iconrepo/internal/repositories/gitrepo"
	"iconrepo/internal/repositories/indexing/dynamodb"
	"iconrepo/internal/repositories/indexing/pgdb"
)

func Start(conf config.Options, ready func(port int, stop func())) error {

	connection, dbErr := pgdb.NewDBConnection(conf)
	if dbErr != nil {
		return dbErr
	}

	dbSchemaAlreadyThere, schemaErr := pgdb.OpenSchema(conf, connection)
	if schemaErr != nil {
		return schemaErr
	}

	var db repositories.DBRepository
	if conf.DynamoDBURL == "" {
		db = pgdb.NewDBRepository(connection)
	} else {
		db = dynamodb.NewDynDBRepository()
	}

	var git repositories.GitRepository
	if len(conf.LocalGitRepo) > 0 {
		git = gitrepo.NewLocalGitRepository(conf.LocalGitRepo)
	}
	if len(conf.GitlabNamespacePath) > 0 {
		var gitlabRepoErr error
		git, gitlabRepoErr = gitrepo.NewGitlabRepositoryClient(
			conf.GitlabNamespacePath,
			conf.GitlabProjectPath,
			conf.GitlabMainBranch,
			conf.GitlabAccessToken,
		)
		if gitlabRepoErr != nil {
			return gitlabRepoErr
		}
	}

	if !dbSchemaAlreadyThere {
		gitErr := git.Create()
		if gitErr != nil {
			return gitErr
		}
	}

	combinedRepo := repositories.RepoCombo{Db: db, Git: git}

	appRef := &AppCore{Repository: &combinedRepo}

	server := httpadapter.CreateServer(
		conf,
		httpadapter.CreateAPI(appRef.GetAPI().IconService),
	)

	server.SetupAndStart(conf, func(port int, stop func()) {
		ready(port, func() {
			stop()
			connection.Pool.Close()
		})
	})

	return nil
}
