package games

import (
	"fmt"
	. "remotechess/src/rc_server/api"
	. "remotechess/src/rc_server/service/games"
	"strings"

	"github.com/notnil/chess"
)

type MoveResponse struct {
	Origin      string `json:"from"`
	Destination string `json:"to"`
	IsCapture   bool   `json:"capture"`
	Promotion   string `json:"promotion"`
	Castle      string `json:"castle"`
}

type GameStateResponse struct {
	GenericResponse
	boardPretty   string
	Pieces        string       `json:"pieces"`
	Turn          string       `json:"turn"`
	LastMove      MoveResponse `json:"lastMove"`
	InCheck       bool         `json:"check"`
	GameOver      bool         `json:"gameOver"`
	IsDrawOffered bool         `json:"drawOffered"`
}

type WonGameStateResponse struct {
	GameStateResponse
	Outcome string `json:"outcome"`
	Method  string `json:"method"`
}

func NewGameStateResponse(cg ChessGame) *GameStateResponse {
	var gsr GameStateResponse

	gsr.GenericResponse = *NewSuccessResponse()
	gsr.boardPretty = cg.PrintBoard()
	gsr.Pieces = strings.Split(cg.GetFEN(), " ")[0]
	gsr.Turn = cg.GetTurn().String()

	lastMove := cg.GetMove(-1)

	if lastMove != nil {
		castle := ""

		if lastMove.HasTag(chess.KingSideCastle) {
			castle = "K"
		} else if lastMove.HasTag(chess.QueenSideCastle) {
			castle = "Q"
		}

		gsr.LastMove = MoveResponse{
			Origin:      lastMove.S1().String(),
			Destination: lastMove.S2().String(),
			IsCapture:   lastMove.HasTag(chess.Capture),
			Promotion:   lastMove.Promo().String(),
			Castle:      castle,
		}

		gsr.InCheck = lastMove.HasTag(chess.Check)
	}

	gsr.GameOver = cg.GetOutcome() != NO_OUTCOME
	gsr.IsDrawOffered = cg.IsDrawOffered

	return &gsr
}

func NewWonGameStateResponse(cg ChessGame) *WonGameStateResponse {
	var wgsr WonGameStateResponse

	wgsr.GameStateResponse = *NewGameStateResponse(cg)
	wgsr.Outcome = cg.GetOutcome().ToStore()
	wgsr.Method = cg.GetMethod().String()

	return &wgsr
}

func (mr *MoveResponse) String() string {
	return fmt.Sprintf("%s%s Capture: %t Promotion: %s Castle: %s", mr.Origin, mr.Destination, mr.IsCapture, mr.Promotion, mr.Castle)
}

func (gsr *GameStateResponse) String() string {
	format := "%s\n\n" +
		"Pieces:\t\t\t%s\n" +
		"Turn:\t\t\t%s\n" +
		"Last Move:\t\t%s\n" +
		"In Check:\t\t%t\n" +
		"Game Over:\t\t%t\n" +
		"Draw Offered:\t%t\n"

	return fmt.Sprintf(format, gsr.boardPretty, gsr.Pieces, gsr.Turn, gsr.LastMove.String(), gsr.InCheck, gsr.GameOver, gsr.IsDrawOffered)
}
