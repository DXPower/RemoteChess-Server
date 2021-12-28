package games

import (
	"net/http"
	. "remotechess/src/rc_server/api"
	"remotechess/src/rc_server/api/utility"
	. "remotechess/src/rc_server/servercore"
	. "remotechess/src/rc_server/service/chessboards"
	. "remotechess/src/rc_server/service/games"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/notnil/chess"
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

		game.Get("/gamestate", gh.GameState)
		game.Get("/legalmoves", gh.LegalMoves)

		game.Group(func(board chi.Router) {
			// This is temporary only for debugging purposes to easily read the board output
			board.Use(render.SetContentType(render.ContentTypePlainText))

			board.Use(utility.CtxFetchFromUrl("boardId", "Board ID", "board", func(x uint64) (interface{}, error) {
				return FetchChessboard(x)
			}))

			board.Group(func(g chi.Router) {
				g.Use(utility.CtxStringFromURL("move", "Move UCI", false))
				g.Get("/move/{boardId}/{move}", gh.Move)
			})

			board.Group(func(g chi.Router) {
				g.Use(utility.CtxStringFromURL("drawMethod", "Draw Method", false))
				g.Get("/draw/{boardId}/offer/{drawMethod}", gh.OfferDraw)
			})

			board.Get("/resign/{boardId}", gh.Resign)
			board.Get("/draw/{boardId}/accept", gh.ResolveDraw(true))
			board.Get("/draw/{boardId}/reject", gh.ResolveDraw(false))
		})

		game.Group(func(g chi.Router) {
			g.Use(render.SetContentType(render.ContentTypePlainText))
			g.Get("/undo", gh.Undo)
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

	_, err := CreateChessGame(white, black)

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

	if game.GetOutcome() == NO_OUTCOME {
		render.Render(w, r, NewGameStateResponse(*game))
	} else {
		render.Render(w, r, NewWonGameStateResponse(*game))
	}
}

func (gh *GameHandler) Undo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	game, ok := ctx.Value("game").(*ChessGame)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := game.UndoMove()

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	err = game.Save()

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewGameStateResponse(*game))
}

func (gh *GameHandler) Resign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	game, ok1 := ctx.Value("game").(*ChessGame)
	chessboard, ok2 := ctx.Value("board").(*Chessboard)

	if !ok1 || !ok2 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := game.ResignGame(*chessboard)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	err = game.Save()

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewGameStateResponse(*game))
}

func (gh *GameHandler) OfferDraw(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	game, ok1 := ctx.Value("game").(*ChessGame)
	chessboard, ok2 := ctx.Value("board").(*Chessboard)
	drawMethodStr, ok3 := ctx.Value("drawMethod").(string)

	if !ok1 || !ok2 || !ok3 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	var drawMethod GameMethod

	switch strings.ToLower(drawMethodStr) {
	case "offer":
		drawMethod = DRAW_OFFER
	case "threefold_repetition":
		drawMethod = THREEFOLD_REPETITION
	case "50_moves":
		drawMethod = FIFTY_MOVE_RULE
	default:
		render.Render(w, r, NewErrResponse("Invalid draw method. Must be offer, threefold_repetition, or 50_moves.", 400, false))
		return
	}

	err := game.OfferDraw(*chessboard, drawMethod)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}

func (gh *GameHandler) ResolveDraw(accept bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		game, ok1 := ctx.Value("game").(*ChessGame)
		chessboard, ok2 := ctx.Value("board").(*Chessboard)

		if !ok1 || !ok2 {
			render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
			return
		}

		var err error

		if accept {
			err = game.AcceptDraw(*chessboard)

			if err != nil {
				render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
				return
			}

			err = game.Save()
		} else {
			err = game.RejectDraw(*chessboard)
		}

		if err != nil {
			render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
			return
		}

		render.Render(w, r, NewSuccessResponse())
	}
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

func (gh *GameHandler) GameState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	game, ok := ctx.Value("game").(*ChessGame)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	if game.Game.Outcome() == chess.NoOutcome {
		render.Render(w, r, NewGameStateResponse(*game))
	} else {
		render.Render(w, r, NewWonGameStateResponse(*game))
	}
}

func (gh *GameHandler) LegalMoves(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	game, ok := ctx.Value("game").(*ChessGame)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	render.Render(w, r, NewLegalMovesResponse(*game))
}
