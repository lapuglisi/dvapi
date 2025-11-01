package main

import (
	"fmt"
	dvapi_db "github.com/lapuglisi/dvapi/database"
	dvapi_http "github.com/lapuglisi/dvapi/http"
	"os"
)

type ApiApplication struct {
	server dvapi_http.ApiHttpServer
	db     *dvapi_db.DuckDatabase
}

const ApiAppDBFileName string = "dvapi.db"

func (app *ApiApplication) Setup(host string, port int) (err error) {
	app.db = dvapi_db.NewDatabase()

	// Get PWDfor the database file as well
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "./"
	}

	err = app.db.Setup(fmt.Sprintf("%s/%s", pwd, ApiAppDBFileName))
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
