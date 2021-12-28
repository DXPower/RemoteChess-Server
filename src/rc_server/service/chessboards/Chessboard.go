package chessboards

import (
	// "database/sql"

	"database/sql"
	. "remotechess/src/rc_server/rcdb/chessboards"

	//. "remotechess/src/rc_server/rcdb/chessboards"
	sv "remotechess/src/rc_server/service"
	. "remotechess/src/rc_server/service/usercore"

	"github.com/lib/pq"
)

type Chessboard struct {
	OnboardId uint64
	OwnerId   sql.NullInt64
	CurGame   sql.NullInt64
}

func FetchChessboard(onboardId uint64) (*Chessboard, error) {
	var cb Chessboard

	row := sv.Db.QueryRow(GetChessboardQuery(SELECT_BOARD), onboardId)

	if row.Err() != nil {
		return &cb, sv.NewInternalError("FetchChessboard " + row.Err().Error())
	}

	err := row.Scan(&cb.OnboardId, &cb.OwnerId, &cb.CurGame)

	if err == sql.ErrNoRows {
		return &cb, sv.NewDoesNotExistError("Chessboard")
	} else {
		return &cb, nil
	}
}

func RegisterNewChessboard(onboardId uint64) (Chessboard, error) {
	var cb Chessboard

	row := sv.Db.QueryRow(GetChessboardQuery(REGISTER_BOARD), onboardId)

	if row.Err() != nil {
		pqErr, ok := row.Err().(*pq.Error)

		if ok {
			if pqErr.Code == "23505" {
				return cb, sv.NewAlreadyExistsError("Chessboard")
			} else {
				return cb, sv.NewInternalError("RegisterNewChessboard " + row.Err().Error())
			}
		} else {
			return cb, sv.NewInternalError("RegisterNewChessboard " + row.Err().Error())
		}
	}

	err := row.Scan(&cb.OnboardId, &cb.OwnerId)

	if err != nil {
		return cb, sv.NewInternalError("RegisterNewChessboard " + err.Error())
	} else {
		return cb, nil
	}
}

// Only works if the chessboard does not previously have an owner
func (cb *Chessboard) AssignFirstOwner(owner UserCore) error {
	res, err := sv.Db.Exec(GetChessboardQuery(ASSIGN_FIRST_OWNER), owner.Id, cb.OnboardId)

	if err != nil {
		return sv.NewInternalError("AssignFirstOwner " + err.Error())
	} else if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewGenericError("Chessboard already has owner", 405, sv.NOT_SENSITIVE)
	}

	return nil
}

func (cb *Chessboard) LeaveGame() error {
	res, err := sv.Db.Exec(GetChessboardQuery(UPDATE_CURRENT_GAME), cb.OnboardId, nil)

	if err != nil {
		return sv.NewInternalError("LeaveGame " + err.Error())
	} else if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewInternalError("LeaveGame did not affect 1 row")
	}

	return nil
}
