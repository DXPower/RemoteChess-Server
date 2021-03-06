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

func NewErrResponseFromServiceErr(err error, httpStatus int, obscureType ObscureError) *ErrResponse {
	var obscure bool = obscureType == ERROR_OBSCURED

	serviceError, ok := err.(*sv.ServiceError)

	if !ok {
		obscure = true
		println("UNSUPPORTED SERVICE ERROR " + err.Error())
		return NewErrResponse("UNSUPPORTED SERVICE ERROR "+err.Error(), 500, true)
	}

	if httpStatus == HTTP_STATUS_DEFAULT {
		httpStatus = serviceError.HttpCodeHint
	}

	if obscureType == ERROR_DEFAULT_OBSCURED {
		obscure = serviceError.IsSensitive()
	}

	return NewErrResponse(err.Error(), httpStatus, obscure)
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
