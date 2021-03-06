package users

// import (
// 	"context"
// 	"net/http"
// 	"strconv"

// 	"github.com/go-chi/chi/v5"

// 	. "remotechess/src/frontend/appcore"
// )

// type UserHandler struct {
// 	app *AppCore
// }

// func NewUserHandler(app *AppCore) UserHandler {
// 	return UserHandler{app}
// }

// func (uh *UserHandler) Router() func(chi.Router) {
// 	return func(r chi.Router) {
// 		r.Use(uh.Ctx)
// 		r.Get("/", uh.GetUser)
// 	}
// }

// func (uh *UserHandler) Ctx(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var userIdStr string = chi.URLParam(r, "userId")

// 		userId, err := strconv.Atoi(userIdStr)

// 		if err != nil {
// 			w.Write([]byte("Invalid user ID!"))
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), "user", userId)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// func (uh *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()

// 	userId, ok := ctx.Value("user").(int)

// 	if !ok {
// 		http.Error(w, http.StatusText(422), 422)
// 		return
// 	}

// 	w.Write([]byte("User ID: " + strconv.Itoa(userId)))
// }
