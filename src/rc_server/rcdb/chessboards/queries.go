package chessboards

type ChessboardQuery int

const (
	SELECT_BOARD ChessboardQuery = iota
	REGISTER_BOARD
	ASSIGN_FIRST_OWNER
)

func GetChessboardQuery(q ChessboardQuery) string {
	switch q {
	case SELECT_BOARD:
		return `SELECT onboard_id, fk_owner FROM chessboards WHERE onboard_id = $1`
	case REGISTER_BOARD:
		return `INSERT INTO chessboards (onboard_id) VALUES ($1) RETURNING onboard_id, fk_owner`
	case ASSIGN_FIRST_OWNER:
		return `UPDATE chessboards
				SET fk_owner = $1
				WHERE onboard_id = $2
					AND fk_owner IS NULL`
	}

	panic("Invalid query select")
}
