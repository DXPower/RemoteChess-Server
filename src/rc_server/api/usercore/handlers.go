package usercore

import (
	// "context"
	// "database/sql"
	// "encoding/json"
	// "fmt"
	// "net/http"
	// "strconv"

	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	. "remotechess/src/rc_server/api"
	"remotechess/src/rc_server/api/utility"
	. "remotechess/src/rc_server/servercore"
	. "remotechess/src/rc_server/service/chessboards"
	. "remotechess/src/rc_server/service/usercore"
)

type UserCoreHandler struct {
	server *ServerCore
}

func NewUserCoreHandler(s *ServerCore) UserCoreHandler {
	return UserCoreHandler{s}
}

func (uch *UserCoreHandler) Router() func(chi.Router) {
	return func(router chi.Router) {
		router.Use(utility.CtxFetchFromUrl("userId", "User ID", "user", func(x uint64) (interface{}, error) {
			return FetchUserCore(x)
		}))

		router.Get("/", uch.Get)

		router.Group(func(g chi.Router) {
			g.Use(utility.CtxFetchFromUrl("boardId", "Board ID", "chessboard", func(x uint64) (interface{}, error) {
				return FetchChessboard(x)
			}))

			g.Get("/registerboard/{boardId}", uch.RegisterBoard)
		})

		router.Group(func(g chi.Router) {
			g.Use(render.SetContentType(render.ContentTypePlainText))
			g.Get("/print", uch.GetPretty)
		})

		router.Route("/friends", func(fr chi.Router) {
			fr.Get("/", uch.GetFriends(false))
			fr.Get("/pending", uch.GetFriends(true))

			fr.Route("/{friendId}", func(fr2 chi.Router) {
				fr2.Use(utility.CtxFetchFromUrl("friendId", "Friend ID", "friend", func(x uint64) (interface{}, error) {
					return FetchUserCore(x)
				}))

				fr2.Get("/send", uch.SendFriendRequest)
				fr2.Get("/accept", uch.AcceptFriendRequest)
				fr2.Get("/reject", uch.RemoveFriend)
				fr2.Get("/remove", uch.RemoveFriend)
			})
		})
	}
}

func (uh *UserCoreHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := ctx.Value("user").(*UserCore)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	render.Render(w, r, &GetUserCoreResponse{
		GenericResponse: *NewSuccessResponse(),
		Id:              user.Id,
		Username:        user.Username},
	)
}

func (uh *UserCoreHandler) GetPretty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := ctx.Value("user").(*UserCore)

	if !ok {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	fmtStr := "User ID:\t\t%d\n" +
		"User Email:\t\t%s\n" +
		"User Username:\t%s\n"

	render.Render(w, r, NewPlainTextResponse(fmt.Sprintf(fmtStr, user.Id, user.Email, user.Username)))
}

func (uch *UserCoreHandler) RegisterBoard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok1 := ctx.Value("user").(*UserCore)
	board, ok2 := ctx.Value("chessboard").(*Chessboard)

	if !ok1 || !ok2 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := board.AssignFirstOwner(*user)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}

func (uch *UserCoreHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok1 := ctx.Value("user").(*UserCore)
	friend, ok2 := ctx.Value("friend").(*UserCore)

	if !ok1 || !ok2 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := user.SendFriendRequest(*friend)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}

func (uch *UserCoreHandler) GetFriends(pending bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, ok := ctx.Value("user").(*UserCore)

		if !ok {
			render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
			return
		}

		pendingRequestsUserCores, err := user.GetFriends(pending)

		if err != nil {
			render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
			return
		}

		pendingRequests := make([]ResponseFriend, len(pendingRequestsUserCores))

		for i, r := range pendingRequestsUserCores {
			pendingRequests[i] = ResponseFriend{Id: r.Id, Username: r.Username}
		}

		render.Render(w, r, &GetFriendsResponse{*NewSuccessResponse(), pendingRequests})
	}
}

func (uch *UserCoreHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok1 := ctx.Value("user").(*UserCore)
	friend, ok2 := ctx.Value("friend").(*UserCore)

	if !ok1 || !ok2 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := user.AcceptFriendRequest(*friend)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}

func (uch *UserCoreHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok1 := ctx.Value("user").(*UserCore)
	friend, ok2 := ctx.Value("friend").(*UserCore)

	if !ok1 || !ok2 {
		render.Render(w, r, NewErrResponse(http.StatusText(422), 422, true))
		return
	}

	err := user.RemoveFriend(*friend)

	if err != nil {
		render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
		return
	}

	render.Render(w, r, NewSuccessResponse())
}
