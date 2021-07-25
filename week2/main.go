package main

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
)

var (
	driverName string
	dsn string
)

func init() {
	driverName = ""
	dsn = ""
}

func queryNoRows() error {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}

	qry := "select id, name from NoRowsTable"
	rows, err := db.Query(qry)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id int64
			name string
		)

		if err := rows.Scan(&id, &name); err != nil {
			if err == sql.ErrNoRows {
				return errors.Wrap(err, "no rows returned")
			}

			return err
		}
	}

	return nil
}

func main() {
	if err := queryNoRows(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("original error. type: %T, value: %v\n", errors.Cause(err), errors.Cause(err))
			fmt.Printf("stack trace. %+v\n", err)
			return
		}

		fmt.Println("other errors happened")
	}

	fmt.Println("success")
}
