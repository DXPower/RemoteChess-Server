package invitations

import (
	// "context"
	// "database/sql"
	// "encoding/json"
	// "fmt"
	// "net/http"
	// "strconv"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	. "remotechess/src/rc_server/api"
	. "remotechess/src/rc_server/api/games"
	"remotechess/src/rc_server/api/utility"
	. "remotechess/src/rc_server/servercore"
	. "remotechess/src/rc_server/service/chessboards"
	service "remotechess/src/rc_server/service/common"
	. "remotechess/src/rc_server/service/invitations"
	. "remotechess/src/rc_server/service/usercore"
)

type InvitationHandler struct {
	server *ServerCore
}

func NewInvitationHandler(s *ServerCore) InvitationHandler {
	return InvitationHandler{s}
}

func (ih *InvitationHandler) Router() func(chi.Router) {
	return func(router chi.Router) {
		router.Group(func(g chi.Router) {
			g.Use(utility.CtxFetchFromUrl("userId", "User ID", "user", func(x uint64) (interface{}, error) {
				return FetchUserCore(x)
			}))

			g.Get("/pending/{userId}", ih.GetPendingInvites)
		})

		router.Group(func(g chi.Router) {
			g.Use(utility.CtxFetchFromUrl("boardId", "Board ID", "chessboard", func(x uint64) (interface{}, error) {
				return FetchChessboard(x)
			}))

			g.Get("/createcode/{boardId}", ih.CreateInvite)

			g.Group(func(g chi.Router) {
				g.Use(utility.CtxFetchFromUrl("userId", "User ID", "user", func(x uint64) (interface{}, error) {
					return FetchUserCore(x)
				}))

				g.Get("/send/f/{boardId}/t/{userId}", ih.SendInvite)
			})
		})

		router.Group(func(g chi.Router) {
			g.Use(utility.CtxIntFromURL("inviteCode", "Invite Code"))

			g.Get("/cancelcode/{inviteCode}", ih.CancelCodeInvite)
		})

		router.Group(func(g chi.Router) {
			g.Use(utility.CtxIntFromURL("inviteId", "Invite ID"))

			g.Get("/reject/{inviteId}", ih.RejectInvite)

			g.Group(func(g chi.Router) {
				g.Use(utility.CtxFetchFromUrl("recipientBid", "Recipient Board ID", "recipient", func(x uint64) (interface{}, error) {
					return FetchChessboard(x)
				}))

				g.Use(utility.CtxStringFromURL("recipientColor", "Recipient Color", false))

				g.Get("/accept/{inviteId}/r/{recipientBid}/{recipientColor}", ih.AcceptInvite)
			})
		})
	}
}

func (ih *InvitationHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	board, ok := ctx.Value("chessboard").(*Chessboard)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	inviteCode, err := CreateCodeInvite(*board)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, &InviteCodeResponse{GenericResponse{Success: true}, inviteCode})
}

func (ih *InvitationHandler) CancelCodeInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inviteCode, ok := ctx.Value("inviteCode").(int)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := CancelCodeInvite(inviteCode)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}

func (ih *InvitationHandler) SendInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	board, ok1 := ctx.Value("chessboard").(*Chessboard)
	user, ok2 := ctx.Value("user").(*UserCore)

	if !ok1 || !ok2 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := SendInvite(*board, *user)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}

func (ih *InvitationHandler) GetPendingInvites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := ctx.Value("user").(*UserCore)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	invites, err := GetPendingInvites(*user)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	invitesResponse := GetPendingInvitesResponse{GenericResponse: *NewSuccessResponse(), Invites: []ResponseInvite{}}

	for _, o := range invites {
		var invite ResponseInvite

		invite.InviteId = int(o.Id)
		invite.SenderId = o.Sender.Id
		invite.Username = o.Sender.Username

		if o.YourColor != nil {
			invite.YourColor = o.YourColor.String()
		} else {
			invite.YourColor = "YOUR_CHOICE"
		}

		invitesResponse.Invites = append(invitesResponse.Invites, invite)
	}

	render.Render(w, r, &invitesResponse)
}

func (ih *InvitationHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recipient, ok1 := ctx.Value("recipient").(*Chessboard)
	inviteId, ok2 := ctx.Value("inviteId").(int)
	recipientColorStr, ok3 := ctx.Value("recipientColor").(string)

	if !ok1 || !ok2 || !ok3 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	recipientColor, err := service.NewPlayerColor(recipientColorStr)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	game, err := AcceptInvite(recipient, uint64(inviteId), recipientColor)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewGameStateResponse(*game))
}

func (ih *InvitationHandler) RejectInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inviteId, ok := ctx.Value("inviteId").(int)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := RejectInvite(uint64(inviteId))

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}
