package usercore

import (
	"database/sql"
	. "remotechess/src/rc_server/rcdb/usercore"
	sv "remotechess/src/rc_server/service"

	"github.com/lib/pq"
)

type UserCore struct {
	Id       uint64
	Email    string
	Username string
	Password string
}

func FetchUserCore(userId uint64) (*UserCore, error) {
	var user UserCore

	row := sv.Db.QueryRow(GetUserCoreQuery(SELECT_USER), userId)

	if row.Err() != nil {
		return nil, sv.NewInternalError("FetchUserCore " + row.Err().Error())
	}

	err := row.Scan(&user.Id, &user.Email, &user.Username)

	if err == sql.ErrNoRows {
		return nil, sv.NewDoesNotExistError("User")
	} else {
		return &user, nil
	}
}

func (user *UserCore) SendFriendRequest(friend UserCore) error {
	res, err := sv.Db.Exec(GetUserCoreQuery(SEND_FRIEND_REQUEST), user.Id, friend.Id)

	if err != nil {
		pqErr, ok := err.(*pq.Error)

		if ok {
			switch pqErr.Code {
			case "23503":
				return sv.NewDoesNotExistError("User")
			case "23505":
				return sv.NewAlreadyExistsError("Friendship")
			case "23514":
				return sv.NewGenericError("Cannot befriend yourself", 405, sv.NOT_SENSITIVE)
			}
		} else {
			return sv.NewInternalError("SendFriendRequest " + err.Error())
		}
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewInternalError("SendFriendRequest Could Not Be Completed")
	}

	return nil
}

func (user *UserCore) GetFriends(pending bool) ([]UserCore, error) {
	pendingRequests := []UserCore{}

	rows, err := sv.Db.Query(GetUserCoreQuery(GET_FRIENDS), user.Id, pending)

	if err != nil {
		return nil, sv.NewInternalError(err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var pending UserCore
		err = rows.Scan(&pending.Id, &pending.Username)

		if err != nil {
			return nil, sv.NewInternalError(err.Error())
		}

		pendingRequests = append(pendingRequests, pending)
	}

	if err != nil {
		return nil, sv.NewInternalError(err.Error())
	}

	return pendingRequests, nil
}

func (user *UserCore) AcceptFriendRequest(incoming UserCore) error {
	res, err := sv.Db.Exec(GetUserCoreQuery(ACCEPT_FRIEND_REQUEST), incoming.Id, user.Id)

	if err != nil {
		return sv.NewInternalError(err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewDoesNotExistError("Friend Request")
	}

	return nil
}

// Works for both existing friends and incoming friend requests
func (user *UserCore) RemoveFriend(friend UserCore) error {
	res, err := sv.Db.Exec(GetUserCoreQuery(REMOVE_FRIEND), friend.Id, user.Id)

	if err != nil {
		return sv.NewInternalError(err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewDoesNotExistError("Friend")
	}

	return nil
}
