package app

import (
	"github.com/pdkovacs/igo-repo/app/services"
)

// Primary port
type API struct {
	IconService services.IconService
}

// Secondary port
type Repository = services.Repository

type App struct {
	Repository Repository
}

func (app *App) GetAPI() *API {
	return &API{
		IconService: services.IconService{Repository: app.Repository},
	}
}
