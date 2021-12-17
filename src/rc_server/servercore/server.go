package servercore

import (
	// "database/sql"
	// "remotechess/src/rc_server/db"

	"github.com/go-chi/chi/v5"
)

type ServerCore struct {
	//Database *sql.DB
	Router *chi.Mux
}

func NewServerCore() ServerCore {
	var s ServerCore

	s.Router = chi.NewRouter()
	//s.Database = db.ConnectToDb()

	return s
}
