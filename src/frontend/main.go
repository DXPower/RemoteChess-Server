package main

import (
	"net/http"

	. "remotechess/src/rc_server"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	Server Server
}

func main() {
	var app App

	app.Server.Router = chi.NewRouter()

	app.Server.Router.Use(middleware.Logger)
	app.Server.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	app.Routes()

	http.ListenAndServe(":3000", app.Server.Router)

}
