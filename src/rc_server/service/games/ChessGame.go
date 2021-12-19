package games

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/notnil/chess"

	. "remotechess/src/rc_server/rcdb/games"
	sv "remotechess/src/rc_server/service"
	. "remotechess/src/rc_server/service/chessboards"
)

type PlayerColor int64

const (
	PLAYER_WHITE PlayerColor = iota
	PLAYER_BLACK
)

type gameOptions struct {
	FetchMoves    bool
	ProvidedFen   string
	ProvidedMoves string
}

type ChessGame struct {
	Id           uint64
	Game         *chess.Game
	White, Black Chessboard
}

type ChessGamePersistent struct {
	Id               uint64
	FkWhite, FkBlack uint64
	Fen              string
	CurrentMove      PlayerColor
	Completed        bool
}

func MakeGameOptionsDefault() gameOptions {
	return gameOptions{FetchMoves: false, ProvidedFen: "", ProvidedMoves: ""}
}

func MakeGameOptionsFetchMoves() gameOptions {
	return gameOptions{FetchMoves: true, ProvidedFen: "", ProvidedMoves: ""}
}

func newChessGame(id uint64, white Chessboard, black Chessboard, options gameOptions) *ChessGame {
	var cg ChessGame

	cg.Id = id
	cg.White = white
	cg.Black = black

	if options.FetchMoves {
		cg.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
		moves, _ := cg.FetchMoves()

		for _, mStr := range moves {
			cg.Game.MoveStr(mStr)
		}
	} else if options.ProvidedFen != "" {
		fen, _ := chess.FEN(options.ProvidedFen)
		cg.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}), fen)
	} else if options.ProvidedMoves != "" {
		panic("Not implemented")
		// cg.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))

		// var moves []chess.Move
		// json.Unmarshal(options.ProvidedMoves, &moves)

		// for _, m := range moves {
		// 	cg.Game.Move(&m)
		// }
	} else {
		cg.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	}

	return &cg
}

// Return the color of the the player whose turn it is to move next
func (cg *ChessGame) GetTurn() PlayerColor {
	if cg.Game.Position().Turn() == chess.White {
		return PLAYER_WHITE
	} else {
		return PLAYER_BLACK
	}
}

// Return the Chessboard whose turn it is to move next
func (cg *ChessGame) GetCurrentMover() *Chessboard {
	if cg.GetTurn() == PLAYER_WHITE {
		return &cg.White
	} else {
		return &cg.Black
	}
}

func CreateChessGame(white Chessboard, black Chessboard) (*ChessGame, error) {
	cg := newChessGame(0, white, black, MakeGameOptionsDefault())

	row := sv.Db.QueryRow(GetGameQuery(CREATE_GAME), white.OnboardId, black.OnboardId, cg.Game.FEN())

	if row.Err() != nil {
		return nil, sv.NewInternalError("CreateChessGame " + row.Err().Error())
	}

	err := row.Scan(&cg.Id)

	if err != nil {
		return nil, sv.NewInternalError("CreateChessGame " + err.Error())
	}

	return cg, nil
}

// Update any changes to the ChessGame to the database
func (cg *ChessGame) Save() error {
	movePtrs := cg.Game.Moves()
	moves := make([]chess.Move, len(movePtrs))

	for i, m := range movePtrs {
		moves[i] = *m
	}

	movesStr, err := json.Marshal(moves)

	if err != nil {
		return sv.NewInternalError("Could not encode moves as JSON")
	}

	res, err := sv.Db.Exec(GetGameQuery(UPDATE_GAME), cg.Id, cg.Game.FEN(), string(movesStr), cg.GetTurn())

	if err != nil {
		return sv.NewInternalError("NewChessGame " + err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewDoesNotExistError("ChessGame")
	}

	return nil
}

func FetchChessGame(id uint64) (*ChessGame, error) {
	var cgp ChessGamePersistent

	row := sv.Db.QueryRow(GetGameQuery(SELECT_GAME), id)

	if row.Err() != nil {
		return nil, sv.NewInternalError("FetchChessGame " + row.Err().Error())
	}

	err := row.Scan(&cgp.Id, &cgp.FkWhite, &cgp.FkBlack, &cgp.Fen, &cgp.CurrentMove, &cgp.Completed)

	if err == sql.ErrNoRows {
		return nil, sv.NewDoesNotExistError("Game")
	} else {
		white, _ := FetchChessboard(cgp.FkWhite)
		black, _ := FetchChessboard(cgp.FkBlack)

		cg := newChessGame(cgp.Id, *white, *black, MakeGameOptionsFetchMoves())

		return cg, nil
	}
}

func (cg *ChessGame) MakeMove(mover Chessboard, moveUci string) error {
	if cg.GetCurrentMover().OnboardId != mover.OnboardId {
		return sv.NewGenericError("Not your turn", 405, sv.NOT_SENSITIVE)
	}

	move, err := cg.Game.MoveStr(moveUci)

	if err != nil {
		if strings.Contains(err.Error(), "decode") {
			return sv.NewInvalidInputError("Move UCI")
		} else if strings.Contains(err.Error(), "invalid move") {
			return sv.NewGenericError("Move is not valid in current position", 405, sv.NOT_SENSITIVE)
		}
	}

	from := move.S1().String()
	to := move.S2().String()

	piece := move.PieceMoved()
	pieceType := CPieceToString(piece.Type())
	pieceColor := strings.ToUpper(piece.Color().Name())

	tags := move.GetTags()

	row := sv.Db.QueryRow(GetGameQuery(CREATE_MOVE), cg.Id, pieceColor, from, to, pieceType, tags)

	if row.Err() != nil {
		return sv.NewInternalError("MakeMove " + row.Err().Error())
	}

	var _id uint64
	err = row.Scan(&_id)

	if err != nil {
		return sv.NewInternalError("MakeMove " + err.Error())
	}
	return nil
}

func (cg *ChessGame) FetchMoves() ([]string, error) {
	moves := []string{}

	rows, err := sv.Db.Query(GetGameQuery(GET_MOVES), cg.Id)

	if err != nil {
		return nil, sv.NewInternalError(err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var from, to string
		err = rows.Scan(&from, &to)

		if err != nil {
			return nil, sv.NewInternalError(err.Error())
		}

		moves = append(moves, (from + to))
	}

	if err != nil {
		return nil, sv.NewInternalError(err.Error())
	}

	return moves, nil
}

func (cg *ChessGame) PrintBoard() string {
	pieceTiles := [13]byte{'-', 'K', 'Q', 'R', 'B', 'N', 'P', 'k', 'q', 'r', 'b', 'n', 'p'}

	output := make([]byte, 0, 64+7)
	squares := cg.Game.Position().Board().Flip(chess.UpDown).SquareMap()

	for i := 0; i < 64; i++ {
		piece := squares[chess.Square(i)] // If not in map, it will return 0-value of the enum, which is NoPiece (=0)
		output = append(output, pieceTiles[int8(piece)])

		if i != 63 {
			if (i+1)%8 == 0 {
				output = append(output, '\n')
			} else {
				output = append(output, ' ')
			}
		}
	}

	return string(output)
}

func (this *PlayerColor) Scan(value interface{}) error {
	b, ok := value.([]byte)

	if !ok {
		return sv.NewInternalError("Scan source is not []byte")
	}

	if bytes.Equal(b, []byte("WHITE")) {
		*this = PLAYER_WHITE
	} else if bytes.Equal(b, []byte("BLACK")) {
		*this = PLAYER_BLACK
	} else {
		return sv.NewInternalError("Invalid PlayerColor enum received: " + string(b))
	}

	return nil
}

func (this PlayerColor) Value() (driver.Value, error) {
	if this == PLAYER_WHITE {
		return "WHITE", nil
	} else if this == PLAYER_BLACK {
		return "BLACK", nil
	} else {
		return nil, sv.NewInternalError("Unknown PlayerColor")
	}
}

func CPieceToString(p chess.PieceType) string {
	switch p {
	case chess.King:
		return "KING"
	case chess.Queen:
		return "QUEEN"
	case chess.Rook:
		return "ROOK"
	case chess.Knight:
		return "KNIGHT"
	case chess.Bishop:
		return "BISHOP"
	case chess.Pawn:
		return "PAWN"
	default:
		panic("Unknown piece type")
	}
}

func StringToCPieceType(s string) chess.PieceType {
	switch s {
	case "KING":
		return chess.King
	case "QUEEN":
		return chess.Queen
	case "ROOK":
		return chess.Rook
	case "KNIGHT":
		return chess.Knight
	case "BISHOP":
		return chess.Bishop
	case "PAWN":
		return chess.Pawn
	default:
		panic("Unknown piece type")
	}
}
