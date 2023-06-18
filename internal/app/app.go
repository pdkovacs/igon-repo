package app

import (
	"igo-repo/internal/config"
	"igo-repo/internal/httpadapter"
	"igo-repo/internal/repositories"
	"igo-repo/internal/repositories/gitrepo"
	"igo-repo/internal/repositories/icondb"
)

func Start(conf config.Options, ready func(port int, stop func())) error {

	connection, dbErr := icondb.NewDBConnection(conf)
	if dbErr != nil {
		return dbErr
	}

	dbSchemaAlreadyThere, schemaErr := icondb.OpenSchema(conf, connection)
	if schemaErr != nil {
		return schemaErr
	}

	db := icondb.NewDBRepository(connection)

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

	combinedRepo := repositories.RepoCombo{DB: db, Git: git}

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
