package api

import (
	"fmt"
	"net/http"
	sv "remotechess/src/rc_server/service"

	"github.com/go-chi/render"
)

const (
	HTTP_STATUS_DEFAULT = -1
)

type ObscureError int

const (
	ERROR_NOT_OBSCURED     = 0
	ERROR_OBSCURED         = 1
	ERROR_DEFAULT_OBSCURED = 2
)

type ErrResponse struct {
	Success    bool   `json:"success"`
	Detail     string `json:"error"`
	StatusCode int    `json:"status"`
	obscured   bool   `json:"-"`
}

func NewErrResponse(detail string, httpStatus int, obscured bool) *ErrResponse {
	return &ErrResponse{false, detail, httpStatus, obscured}
}

func NewErrResponseFromServiceErr(err error, httpStatus int, obscure ObscureError) *ErrResponse {
	var defaultStatus int
	var defaultObscure ObscureError

	switch err.(type) {
	case *sv.DoesNotExist:
		defaultStatus = 404
		defaultObscure = ERROR_NOT_OBSCURED
	case *sv.AlreadyExists:
		defaultStatus = 409
		defaultObscure = ERROR_NOT_OBSCURED
	case *sv.InternalError:
		defaultStatus = 500
		defaultObscure = ERROR_OBSCURED
	case *sv.GenericError:
		gerr, _ := err.(*sv.GenericError)

		defaultStatus = gerr.Code

		if gerr.IsSensitive() {
			defaultObscure = ERROR_OBSCURED
		} else {
			defaultObscure = ERROR_NOT_OBSCURED
		}
	default:
		obscure = ERROR_OBSCURED
		println("UNSUPPORTED SERVICE ERROR " + err.Error())
	}

	if httpStatus == HTTP_STATUS_DEFAULT {
		httpStatus = defaultStatus
	}

	if obscure == ERROR_DEFAULT_OBSCURED {
		obscure = defaultObscure
	}

	return NewErrResponse(err.Error(), httpStatus, obscure == ERROR_OBSCURED)
}

func (this *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, this.StatusCode)

	if this.obscured {
		println("ERROR - " + fmt.Sprint(this.StatusCode) + " " + this.Detail)
		this.StatusCode = 500
		this.Detail = "Internal server error"
	}

	return nil
}
