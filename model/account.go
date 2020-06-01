package model

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Acc struct {
	ID      int     `json:"id"`
	Balance float64 `json:"balance"`
}

func (acc *Acc) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func (acc *Acc) Bind(r *http.Request) error {
	if acc.ID == 0 {
		return errors.New("cannot create account with ID:0")
	}
	existAcc, _ := GetAcc(acc.ID, GetDB(r))
	if existAcc != nil {
		return errors.New(fmt.Sprintf("account with ID:%d exists", acc.ID))
	}
	acc.Balance = math.Round(acc.Balance*100) / 100
	if acc.Balance < 0 {
		return errors.New("cannot assign negative balance")
	}
	return nil
}

func GetDB(r *http.Request) *sqlx.DB {
	return r.Context().Value("db").(*sqlx.DB)
}

func GetAcc(id int, db *sqlx.DB) (*Acc, error) {
	obj := &Acc{}
	sql := "SELECT * FROM accounts WHERE id=$1;"
	if err := db.Get(obj, sql, id); err != nil {
		if !NotExists(err) {
			fmt.Println("Select of user.id", id, ":", err)
		}
		return nil, err
	}

	return obj, nil
}

func GetAccForUpdate(id int, tx *sqlx.Tx) (*Acc, error) {
	obj := &Acc{}
	sql := "SELECT * FROM accounts WHERE id=$1 FOR UPDATE;"
	if err := tx.Get(obj, sql, id); err != nil {
		if !NotExists(err) {
			fmt.Println("Select of user.id", id, ":", err)
		}
		return nil, err
	}

	return obj, nil
}

func NotExists(err error) bool {
	if !strings.Contains(err.Error(), "no rows in result set") {
		return false
	}
	return true
}

func AccBalanceUpdate(acc *Acc, newBalance float64, tx *sqlx.Tx) error {
	sql := "UPDATE accounts SET balance = $1 WHERE id = $2;"
	if _, err := tx.Exec(sql, newBalance, acc.ID); err != nil {
		return err
	}
	acc.Balance = newBalance
	return nil
}

func NewAcc(acc *Acc, db *sqlx.DB) (*Acc, error) {
	_, err := db.NamedExec(`INSERT INTO accounts (id, balance) VALUES (:id,:balance)`, acc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db.NamedExec: %s for %+v\n", err, acc)
	}
	return acc, err
}
