package rc_server

import (
	"net/http"
	. "remotechess/src/rc_server/api"
	. "remotechess/src/rc_server/api/chessboards"
	. "remotechess/src/rc_server/api/usercore"
	"remotechess/src/rc_server/rcdb"
	. "remotechess/src/rc_server/servercore"
	sv "remotechess/src/rc_server/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func ContentResponder(w http.ResponseWriter, r *http.Request, v interface{}) {
	ct := render.GetRequestContentType(r)

	switch ct {
	case render.ContentTypePlainText:
		ptr, ok := v.(*PlainTextResponse)

		if ok {
			render.PlainText(w, r, string(*ptr))
		} else if str, ok := v.(*string); ok {
			render.PlainText(w, r, *str)
		}
	default:
		render.DefaultResponder(w, r, v)
	}
}

func InitServer() {
	sv.Db = rcdb.ConnectToDb()
	render.Respond = ContentResponder
}

func Routes(server *ServerCore) {
	uch := NewUserCoreHandler(server)
	cbh := NewChessboardHandler(server)

	server.Router.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		r.Route("/usercore/{userId}", uch.Router())
		r.Route("/chessboard/{boardId}", cbh.Router)
	})
}
