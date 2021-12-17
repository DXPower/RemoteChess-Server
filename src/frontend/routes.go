package main

import (
	"net/http"
	. "remotechess/src/frontend/appcore"
	//. "remotechess/src/frontend/users"
)

func Routes(app *AppCore) {
	// uh := NewUserHandler(app)

	app.Server.Router.HandleFunc("/hello", handleHello(app))
	// app.Server.Router.Route("/user/{userId}", uh.Router())
}

func handleHello(app *AppCore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!\n"))

		print("Hello world!\n")
		// var rows, err = app.Server.Database.Query("SELECT * FROM public.country")

		// if err != nil {
		// 	println("ERROR: ", err.Error())
		// 	return
		// }

		// defer rows.Close()

		// for rows.Next() {
		// 	var id int
		// 	var name string
		// 	var code string

		// 	rows.Scan(&id, &name, &code)

		// 	w.Write([]byte("Row: " + fmt.Sprint(id) + " " + name + " " + code + "\n"))
		// }

		// print("End rows!\n")
	}
}
