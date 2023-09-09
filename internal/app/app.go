package app

import (
	"iconrepo/internal/app/services"
	"iconrepo/internal/config"
	"iconrepo/internal/httpadapter"
	"iconrepo/internal/logging"
	"iconrepo/internal/repositories"
	"iconrepo/internal/repositories/blobstore/git"
	"iconrepo/internal/repositories/indexing/dynamodb"
	"iconrepo/internal/repositories/indexing/pgdb"
)

func Start(conf config.Options, ready func(port int, stop func())) error {

	var dbSchemaAlreadyThere bool
	var db repositories.IndexRepository
	if conf.DynamodbURL == "" {
		connection, dbErr := pgdb.NewDBConnection(conf)
		if dbErr != nil {
			return dbErr
		}

		var schemaErr error
		dbSchemaAlreadyThere, schemaErr = pgdb.OpenSchema(conf, connection)
		if schemaErr != nil {
			return schemaErr
		}

		db = pgdb.NewPgRepository(connection)
	}

	if len(conf.DynamodbURL) > 0 {
		dyndb, createDyndbErr := dynamodb.NewDynamodbRepository(&conf)
		if createDyndbErr != nil {
			return createDyndbErr
		}
		db = dyndb
	}
	defer db.Close()

	var blobstore repositories.BlobstoreRepository
	if len(conf.LocalGitRepo) > 0 {
		localGit := git.NewLocalGitRepository(conf.LocalGitRepo, logging.CreateUnitLogger(logging.Get(), "local git repository"))
		blobstore = &localGit
	}
	if len(conf.GitlabNamespacePath) > 0 {
		gitlabClient, gitlabRepoErr := git.NewGitlabRepositoryClient(
			conf.GitlabNamespacePath,
			conf.GitlabProjectPath,
			conf.GitlabMainBranch,
			conf.GitlabAccessToken,
			logging.CreateUnitLogger(logging.Get(), "Gitlab client"),
		)
		if gitlabRepoErr != nil {
			return gitlabRepoErr
		}
		blobstore = &gitlabClient
	}

	if !dbSchemaAlreadyThere {
		gitErr := blobstore.CreateRepository()
		if gitErr != nil {
			return gitErr
		}
	}

	combinedRepo := repositories.RepoCombo{Index: db, Blobstore: blobstore}

	server := httpadapter.CreateServer(
		conf,
		*services.NewIconService(&combinedRepo),
	)

	server.SetupAndStart(conf, func(port int, stop func()) {
		ready(port, func() {
			stop()
			db.Close()
		})
	})

	return nil
}
