package chessboards

import (
	"fmt"
	"net/http"
	. "remotechess/src/rc_server/api"
	"remotechess/src/rc_server/api/utility"
	. "remotechess/src/rc_server/servercore"
	. "remotechess/src/rc_server/service/chessboards"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ChessboardHandler struct {
	server *ServerCore
}

func NewChessboardHandler(s *ServerCore) ChessboardHandler {
	return ChessboardHandler{s}
}

func (cbh *ChessboardHandler) Router(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(utility.CtxIntFromURL("boardId", "Board ID"))
		r.Get("/register", cbh.Register)
	})

	router.Group(func(r chi.Router) {
		r.Use(utility.CtxFetchFromUrl("boardId", "Board ID", "chessboard", func(x uint64) (interface{}, error) {
			return FetchChessboard(x)
		}))

		r.Group(func(r chi.Router) {
			r.Use(render.SetContentType(render.ContentTypePlainText))
			r.Get("/print", cbh.GetPretty)
		})

		r.Get("/leavegame", cbh.LeaveGame)
	})

	router.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))

		r.Use(utility.CtxFetchFromUrl("boardId", "Board ID", "chessboard", func(x uint64) (interface{}, error) {
			return FetchChessboard(x)
		}))

		r.Get("/print", cbh.GetPretty)
		r.Get("/leavegame", cbh.LeaveGame)
	})
}

func (cbh *ChessboardHandler) GetPretty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	chessboard, ok := ctx.Value("chessboard").(*Chessboard)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	fmtStr := "Onboard ID:\t\t%d\n" +
		"Owner ID:\t\t%s\n"

	var ownerStr string

	if chessboard.OwnerId.Valid {
		ownerStr = fmt.Sprint(chessboard.OwnerId.Int64)
	} else {
		ownerStr = "NO OWNER"
	}

	render.Render(w, r, NewPlainTextResponse(fmt.Sprintf(fmtStr, chessboard.OnboardId, ownerStr)))
}

func (cbh *ChessboardHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	boardId, ok := ctx.Value("boardId").(int)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	_, err := RegisterNewChessboard(uint64(boardId))

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}

func (cbh *ChessboardHandler) LeaveGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	board, ok := ctx.Value("chessboard").(*Chessboard)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := board.LeaveGame()

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}
