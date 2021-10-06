package main

import (
	"net/http"
	. "remotechess/src/frontend/users"
)

func (app *App) Routes() {
	uh := UserHandler{}

	app.Server.Router.HandleFunc("/hello", app.handleHello())
	app.Server.Router.Route("/user/{userId}", uh.UserRouter())
}

func (app *App) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
	}
}
