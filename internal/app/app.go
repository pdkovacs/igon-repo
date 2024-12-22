package app

import (
	"context"
	"iconrepo/internal/app/services"
	"iconrepo/internal/config"
	"iconrepo/internal/httpadapter"
	"iconrepo/internal/repositories"
	"iconrepo/internal/repositories/blobstore/git"
	"iconrepo/internal/repositories/indexing/dynamodb"
	"iconrepo/internal/repositories/indexing/pgdb"

	"github.com/rs/zerolog"
)

func Start(ctx context.Context, conf config.Options, ready func(port int, stop func())) error {
	logger := zerolog.Ctx(ctx)
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
	if len(conf.GitlabNamespacePath) == 0 && len(conf.LocalGitRepo) > 0 {
		localGit := git.NewLocalGitRepository(conf.LocalGitRepo)
		blobstore = &localGit
		logger.Info().Str("location", conf.LocalGitRepo).Msg("Connecting local git repo...")
	}
	if len(conf.GitlabNamespacePath) > 0 {
		gitlabClient, gitlabRepoErr := git.NewGitlabRepositoryClient(
			ctx,
			conf.GitlabNamespacePath,
			conf.GitlabProjectPath,
			conf.GitlabMainBranch,
			conf.GitlabAccessToken,
		)
		if gitlabRepoErr != nil {
			return gitlabRepoErr
		}
		blobstore = gitlabClient
		logger.Info().
			Str("gitlabNamespacePath,", conf.GitlabNamespacePath).
			Str("gitlabProjectPath,", conf.GitlabProjectPath).
			Str("gitlabMainBranch,", conf.GitlabMainBranch).
			Str("gitlabAccessToken,", conf.GitlabAccessToken).
			Msg("Connecting to Gitlab repo...")
	}

	if !dbSchemaAlreadyThere {
		gitErr := blobstore.CreateRepository(ctx)
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
