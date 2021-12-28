package rc_server

import (
	"fmt"
	"net/http"
	. "remotechess/src/rc_server/api"
	. "remotechess/src/rc_server/api/chessboards"
	. "remotechess/src/rc_server/api/games"
	. "remotechess/src/rc_server/api/invitations"
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
		} else if errResp, ok := v.(*ErrResponse); ok {
			render.PlainText(w, r, errResp.Detail)
		} else if stringer, ok := v.(fmt.Stringer); ok {
			render.PlainText(w, r, stringer.String())
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
	gh := NewGameHandler(server)
	ih := NewInvitationHandler(server)

	server.Router.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		r.Route("/usercore/{userId}", uch.Router())
		r.Route("/chessboard/{boardId}", cbh.Router)
		r.Route("/game", gh.Router)
		r.Route("/invites", ih.Router())
	})
}
