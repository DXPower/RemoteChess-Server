package chessboards

type UserQuery int

const (
	SELECT_USER UserQuery = iota
	SEND_FRIEND_REQUEST
	GET_FRIENDS
	ACCEPT_FRIEND_REQUEST
	REMOVE_FRIEND
)

func GetUserCoreQuery(q UserQuery) string {
	switch q {
	case SELECT_USER:
		return `SELECT id, email, username FROM users WHERE id = $1`
	case SEND_FRIEND_REQUEST:
		return `INSERT INTO friends (fk_friend_left, fk_friend_right, pending) VALUES ($1, $2, true)`
	case GET_FRIENDS:
		return `SELECT 
					  users.id 
					, users.username
				FROM friends 
				INNER JOIN users ON (fk_friend_left = users.id OR fk_friend_right = users.id)
				WHERE 
					(($2 AND fk_friend_right = $1 AND pending = $2)
					OR (
						NOT $2 
						AND (
							   fk_friend_left = $1
							OR fk_friend_right = $1
						)
					)) AND users.id != $1
				ORDER BY users.username ASC`
	case ACCEPT_FRIEND_REQUEST:
		return `UPDATE friends 
				SET pending = false 
				WHERE 
					fk_friend_left  = $1 AND 
					fk_friend_right = $2 AND
					pending = true`
	case REMOVE_FRIEND:
		return `DELETE FROM friends 
				WHERE
					   (fk_friend_left = $1 AND fk_friend_right = $2)
					OR (fk_friend_left = $2 AND fk_friend_right = $1)`
	}

	panic("Invalid query select")
}
