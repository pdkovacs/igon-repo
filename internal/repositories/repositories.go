package repositories

type Repositories struct {
	DB  *DBRepository
	Git *GitRepository
}

func NewRepositories(DB *DBRepository, Git *GitRepository) Repositories {
	return Repositories{DB, Git}
}
