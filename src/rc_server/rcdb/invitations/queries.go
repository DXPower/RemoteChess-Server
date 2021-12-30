package invitations

type InvitationQuery int

const (
	SEND_INVITE InvitationQuery = iota
	CANCEL_SENT_INVITE
	GET_PENDING_INVITES
	GET_SENT_INVITE_SENDER_BOARD
	GET_CODE_INVITE_SENDER_BOARD
	ACCEPT_INVITE
	REJECT_INVITE
	CREATE_INVITE_WITH_CODE
	CANCEL_CODE_INVITE
	DELETE_INVITE
	CLEAR_INVITES
)

func GetInvitationQuery(q InvitationQuery) string {
	switch q {
	case CREATE_INVITE_WITH_CODE:
		return `SELECT "CreateInviteWithCode"($1, $2) as invite_code`
	case CANCEL_CODE_INVITE:
		return `DELETE FROM game_invites WHERE invite_code = $1`
	case SEND_INVITE:
		return `INSERT INTO game_invites (fk_sender, fk_recipient, recipient_color) VALUES ($1, $2, $3) RETURNING id`
	case CANCEL_SENT_INVITE:
		return `DELETE FROM game_invites WHERE fk_sender = $1 AND fk_recipient = $2`
	case GET_PENDING_INVITES:
		return `SELECT  
					game_invites.id,
					users.id,
					users.username,
					recipient_color
				FROM game_invites 
				LEFT JOIN chessboards 
					ON game_invites.fk_sender = chessboards.onboard_id 
				LEFT JOIN users 
					ON chessboards.fk_owner = users.id
				WHERE fk_recipient = $1 AND declined = 'false'`
	case GET_SENT_INVITE_SENDER_BOARD:
		return `SELECT
					sender.onboard_id     as sender_onboard_id,
					sender.fk_owner       as sender_fk_owner,
					sender.fk_cur_game    as sender_fk_cur_game
				FROM game_invites
				INNER JOIN chessboards sender ON sender.onboard_id = game_invites.fk_sender
				INNER JOIN chessboards recipientBoard on recipientBoard.onboard_id = $2
				INNER JOIN users recipientUser on recipientBoard.fk_owner = recipientUser.id
				WHERE 
						game_invites.id = $1
					AND game_invites.declined = 'false'
					AND (recipient_color IS NULL OR recipient_color = $3)`
	case GET_CODE_INVITE_SENDER_BOARD:
		return `SELECT
					sender.onboard_id     as sender_onboard_id,
					sender.fk_owner       as sender_fk_owner,
					sender.fk_cur_game    as sender_fk_cur_game,
					game_invites.recipient_color
				FROM game_invites
				INNER JOIN chessboards sender ON sender.onboard_id = game_invites.fk_sender
				WHERE invite_code = $1`
	case REJECT_INVITE:
		return `UPDATE game_invites SET declined = 'true' WHERE id = $1`
	case DELETE_INVITE:
		return `DELETE FROM game_invites WHERE id = $1`
	case CLEAR_INVITES:
		return `DELETE FROM game_invites WHERE fk_sender = $1`
	}

	panic("Invalid query select")
}
