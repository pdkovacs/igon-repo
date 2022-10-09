package repositories

import (
	"igo-repo/internal/repositories/gitrepo"
	"igo-repo/internal/repositories/icondb"
)

type Repositories struct {
	DB  *icondb.Repository
	Git *gitrepo.Local
}

func NewRepositories(DB *icondb.Repository, Git *gitrepo.Local) Repositories {
	return Repositories{DB, Git}
}
