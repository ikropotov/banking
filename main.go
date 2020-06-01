package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/ikropotov/banking/model"
	"github.com/ikropotov/banking/utils"
	"github.com/jmoiron/sqlx"
)

func seed(db *sqlx.DB) error {
	var err error

	if _, err := db.Exec("TRUNCATE accounts;"); err != nil {
		return err
	}

	for i := 1; i <= 1000; i++ {
		model.NewAcc(&model.Acc{i, float64(100)}, db)
	}
	return err
}

func main() {
	db, err := model.CreateDB()
	if err != nil {
		fmt.Printf("can't connect to db with err: %s", err)
		return
	}
	seed(db)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(model.AddDBContext(db))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Banking REST api."))
	})
	r.Route("/accounts", func(r chi.Router) {
		r.Post("/", CreateAcc)
		r.Get("/{accID}", GetAcc)
	})

	r.Route("/ops", func(r chi.Router) {
		r.Post("/transfer", Transfer)
	})

	http.ListenAndServe(":3333", r)
}

func GetAcc(w http.ResponseWriter, r *http.Request) {
	var acc *model.Acc
	if accID := chi.URLParam(r, "accID"); accID != "" {
		accIDint, err := strconv.Atoi(accID)
		if err == nil {
			acc, err = model.GetAcc(accIDint, model.GetDB(r))
		}
		if err != nil {
			render.Render(w, r, utils.ErrNotFound)
			return
		}
	} else {
		render.Render(w, r, utils.ErrNotFound)
		return
	}

	if err := render.Render(w, r, render.Renderer(acc)); err != nil {
		render.Render(w, r, utils.ErrRender(err))
		return
	}
}
func CreateAcc(w http.ResponseWriter, r *http.Request) {
	acc := &model.Acc{}
	if err := render.Bind(r, acc); err != nil {
		render.Render(w, r, utils.ErrInvalidRequest(err))
		return
	}

	if _, err := model.NewAcc(acc, model.GetDB(r)); err != nil {
		render.Render(w, r, utils.ErrInternalRequest(err))
		return
	}
	render.Status(r, http.StatusCreated)
	render.Render(w, r, render.Renderer(acc))
}

func Transfer(w http.ResponseWriter, r *http.Request) {
	trans := &model.Trans{}
	if err := render.Bind(r, trans); err != nil {
		render.Render(w, r, utils.ErrInvalidRequest(err))
		return
	}

	transResp := trans.Exec(model.GetDB(r))
	render.Render(w, r, render.Renderer(transResp))
}
