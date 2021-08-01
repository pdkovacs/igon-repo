package app

import (
	"github.com/pdkovacs/igo-repo/app/services"
)

type API struct {
	IconService services.IconService
}

type App struct {
	Repository services.Repository
}

func (app *App) GetAPI() *API {
	return &API{
		IconService: services.IconService{Repository: app.Repository},
	}
}
