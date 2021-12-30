package games

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/notnil/chess"

	. "remotechess/src/rc_server/rcdb/chessboards"
	. "remotechess/src/rc_server/rcdb/games"
	sv "remotechess/src/rc_server/service"
	. "remotechess/src/rc_server/service/chessboards"
	. "remotechess/src/rc_server/service/common"
)

type gameOptions struct {
	FetchMoves    bool
	ProvidedFen   string
	ProvidedMoves []*chess.Move
}

type ChessGame struct {
	Id             uint64
	Game           *chess.Game
	White, Black   Chessboard
	OfferedDraw    GameMethod
	OfferingPlayer PlayerColor
}

type ChessGamePersistent struct {
	Id               uint64
	FkWhite, FkBlack uint64
	Fen              string
	CurrentMove      PlayerColor
	Outcome          GameOutcome
	Method           GameMethod
	OfferedDraw      GameMethod
	OfferingPlayer   PlayerColor
}

func MakeGameOptionsDefault() gameOptions {
	return gameOptions{FetchMoves: false, ProvidedFen: "", ProvidedMoves: nil}
}

func MakeGameOptionsFetchMoves() gameOptions {
	return gameOptions{FetchMoves: true, ProvidedFen: "", ProvidedMoves: nil}
}

func MakeGameOptionsProvidedMoves(moves []*chess.Move) gameOptions {
	return gameOptions{FetchMoves: true, ProvidedFen: "", ProvidedMoves: moves}
}

func newChessGame(id uint64, white Chessboard, black Chessboard, outcome GameOutcome, method GameMethod, offeredDraw GameMethod, offeringPlayer PlayerColor, options gameOptions) *ChessGame {
	var cg ChessGame

	cg.Id = id
	cg.White = white
	cg.Black = black
	cg.OfferedDraw = offeredDraw
	cg.OfferingPlayer = offeringPlayer

	if options.FetchMoves {
		cg.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
		moves, _ := cg.FetchMoves()

		for _, mStr := range moves {
			cg.Game.MoveStr(mStr)
		}
	} else if options.ProvidedFen != "" {
		fen, _ := chess.FEN(options.ProvidedFen)
		cg.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}), fen)
	} else if options.ProvidedMoves != nil {
		cg.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))

		for _, m := range options.ProvidedMoves {
			cg.Game.Move(m)
		}
	} else {
		cg.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	}

	if method == RESIGNATION {
		if outcome == WHITE_WON {
			cg.Game.Resign(chess.Black)
		} else if outcome == BLACK_WON {
			cg.Game.Resign(chess.White)
		}
	} else {
		if outcome == DRAW {
			cg.Game.Draw(chess.Method(method))
		}
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

func (cg *ChessGame) GetColorOfBoard(board Chessboard) (PlayerColor, error) {
	if board.OnboardId == cg.White.OnboardId {
		return PLAYER_WHITE, nil
	} else if board.OnboardId == cg.Black.OnboardId {
		return PLAYER_BLACK, nil
	} else {
		return PLAYER_WHITE, errors.New("Board not in this game")
	}
}

func (cg *ChessGame) GetBoardOfPlayer(player PlayerColor) *Chessboard {
	if player == PLAYER_WHITE {
		return &cg.White
	} else {
		return &cg.Black
	}
}

func (cg *ChessGame) GetOutcome() GameOutcome {
	return GameOutcome(cg.Game.Outcome())
}

func (cg *ChessGame) GetMethod() GameMethod {
	return GameMethod(cg.Game.Method())
}

func (cg *ChessGame) GetFEN() string {
	return cg.Game.FEN()
}

func (cg *ChessGame) GetMove(i int) *chess.Move {
	return cg.Game.GetMove(i)
}

func (cg *ChessGame) GetLegalMoves() []*chess.Move {
	return cg.Game.Position().ValidMoves()
}

func CreateChessGame(white *Chessboard, black *Chessboard) (*ChessGame, error) {
	cg := newChessGame(0, *white, *black, NO_OUTCOME, NO_METHOD, NO_METHOD, PLAYER_WHITE, MakeGameOptionsDefault())

	ctx := context.Background()
	tx, err := sv.Db.BeginTx(ctx, nil)

	if err != nil {
		return nil, sv.NewInternalError("CreateChessGame" + err.Error())
	}

	defer tx.Rollback()

	row := sv.Db.QueryRowContext(ctx, GetGameQuery(CREATE_GAME), white.OnboardId, black.OnboardId, cg.Game.FEN())

	if row.Err() != nil {
		return nil, sv.NewInternalError("CreateChessGame " + row.Err().Error())
	}

	err = row.Scan(&cg.Id)

	if err != nil {
		return nil, sv.NewInternalError("CreateChessGame " + err.Error())
	}

	res, err := sv.Db.ExecContext(ctx, GetChessboardQuery(UPDATE_CURRENT_GAME_MULTI), cg.Id, white.OnboardId, black.OnboardId)

	if err != nil {
		return nil, sv.NewInternalError("CreateChessGame " + err.Error())
	} else if rowsAffected, _ := res.RowsAffected(); rowsAffected != 2 {
		return nil, sv.NewInternalError("CreateChessGame fk_cur_game update did not affect 2 rows")
	}

	err = tx.Commit()

	if err != nil {
		return nil, sv.NewInternalError(err.Error())
	}

	white.CurGame.Int64 = int64(cg.Id)
	black.CurGame.Int64 = int64(cg.Id)

	return cg, nil
}

// Update any changes to the ChessGame to the database
func (cg *ChessGame) Save() error {
	res, err := sv.Db.Exec(GetGameQuery(UPDATE_GAME), cg.Id, cg.GetFEN(), cg.GetTurn(), cg.GetOutcome(), cg.GetMethod())

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

	err := row.Scan(&cgp.Id, &cgp.FkWhite, &cgp.FkBlack, &cgp.Fen, &cgp.CurrentMove, &cgp.Outcome, &cgp.Method, &cgp.OfferedDraw, &cgp.OfferingPlayer)

	if err == sql.ErrNoRows {
		return nil, sv.NewDoesNotExistError("Game")
	} else {
		white, _ := FetchChessboard(cgp.FkWhite)
		black, _ := FetchChessboard(cgp.FkBlack)

		cg := newChessGame(cgp.Id, *white, *black, cgp.Outcome, cgp.Method, cgp.OfferedDraw, cgp.OfferingPlayer, MakeGameOptionsFetchMoves())

		return cg, nil
	}
}

func FetchCurrentGame(cb *Chessboard) (*ChessGame, error) {
	if cb.OwnerId.Valid {
		return FetchChessGame(uint64(cb.CurGame.Int64))
	} else {
		return nil, nil
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

func (cg *ChessGame) UndoMove() error {
	if len(cg.Game.Moves()) == 0 {
		return sv.NewGenericError("No moves to undo", 405, sv.NOT_SENSITIVE)
	}

	res, err := sv.Db.Exec(GetGameQuery(DELETE_LAST_MOVE), cg.Id)

	if err != nil {
		return sv.NewInternalError("UndoMove " + err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewInternalError("UndoMove affected " + fmt.Sprint(rowsAffected) + " rows")
	}

	moves := cg.Game.Moves()
	*cg = *newChessGame(cg.Id, cg.White, cg.Black, NO_OUTCOME, NO_METHOD, NO_METHOD, PLAYER_WHITE, MakeGameOptionsProvidedMoves(moves[:len(moves)-1]))

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

func (cg *ChessGame) ResignGame(chessboard Chessboard) error {
	if chessboard.OnboardId == cg.White.OnboardId {
		cg.Game.Resign(chess.White)
	} else if chessboard.OnboardId == cg.Black.OnboardId {
		cg.Game.Resign(chess.Black)
	} else {
		return sv.NewGenericError("Chessboard is not a part of this game", 400, sv.NOT_SENSITIVE)
	}

	return nil
}

func (cg *ChessGame) OfferDraw(chessboard Chessboard, drawMethod GameMethod) error {
	eligbleDraws := cg.Game.EligibleDraws()

	for _, d := range eligbleDraws {
		if chess.Method(drawMethod) == d {
			goto DRAW_IS_ELIGIBLE
		}
	}

	return sv.NewGenericError("Draw method "+drawMethod.String()+" is not eligble", 409, sv.NOT_SENSITIVE)

DRAW_IS_ELIGIBLE:
	player, err := cg.GetColorOfBoard(chessboard)

	if err != nil {
		return sv.NewGenericError("Board is not in this game", 400, sv.NOT_SENSITIVE)
	}

	res, err := sv.Db.Exec(GetGameQuery(UPDATE_DRAW), cg.Id, drawMethod, player)

	if err != nil {
		return sv.NewInternalError(err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewInternalError("Draw update did not affect 1 row")
	}

	return nil
}

func (cg *ChessGame) AcceptDraw(chessboard Chessboard) error {
	if cg.OfferedDraw != DRAW_OFFER && cg.OfferedDraw != FIFTY_MOVE_RULE && cg.OfferedDraw != THREEFOLD_REPETITION {
		return sv.NewGenericError("There is no pending draw for this game", 409, sv.NOT_SENSITIVE)
	}

	if chessboard.OnboardId == cg.GetBoardOfPlayer(cg.OfferingPlayer).OnboardId {
		return sv.NewGenericError("You cannot accept this draw", 409, sv.NOT_SENSITIVE)
	}

	if chessboard.OnboardId != cg.White.OnboardId && chessboard.OnboardId != cg.Black.OnboardId {
		return sv.NewGenericError("You are not a player in this game", 403, sv.NOT_SENSITIVE)
	}

	err := cg.Game.Draw(chess.Method(cg.OfferedDraw))

	if err != nil {
		return sv.NewInternalError("Draw method is invalid when it should be")
	}

	return nil
}

func (cg *ChessGame) RejectDraw(chessboard Chessboard) error {
	if cg.OfferedDraw != DRAW_OFFER && cg.OfferedDraw != FIFTY_MOVE_RULE && cg.OfferedDraw != THREEFOLD_REPETITION {
		return sv.NewGenericError("There is no pending draw for this game", 409, sv.NOT_SENSITIVE)
	}

	if chessboard.OnboardId == cg.GetBoardOfPlayer(cg.OfferingPlayer).OnboardId {
		return sv.NewGenericError("You cannot accept this draw", 409, sv.NOT_SENSITIVE)
	}

	if chessboard.OnboardId != cg.White.OnboardId && chessboard.OnboardId != cg.Black.OnboardId {
		return sv.NewGenericError("You are not a player in this game", 403, sv.NOT_SENSITIVE)
	}

	res, err := sv.Db.Exec(GetGameQuery(UPDATE_DRAW), cg.Id, NO_METHOD, PLAYER_WHITE)

	if err != nil {
		return sv.NewInternalError(err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewInternalError("Draw update did not affect 1 row")
	}

	cg.OfferedDraw = NO_METHOD
	cg.OfferingPlayer = PLAYER_WHITE
	return nil
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
