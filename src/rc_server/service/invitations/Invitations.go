package invitations

import (
	"database/sql"
	. "remotechess/src/rc_server/rcdb/invitations"
	sv "remotechess/src/rc_server/service"
	. "remotechess/src/rc_server/service/chessboards"
	. "remotechess/src/rc_server/service/common"
	. "remotechess/src/rc_server/service/games"
	. "remotechess/src/rc_server/service/usercore"

	"github.com/lib/pq"
)

func CreateCodeInvite(cb Chessboard) (int, error) {
	var inviteCode int

	row := sv.Db.QueryRow(GetInvitationQuery(CREATE_INVITE_WITH_CODE), cb.OnboardId, PLAYER_BLACK)

	if row.Err() != nil {
		return 0, sv.NewInternalError("CreateGameInviteWithCode " + row.Err().Error())
	}

	err := row.Scan(&inviteCode)

	if err != nil {
		return 0, sv.NewInternalError("CreateGameInviteWithCode " + err.Error())
	}

	return inviteCode, nil
}

func CancelCodeInvite(inviteCode int) error {
	res, err := sv.Db.Exec(GetInvitationQuery(CANCEL_CODE_INVITE), inviteCode)

	if err != nil {
		return sv.NewInternalError(err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewDoesNotExistError("Invite")
	}

	return nil
}

func SendInvite(sender Chessboard, recipient UserCore) error {
	res, err := sv.Db.Exec(GetInvitationQuery(SEND_INVITE), sender.OnboardId, recipient.Id, PLAYER_BLACK)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return sv.NewGenericError("There is already a pending invite for that user", 409, sv.NOT_SENSITIVE)
		} else {
			return sv.NewInternalError(err.Error())
		}
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewInternalError("SendInvite did not affect 1 row")
	}

	return nil
}

func GetPendingInvites(user UserCore) ([]PendingInvite, error) {
	rows, err := sv.Db.Query(GetInvitationQuery(GET_PENDING_INVITES), user.Id)

	if err != nil {
		return nil, sv.NewInternalError(err.Error())
	}

	invites := []PendingInvite{}

	defer rows.Close()

	for rows.Next() {
		var pending PendingInvite
		var recipientColor NullablePlayerColor

		err = rows.Scan(&pending.Id, &pending.Sender.Id, &pending.Sender.Username, &recipientColor)

		if err != nil {
			return nil, sv.NewInternalError(err.Error())
		}

		pending.YourColor = recipientColor.ToPointer()
		invites = append(invites, pending)
	}

	if err != nil {
		return nil, sv.NewInternalError(err.Error())
	}

	return invites, nil
}

func AcceptInvite(recipient *Chessboard, inviteId uint64, recipientColor PlayerColor) (*ChessGame, error) {
	var sender Chessboard

	row := sv.Db.QueryRow(GetInvitationQuery(GET_INVITE_SENDER_BOARD), inviteId, recipient.OnboardId, recipientColor)

	if row.Err() != nil {
		return nil, sv.NewInternalError("AcceptInvite " + row.Err().Error())
	}

	err := row.Scan(&sender.OnboardId, &sender.OwnerId, &sender.CurGame)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sv.NewDoesNotExistError("Invite")
		} else {
			return nil, sv.NewInternalError("AcceptInvite " + err.Error())
		}
	}

	if sender.CurGame.Valid || recipient.CurGame.Valid {
		return nil, sv.NewGenericError("Player(s) already in game", 409, sv.NOT_SENSITIVE)
	}

	var game *ChessGame

	if recipientColor == PLAYER_WHITE {
		game, err = CreateChessGame(recipient, &sender)
	} else {
		game, err = CreateChessGame(&sender, recipient)
	}

	if err != nil {
		return game, err
	}

	err = DeleteInvite(inviteId)

	return game, err
}

func RejectInvite(inviteId uint64) error {
	res, err := sv.Db.Exec(GetInvitationQuery(REJECT_INVITE), inviteId)

	if err != nil {
		return sv.NewInternalError("RejectInvite " + err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewDoesNotExistError("Invite")
	}

	return nil
}

func DeleteInvite(inviteId uint64) error {
	res, err := sv.Db.Exec(GetInvitationQuery(DELETE_INVITE), inviteId)

	if err != nil {
		return sv.NewInternalError("DeleteInvite " + err.Error())
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		return sv.NewDoesNotExistError("Invite")
	}

	return nil
}

type PendingInvite struct {
	Id        uint64
	Sender    UserCore
	YourColor *PlayerColor
}
