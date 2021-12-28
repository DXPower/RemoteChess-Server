package invitations

import (
	. "remotechess/src/rc_server/api"
)

type ResponseInvite struct {
	InviteId  int    `json:"inviteId"`
	SenderId  uint64 `json:"senderId"`
	Username  string `json:"senderUsername"`
	YourColor string `json:"yourColor"`
}

type GetPendingInvitesResponse struct {
	GenericResponse
	Invites []ResponseInvite
}

type InviteCodeResponse struct {
	GenericResponse
	InviteCode int `json:"inviteCode"`
}
