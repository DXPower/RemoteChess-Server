package main

import (
	"net/http"

	. "remotechess/src/frontend/appcore"
	"remotechess/src/rc_server"
	"remotechess/src/rc_server/servercore"

	_ "github.com/lib/pq"
)

func main() {
	var app AppCore

	app.Server = servercore.NewServerCore()

	rc_server.InitServer()

	rc_server.Routes(&app.Server)
	Routes(&app)

	println("Beginning to listen on localhost:3000")
	http.ListenAndServe(":3000", app.Server.Router)
}
