package utils

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/ikropotov/alter/model"
)

type TransResponse struct {
	From      *model.Acc `json:"from"`
	To        *model.Acc `json:"to"`
	ErrorText error      `json:"-"`
}

func (e *TransResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if e.ErrorText != nil {
		render.Status(r, http.StatusInternalServerError)
	} else {
		render.Status(r, http.StatusAccepted)
	}
	return nil
}
