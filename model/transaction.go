package model

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"

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

func (trans *Trans) Exec(db *sqlx.DB) (*Acc, *Acc, error) {
	tx := db.MustBegin()
	var fromAcc, toAcc *Acc
	var err error
	if trans.FromID < trans.ToID {
		fromAcc, err = GetAccForUpdate(trans.FromID, tx)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "fromAcc %d GetAccForUpdate: %s \n", trans.FromID, err)
			return nil, nil, errors.New("source account doesn't exist")
		}

		if fromAcc.Balance < trans.Amount {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "fromAcc.Balance < trans.Amount: %s\n", err)
			return fromAcc, nil, errors.New("source account doesn't have enough balance")
		}

		toAcc, err = GetAccForUpdate(trans.ToID, tx)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "toAcc %d GetAccForUpdate: %s\n", trans.ToID, err)

			return fromAcc, nil, errors.New("dest account doesn't exist")
		}
	} else {
		toAcc, err = GetAccForUpdate(trans.ToID, tx)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "toAcc %d GetAccForUpdate: %s\n", trans.ToID, err)
			return nil, toAcc, errors.New("dest account doesn't exist")
		}

		fromAcc, err = GetAccForUpdate(trans.FromID, tx)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "fromAcc %d GetAccForUpdate: %s\n", trans.FromID, err)
			return nil, toAcc, errors.New("source account doesn't exist")
		}

		if fromAcc.Balance < trans.Amount {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "fromAcc.Balance < trans.Amount: %s\n", err)
			return fromAcc, toAcc, errors.New("source account doesn't have enough balance")
		}
	}

	fromAccBalance := math.Round((fromAcc.Balance-trans.Amount)*100) / 100
	err = AccBalanceUpdate(fromAcc, fromAccBalance, tx)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "AccBalanceUpdate(fromAcc, fromAccBalance, tx): %s\n", err)
		return fromAcc, toAcc, errors.New("dest account doesn't exist")
	}

	toAccBalance := math.Round((toAcc.Balance+trans.Amount)*100) / 100
	err = AccBalanceUpdate(toAcc, toAccBalance, tx)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "AccBalanceUpdate(toAcc, toAccBalance, tx): %s\n", err)
		return fromAcc, toAcc, errors.New("dest account doesn't exist")
	}

	err = tx.Commit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tx.Commit(): %s\n", err)
		fmt.Println("TRANSACTION ERR: ", err)
	}

	return fromAcc, toAcc, err
}
