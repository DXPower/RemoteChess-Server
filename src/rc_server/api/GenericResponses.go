package api

import "net/http"

type GenericResponse struct {
	Success bool `json:"success"`
}

func NewSuccessResponse() *GenericResponse {
	return &GenericResponse{true}
}

func NewFailureResponse() *GenericResponse {
	return &GenericResponse{false}
}

func (this *GenericResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (this *GenericResponse) String() string {
	if this.Success {
		return "Success"
	} else {
		return "Error"
	}
}
