package main

import (
	dvapi_db "github.com/lapuglisi/dvapi/database"
	dvapi_http "github.com/lapuglisi/dvapi/http"
)

type ApiApplication struct {
	server dvapi_http.ApiHttpServer
	db     *dvapi_db.DuckDatabase
}

func (app *ApiApplication) Setup(host string, port int) (err error) {
	app.db = dvapi_db.NewDatabase()
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
