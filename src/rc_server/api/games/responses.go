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
	IsEnPassant bool   `json:"enPassant"`
	Promotion   string `json:"promotion"`
	Castle      string `json:"castle"`
}

type GameStateResponse struct {
	GenericResponse
	boardPretty    string
	Id             uint64       `json:"id"`
	Pieces         string       `json:"pieces"`
	Turn           string       `json:"turn"`
	LastMove       MoveResponse `json:"lastMove"`
	InCheck        bool         `json:"check"`
	GameOver       bool         `json:"gameOver"`
	OfferedDraw    string       `json:"offeredDraw"`
	OfferingPlayer string       `json:"offeringPlayer"`
}

type WonGameStateResponse struct {
	GameStateResponse
	Outcome string `json:"outcome"`
	Method  string `json:"method"`
}

func NewGameStateResponse(cg ChessGame) *GameStateResponse {
	var gsr GameStateResponse

	gsr.GenericResponse = *NewSuccessResponse()
	gsr.Id = cg.Id
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
			IsEnPassant: lastMove.HasTag(chess.EnPassant),
			IsCapture:   lastMove.HasTag(chess.Capture),
			Promotion:   lastMove.Promo().String(),
			Castle:      castle,
		}

		gsr.InCheck = lastMove.HasTag(chess.Check)
	}

	gsr.GameOver = cg.GetOutcome() != NO_OUTCOME
	gsr.OfferedDraw = cg.OfferedDraw.String()
	gsr.OfferingPlayer = cg.OfferingPlayer.String()

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
	return fmt.Sprintf("%s%s Capture: %t En Passant: %t Promotion: %s Castle: %s", mr.Origin, mr.Destination, mr.IsCapture, mr.IsEnPassant, mr.Promotion, mr.Castle)
}

func (gsr *GameStateResponse) String() string {
	format := "%s\n\n" +
		"Pieces:\t\t\t%s\n" +
		"Turn:\t\t\t%s\n" +
		"Last Move:\t\t%s\n" +
		"In Check:\t\t%t\n" +
		"Game Over:\t\t%t\n" +
		"Offered Draw:\t%s\n" +
		"Offering Player:\t%s\n"

	return fmt.Sprintf(format, gsr.boardPretty, gsr.Pieces, gsr.Turn, gsr.LastMove.String(), gsr.InCheck, gsr.GameOver, gsr.OfferedDraw, gsr.OfferingPlayer)
}
