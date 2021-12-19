package games

import (
	"net/http"
	. "remotechess/src/rc_server/api"
	"remotechess/src/rc_server/api/utility"
	. "remotechess/src/rc_server/servercore"
	. "remotechess/src/rc_server/service/chessboards"
	. "remotechess/src/rc_server/service/games"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type GameHandler struct {
	server *ServerCore
}

func NewGameHandler(s *ServerCore) GameHandler {
	return GameHandler{s}
}

func (gh *GameHandler) Router(router chi.Router) {
	router.Group(func(g chi.Router) {
		g.Use(utility.CtxFetchFromUrl("whiteBid", "White Board ID", "whiteBoard", func(x uint64) (interface{}, error) {
			return FetchChessboard(x)
		}))

		g.Use(utility.CtxFetchFromUrl("blackBid", "Black Board ID", "blackBoard", func(x uint64) (interface{}, error) {
			return FetchChessboard(x)
		}))

		g.Get("/create/w/{whiteBid}/b/{blackBid}", gh.CreateGame)
	})

	router.Route("/{gameId}", func(game chi.Router) {
		game.Use(utility.CtxFetchFromUrl("gameId", "Game ID", "game", func(x uint64) (interface{}, error) {
			return FetchChessGame(x)
		}))

		game.Group(func(move chi.Router) {
			// This is temporary only for debugging purposes to easily read the board output
			move.Use(render.SetContentType(render.ContentTypePlainText))

			move.Use(utility.CtxFetchFromUrl("boardId", "Board ID", "board", func(x uint64) (interface{}, error) {
				return FetchChessboard(x)
			}))

			move.Use(utility.CtxStringFromURL("move", "Move UCI", false))
			move.Get("/move/{boardId}/{move}", gh.Move)
		})

		game.Group(func(g chi.Router) {
			g.Use(render.SetContentType(render.ContentTypePlainText))
			g.Get("/print", gh.Print)
		})
	})
}

func (gh *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	white, ok1 := ctx.Value("whiteBoard").(*Chessboard)
	black, ok2 := ctx.Value("blackBoard").(*Chessboard)

	if !ok1 || !ok2 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	_, err := CreateChessGame(*white, *black)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}

func (gh *GameHandler) Move(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	game, ok1 := ctx.Value("game").(*ChessGame)
	board, ok3 := ctx.Value("board").(*Chessboard)
	move, ok2 := ctx.Value("move").(string)

	if !ok1 || !ok2 || !ok3 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := game.MakeMove(*board, move)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	err = game.Save()

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewPlainTextResponse(game.PrintBoard()))
}

func (gh *GameHandler) Print(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	game, ok := ctx.Value("game").(*ChessGame)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	boardStr := game.PrintBoard()
	render.Render(w, r, NewPlainTextResponse(boardStr))
}
