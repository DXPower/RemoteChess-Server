package utility

import (
	"context"
	"net/http"
	"strconv"

	. "remotechess/src/rc_server/api"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CtxIntFromURL(chiUrlVarName string, displayName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var dataStr string = chi.URLParam(r, chiUrlVarName)
			data, err := strconv.Atoi(dataStr)

			if err != nil {
				http.Error(w, "Invalid "+displayName, 400)
				return
			}

			ctx := context.WithValue(r.Context(), chiUrlVarName, data)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CtxFetchFromUrl(chiUrlVarName string, displayName string, ctxFetchedName string, fetcher func(uint64) (interface{}, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var dataStr string = chi.URLParam(r, chiUrlVarName)
			data, err := strconv.Atoi(dataStr)

			if err != nil {
				render.Render(w, r, NewErrResponse("Invalid "+displayName, 400, false))
				return
			}

			fetched, err := fetcher(uint64(data))

			if err != nil {
				render.Render(w, r, NewErrResponseFromServiceErr(err, HTTP_STATUS_DEFAULT, ERROR_DEFAULT_OBSCURED))
				return
			}

			ctx := context.WithValue(r.Context(), ctxFetchedName, fetched)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// 	var dataStr string = chi.URLParam(r, chiVarName)
// 	data, err := strconv.Atoi(dataStr)

// 	if err != nil {
// 		return 0, errors.New("Invalid " + displayName)
// 	}

// 	return data, nil
// }
