package chessboards

type GameQuery int

const (
	SELECT_GAME GameQuery = iota
	CREATE_GAME
	UPDATE_GAME
	CREATE_MOVE
	GET_MOVES
)

func GetGameQuery(q GameQuery) string {
	switch q {
	case SELECT_GAME:
		return `SELECT id, fk_white, fk_black, fen, current_move, completed FROM games WHERE id = $1`
	case CREATE_GAME:
		return `INSERT INTO games (fk_white, fk_black, fen) VALUES ($1, $2, $3) RETURNING id`
	case UPDATE_GAME:
		return `UPDATE games SET fen = $2, moves = $3, current_move = $4 WHERE id = $1`
	case CREATE_MOVE:
		return `INSERT INTO moves (fk_game, player, cell_from, cell_to, piece, tags) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	case GET_MOVES:
		return `SELECT cell_from, cell_to FROM moves WHERE fk_game = $1 ORDER BY move_num ASC`
	}

	panic("Invalid query select")
}
