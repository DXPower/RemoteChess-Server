package usercore

import (
	"net/http"
)

type UserCoreResponse struct {
	Success bool `json:"success"`
}

func NewSuccessfulUserCoreResponse() UserCoreResponse {
	return UserCoreResponse{true}
}

func NewFailureUserCoreResponse() UserCoreResponse {
	return UserCoreResponse{false}
}

func (_ *UserCoreResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type GetUserCoreResponse struct {
	UserCoreResponse
	Id       uint64 `json:"id"`
	Username string `json:"username"`
}

type ResponseFriend struct {
	Id       uint64 `json:"id"`
	Username string `json:"username"`
}

type GetFriendsResponse struct {
	UserCoreResponse
	Friends []ResponseFriend `json:"friends"`
}
