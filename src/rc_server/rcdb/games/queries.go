package chessboards

type GameQuery int

const (
	SELECT_GAME GameQuery = iota
	CREATE_GAME
	UPDATE_GAME
	CREATE_MOVE
	GET_MOVES
	GET_LAST_MOVE
	DELETE_LAST_MOVE
	UPDATE_DRAW
)

func GetGameQuery(q GameQuery) string {
	switch q {
	case SELECT_GAME:
		return `SELECT id, fk_white, fk_black, fen, current_move, outcome, method, offered_draw, offering_player FROM games WHERE id = $1`
	case CREATE_GAME:
		return `INSERT INTO games (fk_white, fk_black, fen) VALUES ($1, $2, $3) RETURNING id`
	case UPDATE_GAME:
		return `UPDATE games SET fen = $2, current_move = $3, outcome = $4, method = $5 WHERE id = $1`
	case CREATE_MOVE:
		return `INSERT INTO moves (fk_game, player, cell_from, cell_to, piece, tags) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	case GET_MOVES:
		return `SELECT cell_from, cell_to FROM moves WHERE fk_game = $1 ORDER BY move_num ASC`
	case GET_LAST_MOVE:
		return `SELECT DISTINCT ON(fk_game) fk_game, move_num, player, cell_from, cell_to, piece, tags FROM moves ORDER BY fk_game, move_num DESC`
	case DELETE_LAST_MOVE:
		return `DELETE FROM moves
				WHERE
					fk_game = $1 AND
					move_num = (SELECT MAX(move_num) FROM moves WHERE fk_game = $1)`
	case UPDATE_DRAW:
		return `UPDATE games SET offered_draw = $2, offering_player = $3 WHERE id = $1`
	}

	panic("Invalid query select")
}
