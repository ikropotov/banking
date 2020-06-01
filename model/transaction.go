package model

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"

	"github.com/go-chi/render"

	"github.com/jmoiron/sqlx"
)

type Trans struct {
	FromID int     `json:"fromId"`
	ToID   int     `json:"toId"`
	Amount float64 `json:"amount"`
}

func (trans *Trans) Bind(r *http.Request) error {
	trans.Amount = math.Round(trans.Amount*100) / 100
	if trans.Amount <= 0 {
		return errors.New("amount should be positive")
	}
	if trans.FromID == trans.ToID {
		return errors.New("cannot transfer to same account")
	}
	return nil
}

func (trans *Trans) Exec(db *sqlx.DB) *TransResponse {
	tx := db.MustBegin()
	var fromAcc, toAcc *Acc
	var err error
	var errCode int
	if trans.FromID < trans.ToID {
		fromAcc, err = GetAccForUpdate(trans.FromID, tx)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "fromAcc %d GetAccForUpdate: %s \n", trans.FromID, err)
			if NotExists(err) {
				errCode = http.StatusNotFound

			} else {
				errCode = http.StatusInternalServerError
			}
			return &TransResponse{
				From:    fromAcc,
				To:      toAcc,
				ErrCode: errCode,
			}
		}

		toAcc, err = GetAccForUpdate(trans.ToID, tx)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "toAcc %d GetAccForUpdate: %s\n", trans.ToID, err)
			if NotExists(err) {
				errCode = http.StatusNotFound
			} else {
				errCode = http.StatusInternalServerError
			}
			return &TransResponse{
				From:    fromAcc,
				To:      toAcc,
				ErrCode: errCode,
			}
		}

		if fromAcc.Balance < trans.Amount {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "fromAcc.Balance < trans.Amount: %s\n", err)
			return &TransResponse{
				From:    fromAcc,
				To:      toAcc,
				ErrCode: http.StatusNotAcceptable,
			}
		}

	} else {
		toAcc, err = GetAccForUpdate(trans.ToID, tx)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "toAcc %d GetAccForUpdate: %s\n", trans.ToID, err)
			if NotExists(err) {
				errCode = http.StatusNotFound
			} else {
				errCode = http.StatusInternalServerError
			}
			return &TransResponse{
				From:    fromAcc,
				To:      toAcc,
				ErrCode: errCode,
			}
		}

		fromAcc, err = GetAccForUpdate(trans.FromID, tx)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "fromAcc %d GetAccForUpdate: %s\n", trans.FromID, err)
			if NotExists(err) {
				errCode = http.StatusNotFound
			} else {
				errCode = http.StatusInternalServerError
			}
			return &TransResponse{
				From:    fromAcc,
				To:      toAcc,
				ErrCode: errCode,
			}
		}

		if fromAcc.Balance < trans.Amount {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "fromAcc.Balance < trans.Amount: %s\n", err)
			return &TransResponse{
				From:    fromAcc,
				To:      toAcc,
				ErrCode: http.StatusNotAcceptable,
			}
		}
	}

	fromAccBalance := math.Round((fromAcc.Balance-trans.Amount)*100) / 100
	err = AccBalanceUpdate(fromAcc, fromAccBalance, tx)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "AccBalanceUpdate(fromAcc, fromAccBalance, tx): %s\n", err)
		return &TransResponse{
			From:    fromAcc,
			To:      toAcc,
			ErrCode: http.StatusInternalServerError,
		}
	}

	toAccBalance := math.Round((toAcc.Balance+trans.Amount)*100) / 100
	err = AccBalanceUpdate(toAcc, toAccBalance, tx)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "AccBalanceUpdate(toAcc, toAccBalance, tx): %s\n", err)
		return &TransResponse{
			From:    fromAcc,
			To:      toAcc,
			ErrCode: http.StatusInternalServerError,
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tx.Commit(): %s\n", err)
		return &TransResponse{
			From:    fromAcc,
			To:      toAcc,
			ErrCode: http.StatusInternalServerError,
		}
	}

	return &TransResponse{
		From:    fromAcc,
		To:      toAcc,
		ErrCode: http.StatusAccepted,
	}
}

type TransResponse struct {
	From    *Acc `json:"from"`
	To      *Acc `json:"to"`
	ErrCode int  `json:"-"`
}

func (e *TransResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if e.ErrCode == 0 {
		render.Status(r, http.StatusAccepted)
	} else {
		render.Status(r, e.ErrCode)
	}
	return nil
}
