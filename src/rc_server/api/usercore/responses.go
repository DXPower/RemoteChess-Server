package usercore

import (
	. "remotechess/src/rc_server/api"
)

type GetUserCoreResponse struct {
	GenericResponse
	Id       uint64 `json:"id"`
	Username string `json:"username"`
}

type ResponseFriend struct {
	Id       uint64 `json:"id"`
	Username string `json:"username"`
}

type GetFriendsResponse struct {
	GenericResponse
	Friends []ResponseFriend `json:"friends"`
}
