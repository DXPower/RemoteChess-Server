package api

import (
	"net/http"
)

type PlainTextResponse string

func NewPlainTextResponse(pt string) *PlainTextResponse {
	var ptr PlainTextResponse = PlainTextResponse(pt)

	return &ptr
}

func (this *PlainTextResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
