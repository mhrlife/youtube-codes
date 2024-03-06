package main

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"strconv"
)

func main() {
	// set up the database
	db, err := sqlx.Open("mysql", "root:mamad@/idempotent")
	if err != nil {
		panic(err)
	}

	// set up the web server
	e := echo.New()

	e.GET("/account/:account_id/purchases", func(c echo.Context) error {
		accountId, _ := strconv.Atoi(c.Param("account_id"))
		purchases, err := GetUserPurchases(db, accountId)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.JSON(200, purchases)
	})

	e.GET("/account/:account_id/purchase", func(c echo.Context) error {
		accountId, _ := strconv.Atoi(c.Param("account_id"))
		stockId, _ := strconv.Atoi(c.QueryParam("stock"))
		if stockId <= 0 {
			return c.String(401, "bad stock id")
		}

		if err := Purchase(db, accountId, stockId); err != nil {
			return c.String(500, err.Error())
		}

		return c.String(201, "purchase is done successfully")
	})

	e.GET("/account/:account_id/idempotentPurchase", func(c echo.Context) error {
		accountId, _ := strconv.Atoi(c.Param("account_id"))
		stockId, _ := strconv.Atoi(c.QueryParam("stock"))
		if stockId <= 0 {
			return c.String(401, "bad stock id")
		}

		idempotentKey := c.Request().Header.Get("X-Idempotent-Key")
		if idempotentKey == "" {
			return c.String(401, "bad idempotent key")
		}

		if err := PurchaseIdempotent(db, accountId, stockId, idempotentKey); err != nil {
			return c.String(500, err.Error())
		}

		return c.String(201, "purchase is done successfully")
	})

	e.Logger.Error(e.Start(":8080"))

}

type UserPurchase struct {
	ID           int `db:"purchase_id" json:"id"`
	AccountID    int `db:"account_id" json:"account_id"`
	StockID      int `db:"stock_id" json:"stock_id"`
	StockRemains int `db:"stock_remains" json:"stock_remains"`
	Price        int `db:"price" json:"price"`
}

func GetUserPurchases(db *sqlx.DB, accountId int) ([]UserPurchase, error) {
	userPurchases := make([]UserPurchase, 0)
	if err := db.Select(&userPurchases, `SELECT p.id AS purchase_id, p.account_id , p.stock_id, s.stock AS stock_remains, s.price
FROM purchase p
         JOIN account a ON p.account_id = a.id
         JOIN stocks s ON p.stock_id = s.id
WHERE p.account_id = ? ORDER BY p.id`, accountId); err != nil {
		return nil, err
	}
	return userPurchases, nil
}

func Tx(db *sqlx.DB, exec func(tx *sqlx.Tx) error) error {
	tx := db.MustBegin()
	err := exec(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInsufficientStock   = errors.New("insufficient stock")
)

type Stock struct {
	ID    int `db:"id"`
	Stock int `db:"stock"`
	Price int `db:"price"`
}

type Account struct {
	ID      int `db:"id"`
	Balance int `db:"balance"`
}

func Purchase(db *sqlx.DB, accountId int, stockId int) error {
	return Tx(db, func(tx *sqlx.Tx) error {
		var balance int
		if err := tx.Get(&balance,
			"SELECT  balance FROM account WHERE id=? LIMIT 1 FOR UPDATE ", accountId); err != nil {
			return err
		}

		var s Stock
		if err := tx.Get(&s,
			"SELECT id, stock, price FROM stocks WHERE id=? LIMIT 1 FOR UPDATE ", stockId); err != nil {
			return err
		}

		if s.Stock <= 0 {
			return ErrInsufficientStock
		}

		if s.Price > balance {
			return ErrInsufficientBalance
		}

		balance -= s.Price
		if _, err := tx.Exec("UPDATE account SET balance=? WHERE id=?", balance, accountId); err != nil {
			return err
		}

		if _, err := tx.Exec("INSERT INTO purchase (account_id, stock_id) VALUES (?,?)", accountId, stockId); err != nil {
			return err
		}

		if _, err := tx.Exec("UPDATE stocks SET stock=stock-1 WHERE id=?", stockId); err != nil {
			return err
		}

		return nil
	})
}

func PurchaseIdempotent(db *sqlx.DB, accountId int, stockId int, idempotentKey string) error {
	return Tx(db, func(tx *sqlx.Tx) error {

		var exists bool
		if err := tx.Get(&exists, `SELECT EXISTS(SELECT 1 FROM idempotency i
              INNER JOIN purchase p on i.purchase_id = p.id
              WHERE i.id = UUID_TO_BIN(?))`, idempotentKey); err != nil {
			return err
		}
		if exists {
			return nil
		}

		var balance int
		if err := tx.Get(&balance,
			"SELECT  balance FROM account WHERE id=? LIMIT 1 FOR UPDATE ", accountId); err != nil {
			return err
		}

		var s Stock
		if err := tx.Get(&s,
			"SELECT id, stock, price FROM stocks WHERE id=? LIMIT 1 FOR UPDATE ", stockId); err != nil {
			return err
		}

		if s.Stock <= 0 {
			return ErrInsufficientStock
		}

		if s.Price > balance {
			return ErrInsufficientBalance
		}

		balance -= s.Price
		if _, err := tx.Exec("UPDATE account SET balance=? WHERE id=?", balance, accountId); err != nil {
			return err
		}

		result, err := tx.Exec("INSERT INTO purchase (account_id, stock_id) VALUES (?,?)", accountId, stockId)
		if err != nil {
			return err
		}

		lastInsertedId, _ := result.LastInsertId()
		if _, err := tx.Exec(
			"INSERT INTO idempotency (id, purchase_id) VALUES (UUID_TO_BIN(?),?)",
			idempotentKey, lastInsertedId); err != nil {
			return err
		}

		if _, err := tx.Exec("UPDATE stocks SET stock=stock-1 WHERE id=?", stockId); err != nil {
			return err
		}

		return nil
	})
}
