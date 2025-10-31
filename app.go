package main

import (
	api_db "github.com/lapuglisi/dvapi/database"
	api_http "github.com/lapuglisi/dvapi/http"
)

type ApiApplication struct {
	server api_http.ApiHttpServer
	db     *api_db.DuckDatabase
}

func (app *ApiApplication) Setup(host string, port int) (err error) {
	app.db = api_db.NewDatabase()
	err = app.db.Setup()
	if err != nil {
		return err
	}

	app.server.Setup(host, port, app.db)

	return err
}

func (app *ApiApplication) Run() (err error) {
	// defer app.shutdown()

	err = app.server.Run()

	app.shutdown()

	return err
}

func (app *ApiApplication) shutdown() (err error) {
	return app.db.Release()
}
